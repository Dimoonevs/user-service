package models

type UsersReq struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Code     string `json:"code,omitempty"`
}

type UserData struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
	IsVerify bool   `json:"is_verify"`
}
