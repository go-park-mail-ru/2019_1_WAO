package methods

import (
	"github.com/DmitriyPrischep/backend-WAO/pkg/model"
)

type UserMethods interface {
	GetUser(userdata model.NicknameUser) (user *model.User, err error)
	GetUsers() (users []model.Player, err error)
	CreateUser(user model.UserRegister) (out *model.RegUser, err error)
	UpdateUser(user model.UpdateDataImport) (out model.UpdateDataExport, err error)
	CheckUser(user model.SigninUser) (out *model.UserRegister, err error)
	ChangeData(user model.GameInfo) (err error)
}