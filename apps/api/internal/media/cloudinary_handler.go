package media

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"newMap/pkg/response"
)

type CloudinarySignRequest struct {
	PublicID        string                 `json:"publicId" binding:"required"`
	Transformations map[string]interface{} `json:"transformations"`
}

type CloudinarySignResponse struct {
	SignedURL string `json:"signedUrl"`
	PublicID  string `json:"publicId"`
}

// SignCloudinaryURL generates a signed URL for private Cloudinary images
func SignCloudinaryURL(c *gin.Context) {
	var req CloudinarySignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get Cloudinary credentials from environment
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		response.Error(c, http.StatusInternalServerError, "Cloudinary configuration missing", nil)
		return
	}

	// Build transformation string
	var transformations []string
	for key, value := range req.Transformations {
		switch key {
		case "width":
			if w, ok := value.(float64); ok {
				transformations = append(transformations, fmt.Sprintf("w_%d", int(w)))
			}
		case "height":
			if h, ok := value.(float64); ok {
				transformations = append(transformations, fmt.Sprintf("h_%d", int(h)))
			}
		case "crop":
			if c, ok := value.(string); ok {
				transformations = append(transformations, fmt.Sprintf("c_%s", c))
			}
		case "quality":
			if q, ok := value.(string); ok {
				transformations = append(transformations, fmt.Sprintf("q_%s", q))
			} else if q, ok := value.(float64); ok {
				transformations = append(transformations, fmt.Sprintf("q_%d", int(q)))
			}
		case "format":
			if f, ok := value.(string); ok {
				transformations = append(transformations, fmt.Sprintf("f_%s", f))
			}
		case "gravity":
			if g, ok := value.(string); ok {
				transformations = append(transformations, fmt.Sprintf("g_%s", g))
			}
		}
	}

	// Create timestamp
	timestamp := time.Now().Unix()

	// Build parameters for signature
	params := map[string]string{
		"public_id": req.PublicID,
		"timestamp": strconv.FormatInt(timestamp, 10),
	}

	// Add transformation if present
	if len(transformations) > 0 {
		params["transformation"] = strings.Join(transformations, ",")
	}

	// Generate signature
	signature := generateSignature(params, apiSecret)

	// Build the signed URL
	baseURL := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload", cloudName)
	
	var urlParts []string
	if len(transformations) > 0 {
		urlParts = append(urlParts, strings.Join(transformations, ","))
	}
	
	// Add signature parameters
	signaturePart := fmt.Sprintf("a_%s,t_%d", signature, timestamp)
	urlParts = append(urlParts, signaturePart)
	
	// Add public ID
	urlParts = append(urlParts, req.PublicID)

	signedURL := fmt.Sprintf("%s/%s", baseURL, strings.Join(urlParts, "/"))

	response.Success(c, CloudinarySignResponse{
		SignedURL: signedURL,
		PublicID:  req.PublicID,
	})
}

// generateSignature creates a SHA1 signature for Cloudinary authentication
func generateSignature(params map[string]string, apiSecret string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, key := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", key, params[key]))
	}
	
	paramString := strings.Join(paramPairs, "&")
	stringToSign := paramString + apiSecret

	// Generate SHA1 hash
	hash := sha1.New()
	hash.Write([]byte(stringToSign))
	signature := fmt.Sprintf("%x", hash.Sum(nil))

	return signature
}