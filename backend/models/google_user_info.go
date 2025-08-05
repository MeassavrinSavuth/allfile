package models

type GoogleUserInfo struct {
    Sub           string `json:"sub"`
    Name          string `json:"name"`
    Email         string `json:"email"`
    Picture       string `json:"picture"`
    EmailVerified bool   `json:"email_verified"`
}
