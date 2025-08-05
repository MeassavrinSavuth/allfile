package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"social-sync-backend/models"
)

func GetGoogleUserInfo(idToken string) (*models.GoogleUserInfo, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid ID token")
	}

	var userInfo models.GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
