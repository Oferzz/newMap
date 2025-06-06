package media

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

type CloudinaryConfigResponse struct {
	CloudName string `json:"cloudName"`
}

// GetCloudinaryConfig returns the Cloudinary cloud name for frontend use
func GetCloudinaryConfig(c *gin.Context) {
	// Get Cloudinary credentials from CLOUDINARY_URL environment variable
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		response.InternalServerError(c, "CLOUDINARY_URL environment variable not set")
		return
	}

	// Parse the Cloudinary URL to extract cloud name
	cloudName, _, _, err := parseCloudinaryToken(cloudinaryURL)
	if err != nil {
		response.InternalServerError(c, "Invalid CLOUDINARY_URL format: "+err.Error())
		return
	}

	response.Success(c, CloudinaryConfigResponse{
		CloudName: cloudName,
	})
}