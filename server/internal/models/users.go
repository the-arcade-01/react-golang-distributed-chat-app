package models

type User struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserDetails struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
}

type UserLoginResponse struct {
	UserDetails
	Token string `json:"token"`
}
