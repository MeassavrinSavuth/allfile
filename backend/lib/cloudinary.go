package lib

import (
	"context"
	// "io" // io.ReadAll is no longer used, so 'io' import can be removed if not used elsewhere
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var Cloud *cloudinary.Cloudinary

// Initialize Cloudinary client (call this in main.go)
func InitCloudinary() error {
	var err error
	Cloud, err = cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	return err
}

// Upload to Cloudinary and return the secure URL
func UploadToCloudinary(file multipart.File, folder string, publicID string) (string, error) {
	// REMOVED: fileBytes, err := io.ReadAll(file)
	// If you remove io.ReadAll, you can remove the "io" import as well if it's not used anywhere else in this file.
	// if err != nil {
	//     return "", err
	// }

	// FIX: Pass 'file' directly (multipart.File implements io.Reader, which is a supported source type)
	uploadResult, err := Cloud.Upload.Upload(context.Background(), file, uploader.UploadParams{
		PublicID:     folder + "/" + publicID,
		Folder:       folder,
		Overwrite:    api.Bool(true),
		ResourceType: "auto", // auto-detects image, video, etc.
	})
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}