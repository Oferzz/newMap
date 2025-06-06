package media

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"newMap/pkg/response"
)

type CloudinaryConfigResponse struct {
	CloudName string `json:"cloudName"`
}

// GetCloudinaryConfig returns the Cloudinary cloud name for frontend use
func GetCloudinaryConfig(c *gin.Context) {
	// Get Cloudinary credentials from CLOUDINARY_URL environment variable
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		response.Error(c, http.StatusInternalServerError, "CLOUDINARY_URL environment variable not set", nil)
		return
	}

	// Parse the Cloudinary URL to extract cloud name
	cloudName, _, _, err := parseCloudinaryToken(cloudinaryURL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Invalid CLOUDINARY_URL format", err)
		return
	}

	response.Success(c, CloudinaryConfigResponse{
		CloudName: cloudName,
	})
}