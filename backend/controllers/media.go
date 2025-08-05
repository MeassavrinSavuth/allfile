package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"social-sync-backend/lib"
	"social-sync-backend/middleware"
	"social-sync-backend/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// MediaUploadRequest represents the request for uploading media
type MediaUploadRequest struct {
	Tags []string `json:"tags"`
}

// MediaListResponse represents the response for listing media
type MediaListResponse struct {
	Media []models.Media `json:"media"`
	Total int            `json:"total"`
}

// UploadMedia handles media upload to workspace
func UploadMedia(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	// Parse multipart form with max memory 50MB
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		log.Println("Failed to parse multipart form:", err)
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("Failed to get file from request:", err)
		http.Error(w, "Failed to get file from request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get tags from form data
	tags := []string{}
	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			log.Println("Invalid tags format:", err)
			http.Error(w, "Invalid tags format", http.StatusBadRequest)
			return
		}
	}

	// Validate file type
	fileType := getFileType(header.Filename)
	if fileType == "" {
		log.Println("Unsupported file type")
		http.Error(w, "Unsupported file type. Supported: images (jpg,jpeg,png,gif,webp) and videos (mp4,mov,avi,mkv,wmv,flv,webm)", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + filepath.Ext(header.Filename)
	publicID := fmt.Sprintf("workspaces/%s/media/%s", workspaceID, filename)

	// Upload to Cloudinary
	cloudinaryURL, err := lib.UploadToCloudinary(file, "socialsync_uploads", publicID)
	if err != nil {
		log.Println("Failed to upload to Cloudinary:", err)
		http.Error(w, "Failed to upload to Cloudinary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get file size
	fileSize := header.Size

	// Create media record in database
	mediaID := uuid.New().String()
	now := time.Now()

	_, err = lib.DB.Exec(`
		INSERT INTO media (
			id, workspace_id, uploaded_by, filename, original_name, file_url, 
			file_type, mime_type, file_size, tags, cloudinary_public_id, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, mediaID, workspaceID, userID, filename, header.Filename, cloudinaryURL,
		fileType, header.Header.Get("Content-Type"), fileSize, tags, publicID, now, now)

	if err != nil {
		log.Println("Failed to save media record:", err)
		http.Error(w, "Failed to save media record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the created media with uploader info
	var media models.Media
	err = lib.DB.QueryRow(`
		SELECT m.id, m.workspace_id, m.uploaded_by, m.filename, m.original_name, 
		       m.file_url, m.file_type, m.mime_type, m.file_size, m.width, m.height, 
		       m.duration, m.tags, m.cloudinary_public_id, m.created_at, m.updated_at,
		       u.name as uploader_name
		FROM media m
		LEFT JOIN users u ON m.uploaded_by = u.id
		WHERE m.id = $1
	`, mediaID).Scan(
		&media.ID, &media.WorkspaceID, &media.UploadedBy, &media.Filename,
		&media.OriginalName, &media.FileURL, &media.FileType, &media.MimeType,
		&media.FileSize, &media.Width, &media.Height, &media.Duration,
		&media.Tags, &media.CloudinaryPublicID, &media.CreatedAt, &media.UpdatedAt,
		&media.UploaderName,
	)

	if err != nil {
		log.Println("Failed to retrieve created media:", err)
		http.Error(w, "Failed to retrieve created media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created media
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)

	// Broadcast the event to all workspace clients
	msg, _ := json.Marshal(map[string]interface{}{
		"type":  "media_uploaded",
		"media": media,
	})
	hub.broadcast(workspaceID, websocket.TextMessage, msg)
}

// ListMedia handles listing media for a workspace
func ListMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	// Get query parameters
	fileType := r.URL.Query().Get("type") // "image", "video", or empty for all
	search := r.URL.Query().Get("search")
	tag := r.URL.Query().Get("tag")
	limit := 50 // Default limit
	offset := 0

	// Build query
	query := `
		SELECT m.id, m.workspace_id, m.uploaded_by, m.filename, m.original_name, 
		       m.file_url, m.file_type, m.mime_type, m.file_size, m.width, m.height, 
		       m.duration, m.tags, m.cloudinary_public_id, m.created_at, m.updated_at,
		       u.name as uploader_name
		FROM media m
		LEFT JOIN users u ON m.uploaded_by = u.id
		WHERE m.workspace_id = $1
	`
	args := []interface{}{workspaceID}
	argIndex := 2

	// Add filters
	if fileType != "" {
		query += fmt.Sprintf(" AND m.file_type = $%d", argIndex)
		args = append(args, fileType)
		argIndex++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (m.original_name ILIKE $%d OR m.filename ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(m.tags)", argIndex)
		args = append(args, tag)
		argIndex++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY m.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := lib.DB.Query(query, args...)
	if err != nil {
		log.Println("Failed to query media:", err)
		http.Error(w, "Failed to query media: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var mediaList []models.Media
	for rows.Next() {
		var media models.Media
		err := rows.Scan(
			&media.ID, &media.WorkspaceID, &media.UploadedBy, &media.Filename,
			&media.OriginalName, &media.FileURL, &media.FileType, &media.MimeType,
			&media.FileSize, &media.Width, &media.Height, &media.Duration,
			&media.Tags, &media.CloudinaryPublicID, &media.CreatedAt, &media.UpdatedAt,
			&media.UploaderName,
		)
		if err != nil {
			log.Println("Failed to scan media row:", err)
			http.Error(w, "Failed to scan media row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mediaList = append(mediaList, media)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM media m
		WHERE m.workspace_id = $1
	`
	countArgs := []interface{}{workspaceID}
	countArgIndex := 2

	if fileType != "" {
		countQuery += fmt.Sprintf(" AND m.file_type = $%d", countArgIndex)
		countArgs = append(countArgs, fileType)
		countArgIndex++
	}

	if search != "" {
		countQuery += fmt.Sprintf(" AND (m.original_name ILIKE $%d OR m.filename ILIKE $%d)", countArgIndex, countArgIndex)
		countArgs = append(countArgs, "%"+search+"%")
		countArgIndex++
	}

	if tag != "" {
		countQuery += fmt.Sprintf(" AND $%d = ANY(m.tags)", countArgIndex)
		countArgs = append(countArgs, tag)
		countArgIndex++
	}

	var total int
	err = lib.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Println("Failed to get media count:", err)
		http.Error(w, "Failed to get media count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := MediaListResponse{
		Media: mediaList,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteMedia handles deleting media from workspace
func DeleteMedia(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]
	mediaID := vars["mediaId"]

	// Check if user has permission to delete this media
	var uploadedBy string
	err := lib.DB.QueryRow(`
		SELECT uploaded_by FROM media 
		WHERE id = $1 AND workspace_id = $2
	`, mediaID, workspaceID).Scan(&uploadedBy)

	if err == sql.ErrNoRows {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Failed to check media ownership:", err)
		http.Error(w, "Failed to check media ownership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Only the uploader can delete the media
	if uploadedBy != userID {
		http.Error(w, "Unauthorized to delete this media", http.StatusForbidden)
		return
	}

	// Get cloudinary public ID for deletion
	var cloudinaryPublicID string
	err = lib.DB.QueryRow(`
		SELECT cloudinary_public_id FROM media WHERE id = $1
	`, mediaID).Scan(&cloudinaryPublicID)

	if err != nil {
		log.Println("Failed to get media info:", err)
		http.Error(w, "Failed to get media info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete from Cloudinary (optional - you might want to keep files for a while)
	// if cloudinaryPublicID != "" {
	//     // Add Cloudinary deletion logic here
	// }

	// Delete from database
	_, err = lib.DB.Exec("DELETE FROM media WHERE id = $1", mediaID)
	if err != nil {
		log.Println("Failed to delete media:", err)
		http.Error(w, "Failed to delete media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMediaTags handles updating media tags
func UpdateMediaTags(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]
	mediaID := vars["mediaId"]

	var req struct {
		Tags []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user has permission to update this media
	var uploadedBy string
	err := lib.DB.QueryRow(`
		SELECT uploaded_by FROM media 
		WHERE id = $1 AND workspace_id = $2
	`, mediaID, workspaceID).Scan(&uploadedBy)

	if err == sql.ErrNoRows {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Failed to check media ownership:", err)
		http.Error(w, "Failed to check media ownership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Only the uploader can update the media
	if uploadedBy != userID {
		http.Error(w, "Unauthorized to update this media", http.StatusForbidden)
		return
	}

	// Update tags
	_, err = lib.DB.Exec(`
		UPDATE media SET tags = $1, updated_at = $2 WHERE id = $3
	`, req.Tags, time.Now(), mediaID)

	if err != nil {
		log.Println("Failed to update media tags:", err)
		http.Error(w, "Failed to update media tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to determine file type
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	videoExts := []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "image"
		}
	}

	for _, vidExt := range videoExts {
		if ext == vidExt {
			return "video"
		}
	}

	return ""
}
