package controllers

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "social-sync-backend/lib"
    "social-sync-backend/middleware"
    "social-sync-backend/models"
)

// ProfileHandler manages user profile GET, PUT, DELETE.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(middleware.UserIDKey).(string)

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    switch r.Method {
    case http.MethodGet:
        var user models.User

        // FIXED QUERY: Select columns in the correct order matching your database schema
        // Excluding password for security reasons
        err := lib.DB.QueryRow(`
            SELECT id, email, created_at, updated_at, is_verified, is_active, name, provider, provider_id, profile_picture
            FROM users WHERE id = $1
        `, userID).Scan(
            &user.ID,           // id (uuid)
            &user.Email,        // email (text)
            &user.CreatedAt,    // created_at (timestamp)
            &user.UpdatedAt,    // updated_at (timestamp) 
            &user.IsVerified,   // is_verified (boolean)
            &user.IsActive,     // is_active (boolean)
            &user.Name,         // name (text)
            &user.Provider,     // provider (character varying)
            &user.ProviderID,   // provider_id (character varying) - you'll need this in your struct
            &user.ProfilePicture, // profile_picture (text)
        )

        if err != nil {
            log.Printf("ERROR: ProfileHandler - Failed to fetch profile: %v", err)
            http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
            return
        }
        
        log.Println("DEBUG: Profile successfully fetched from database")

        // Encode and send response
        if err := json.NewEncoder(w).Encode(user); err != nil {
            log.Printf("ERROR: ProfileHandler - Failed to encode JSON response: %v", err)
            http.Error(w, "Failed to encode profile data", http.StatusInternalServerError)
            return
        }
        
        log.Println("INFO: Profile data successfully sent as JSON")

    case http.MethodPut:
        var updateData struct {
            Name  string `json:"name,omitempty"`
            Email string `json:"email,omitempty"`
        }

        if r.Body == nil {
            http.Error(w, "Empty request body", http.StatusBadRequest)
            return
        }
        if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
            http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
            return
        }

        query := "UPDATE users SET updated_at = NOW()"
        args := []interface{}{}
        argCount := 1

        if updateData.Name != "" {
            query += fmt.Sprintf(", name = $%d", argCount)
            args = append(args, updateData.Name)
            argCount++
        }
        if updateData.Email != "" {
            query += fmt.Sprintf(", email = $%d", argCount)
            args = append(args, updateData.Email)
            argCount++
        }

        if argCount == 1 { // No name or email provided
            http.Error(w, "No fields to update", http.StatusBadRequest)
            return
        }

        query += fmt.Sprintf(" WHERE id = $%d", argCount)
        args = append(args, userID)

        _, err := lib.DB.Exec(query, args...)
        if err != nil {
            log.Printf("Error updating profile: %v", err)
            http.Error(w, "Failed to update profile", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated"})

    case http.MethodDelete:
        _, err := lib.DB.Exec("DELETE FROM users WHERE id = $1", userID)
        if err != nil {
            log.Printf("Error deleting account: %v", err)
            http.Error(w, "Failed to delete account", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Account deleted"})

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// ProfileImageHandler handles POST requests for uploading profile pictures.
func ProfileImageHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(middleware.UserIDKey).(string)

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    switch r.Method {
    case http.MethodPost:
        maxMemory := int64(10 << 20)
        if err := r.ParseMultipartForm(maxMemory); err != nil {
            http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
            return
        }

        file, _, err := r.FormFile("profileImage")
        if err != nil {
            http.Error(w, "Failed to get file 'profileImage': "+err.Error(), http.StatusBadRequest)
            return
        }
        defer file.Close()

        folderName := "user_profile_pictures"
        imagePublicID := userID + "_main_profile_pic"

        imageURL, err := lib.UploadToCloudinary(file, folderName, imagePublicID)
        if err != nil {
            log.Printf("Cloudinary upload error: %v", err)
            http.Error(w, "Failed to upload image", http.StatusInternalServerError)
            return
        }

        _, err = lib.DB.Exec("UPDATE users SET profile_picture = $1 WHERE id = $2", imageURL, userID)
        if err != nil {
            log.Printf("Error saving image URL to DB: %v", err)
            http.Error(w, "Failed to save image URL", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "message":  "Profile image uploaded",
            "imageUrl": imageURL,
        })

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}