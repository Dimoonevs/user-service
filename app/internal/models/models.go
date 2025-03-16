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

type UserSettings struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	UserID       int    `json:"user_id"`
	AIToken      string `json:"ai_token"`
	WhisperModel string `json:"whisper_model"`
	TTSModel     string `json:"tts_model"`
	GPTModel     string `json:"gpt_model"`
}
