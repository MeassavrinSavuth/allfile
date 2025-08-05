package controllers

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "os"

    "social-sync-backend/lib"
    "social-sync-backend/models"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

func getGoogleOAuthConfig() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
        Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
        Endpoint:     google.Endpoint,
    }
}

func GoogleRedirectHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        config := getGoogleOAuthConfig()
        url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        http.Redirect(w, r, url, http.StatusTemporaryRedirect)
    }
}

func GoogleCallbackHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        if code == "" {
            http.Error(w, "Missing code", http.StatusBadRequest)
            return
        }

        config := getGoogleOAuthConfig()
        token, err := config.Exchange(context.Background(), code)
        if err != nil {
            http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
            return
        }

        client := config.Client(context.Background(), token)
        resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
        if err != nil {
            http.Error(w, "Failed to get user info", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        bodyBytes, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            http.Error(w, "Failed to read user info", http.StatusInternalServerError)
            return
        }

        resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

        var userInfo models.GoogleUserInfo
        if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
            http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
            return
        }

        var userID string
        err = db.QueryRow(`
            SELECT id FROM users
            WHERE provider = 'google' AND provider_id = $1
        `, userInfo.Sub).Scan(&userID)

        if err == sql.ErrNoRows {
            err = db.QueryRow(`
                INSERT INTO users (name, email, provider, provider_id, profile_picture, is_verified, is_active, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, true, true, NOW(), NOW())
                RETURNING id
            `, userInfo.Name, userInfo.Email, "google", userInfo.Sub, userInfo.Picture).Scan(&userID)

            if err != nil {
                http.Error(w, "Failed to create user", http.StatusInternalServerError)
                return
            }
        } else if err != nil {
            http.Error(w, "DB error", http.StatusInternalServerError)
            return
        }

        accessToken, err := lib.GenerateAccessToken(userID)
        if err != nil {
            http.Error(w, "Token error", http.StatusInternalServerError)
            return
        }

        refreshToken, err := lib.GenerateRefreshToken(userID)
        if err != nil {
            http.Error(w, "Token error", http.StatusInternalServerError)
            return
        }

        redirectURL := "http://localhost:3000/auth/callback?access_token=" + accessToken + "&refresh_token=" + refreshToken
        http.Redirect(w, r, redirectURL, http.StatusSeeOther)
    }
}
