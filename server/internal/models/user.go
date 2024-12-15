package models

type User struct {
	Username string `gorm:"column:username;primaryKey;size:50;not null;unique" json:"username"`
	Password string `gorm:"column:password;size:255;not null" json:"password"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}
