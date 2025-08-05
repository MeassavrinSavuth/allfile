package controllers

import (
	"net/http"
	"social-sync-backend/lib"
	"github.com/google/uuid"
	// "log"
)

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max memory 10MB (adjust if needed)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique public ID for Cloudinary
	publicID := uuid.New().String()

	// Call your helper
	uploadedURL, err := lib.UploadToCloudinary(file, "socialsync_uploads", publicID)
	if err != nil {
		http.Error(w, "Failed to upload image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON with URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"url":"` + uploadedURL + `"}`))
}
