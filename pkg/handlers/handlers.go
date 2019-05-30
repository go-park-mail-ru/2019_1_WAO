package handlers

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"time"
	"context"
	"io/ioutil"
	"github.com/DmitriyPrischep/backend-WAO/pkg/model"
	"github.com/DmitriyPrischep/backend-WAO/pkg/aws"
	"github.com/DmitriyPrischep/backend-WAO/pkg/db"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/methods"
)

func NewUserHandler(database *driver.DB, client auth.AuthCheckerClient, setting *aws.ConnectSetting, path string) *Handler {
	return &Handler{
		hand: db.NewDataBase(database.DB),
		auth: client,
		aws: setting,
		imagePath: path,
	}
}

type Handler struct {
	hand methods.UserMethods
	auth auth.AuthCheckerClient
	aws *aws.ConnectSetting
	imagePath string
}

func (h *Handler)GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	users, err := h.hand.GetUsers()
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := model.SendUsers{
		Users: users,
	}.MarshalJSON()
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
	}
	var user model.UserRegister
	if err := user.UnmarshalJSON(body); err != nil {
		log.Printf("Decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Email == "" || user.Nickname == "" || user.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("InputPASS: ", user.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Generate hash password error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	log.Println("DEBUG: ", user)

	nickname, err := h.hand.CreateUser(user)
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
	}
	log.Println("New record NICK is:", nickname)

	sess, err := h.auth.Create(
		context.Background(),
		&auth.UserData{
			Login: nickname,
			Agent: r.UserAgent(),
		})
	if err != nil {
		log.Println("Authentification server is not available: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.Value,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	})
}

func (h *Handler) GetUsersByNick(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	
	params := mux.Vars(r)

	user, err := h.hand.GetUser(model.NicknameUser{Nickname: params["login"]})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user.Image != "" {
		user.Image = h.imagePath + user.Image
	}

	w.Header().Set("Content-Type", "application/json")
	out, err := user.MarshalJSON()
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (h *Handler)ModifiedUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	params := mux.Vars(r)
	userLogin := params["login"]
	fmt.Println("userLogin: ", userLogin)
	if userLogin == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	newData := model.UpdateDataImport{
		Email: r.FormValue("email"),
		Password: r.FormValue("password"),
		Nickname: r.FormValue("nickname"),
		OldNick: userLogin,
	}

	if newData.OldNick != newData.Nickname {
		log.Println("old:", newData.OldNick, " new:", newData.Nickname)
		log.Println("NEW DATA:", newData)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(newData.Password) > 5 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newData.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Generate hash password error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		newData.Password = string(hashedPassword)
	}

	var url string
	if _, _, err := r.FormFile("image"); err != nil {
		log.Println("Field with this name not exist")
	} else {
		err = r.ParseMultipartForm(5 * 1024 * 1024)
		if err != nil {
			log.Println("Multipart form parse errror", err.Error())
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		file, handler, err := r.FormFile("image")
		if err != nil {
			log.Println("Error read file from 'image' field:", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		
		conn := aws.NewConnectAWS(h.aws)
		url, err = conn.UploadImage(file, handler.Filename, handler.Size)
		if err != nil {
			log.Printf("Upload image error: %T\n %s\n", err, err.Error())
			w.WriteHeader(http.StatusTeapot)
			return
		}
	}
	newData.Image = url
	log.Println("Before UPDATE: ", newData)
	user, err := h.hand.UpdateUser(newData)
	if err != nil {
		log.Printf("Update User data error: %T\n %s\n", err, err.Error())
		w.WriteHeader(http.StatusConflict)
		return
	}
	log.Println("Before AFTER: ", user)
	if user.Image != "" {
		user.Image = h.imagePath + user.Image
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := user.MarshalJSON()
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (h *Handler) Signout(w http.ResponseWriter, r *http.Request) {
	val, err := r.Cookie("session_id")
	if err != nil {
		log.Println("Error: cookie not found", val)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("Session: ", val)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().AddDate(0, -1, 0),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signin" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
	}
	data := model.SigninUser{}
	if err := data.UnmarshalJSON(body); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	user, err := h.hand.CheckUser(data)
	if err != nil {
		log.Printf("Check User Error: %T\n %s\n", err, err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
	log.Println("DEBUG Structure AFTER: ", data)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		log.Printf("Compare hash password error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := h.auth.Create(
		context.Background(),
		&auth.UserData{
			Login:    user.Nickname,
			Password: user.Password,
			Agent:    r.UserAgent(),
		})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    token.Value,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func GetSession(r *http.Request, authClient auth.AuthCheckerClient) (*auth.UserData, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}
	log.Println("CookSession IS", cookieSessionID.Value)
	sess, err := authClient.Check(
		context.Background(),
		&auth.Token{
			Value: cookieSessionID.Value,
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (h *Handler) CheckSession(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/session" {
		log.Println(r.URL.Path, "ERROR")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	session, err := GetSession(r, h.auth)
	if err != nil {
		log.Println("Error checking of session")
		w.WriteHeader(http.StatusBadRequest)
	}

	nickname := model.NicknameUser{
		Nickname: session.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := nickname.MarshalJSON()
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}