package router

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
	"os"
	"hash/fnv"
	"io"
	"io/ioutil"
	"github.com/DmitriyPrischep/backend-WAO/pkg/model"
	"github.com/DmitriyPrischep/backend-WAO/pkg/db"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/methods"
)

const (
	PathStaticServer = "./static"
)

func NewUserHandler(database *driver.DB) *Handler {
	return &Handler{
		hand: db.NewDataBase(database.DB),
	}
}

type Handler struct {
	hand methods.UserMethods
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
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var user model.UserRegister
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Email == "" || user.Nickname == "" || user.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nickname, err := h.hand.CreateUser(user)
	if err != nil {
		log.Printf("Error type: %T: %s\n", err, err.Error())
	}

	log.Println("New record NICK is:", nickname)

	sess, err := sessionManager.Create(
		context.Background(),
		&auth.UserData{
			Login: nickname,
			Agent: r.UserAgent(),
		})
	if err != nil {
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
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	user, err := h.hand.GetUser(model.NicknameUser{Nickname: params["login"]})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user.Image != "" {
		user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
	}
	json.NewEncoder(w).Encode(user)

	http.Error(w, `{"error": "This user is not found"}`, http.StatusNotFound)
}

func (h *Handler)ModifiedUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	params := mux.Vars(r)
	userLogin := params["login"]
	fmt.Println("userLogin: ", userLogin)
	if userLogin == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	newData := model.UpdateDataImport{}
	newData.Email = r.FormValue("email")
	newData.Password = r.FormValue("password")
	newData.Nickname = r.FormValue("nickname")
	var url string
	if _, _, err := r.FormFile("file"); err != nil {
		log.Println("Field with this name not exist")
	} else {
		url, err = uploadAvatar(r)
		if err != nil {
			log.Printf("Upload Error: %T\n %s\n", err, err.Error())
			w.WriteHeader(http.StatusTeapot)
			return
		}
	}
	newData.Image = url
	user, err := h.hand.UpdateUser(newData)
	if err != nil {
		log.Printf("Upload Error: %T\n %s\n", err, err.Error())
	}
}

func Signout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	val, err := r.Cookie("session_id")
	if err != nil {
		log.Println("Error: ", val)
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
	http.SetCookie(w, &http.Cookie{
		Name:     "VID",
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
	log.Println("Body: ", string(body))
	data := model.SigninUser{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Println(err)
	}
	log.Println("Structure: ", data)

	data = model.SigninUser{
		Nickname: r.FormValue("login"),
		Password: r.FormValue("password"),
	}
	log.Println("User -- ", data)

	user, err := h.hand.CheckUser(data)
	if err != nil {
		log.Printf("Upload Error: %T\n %s\n", err, err.Error())
	}

	token, err := sessionManager.Create(
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

func getSession(r *http.Request) (*auth.UserData, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}

	sess, err := sessionManager.Check(
		context.Background(),
		&auth.Token{
			Value: cookieSessionID.Value,
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func uploadAvatar(r *http.Request) (urlAvatar string, err error) {
	cookieVID, err := r.Cookie("VID")
	if err != nil {
		return "", err
	}
	log.Println("DEBUG", "CookieVID", cookieVID.Value)
	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		return "", err
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()
	log.Println("Molochnik")
	if _, err := os.Stat(PathStaticServer); os.IsNotExist(err) {
		err = os.Mkdir(PathStaticServer, 0700)
		if err != nil {
			return "", err
		}
	}
	dirname := cookieVID.Value
	if _, err := os.Stat(PathStaticServer + "/" + dirname); os.IsNotExist(err) {
		err = os.Mkdir(PathStaticServer+"/"+dirname, 0400)
		if err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}

	hash := fnv.New64a()
	hash.Write([]byte(handler.Filename + time.Now().Format("15:04:05.00000")))
	hashname := string(hash.Sum64())
	fmt.Println("HASH:", hashname)

	saveFile, err := os.Create(PathStaticServer + "/" + dirname + "/" + hashname)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, file)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return hashname, nil
}

func CheckSession(w http.ResponseWriter, r *http.Request) {
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
	session, err := getSession(r)
	if err != nil {
		log.Println("Error checking of session")
	}
	if session == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nickname := model.NicknameUser{
		Nickname: session.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nickname)
}