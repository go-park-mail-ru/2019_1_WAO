package model

type UserRegister struct {
	Email    string `json:"email, omitempty"`
	Password string `json:"password, omitempty"`
	Nickname string `json:"nickname, omitempty"`
}

type SendUsers struct {
	Users []User `json:"users, string, omitempty"`
}

type User struct {
	ID       int    `json:"id, string, omitempty"`
	Email    string `json:"email, omitempty"`
	password string `json:"-,"`
	Nick     string `json:"nickname, omitempty"`
	Score    int    `json:"score, string, omitempty"`
	Games    int    `json:"games, string, omitempty"`
	Wins     int    `json:"wins, string, omitempty"`
	Image    string `json:"image, omitempty"`
}

type SigninUser struct {
	Nickname string `json:"nickname, omitempty"`
	Password string `json:"password, omitempty"`
}

type NicknameUser struct {
	Nickname string `json:"nickname, omitempty"`
}

type UpdateDataImport struct {
	Nickname string `json:"nickname, omitempty"`
	Password string `json:"password, omitempty"`
	Email    string `json:"email, omitempty"`
	Image    string `json:"image, omitempty"`
	OldNick    string `json:"oldnick, omitempty"`
}

type UpdateDataExport struct {
	Email    string `json:"email, omitempty"`
	Nickname string `json:"nickname, omitempty"`
	Image    string `json:"image, omitempty"`
}
