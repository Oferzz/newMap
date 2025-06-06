package media

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

type CloudinarySignRequest struct {
	PublicID        string                 `json:"publicId" binding:"required"`
	Transformations map[string]interface{} `json:"transformations"`
}

type CloudinaryListRequest struct {
	Folder     string `json:"folder"`
	Collection string `json:"collection"`
	MaxImages  int    `json:"maxImages"`
}

type CloudinarySignResponse struct {
	SignedURL string `json:"signedUrl"`
	PublicID  string `json:"publicId"`
}

type CloudinaryImage struct {
	PublicID  string `json:"publicId"`
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	CreatedAt string `json:"createdAt"`
	Tags      []string `json:"tags,omitempty"`
}

type CloudinaryListResponse struct {
	Resources []struct {
		PublicID  string   `json:"public_id"`
		Format    string   `json:"format"`
		Width     int      `json:"width"`
		Height    int      `json:"height"`
		CreatedAt string   `json:"created_at"`
		Tags      []string `json:"tags"`
	} `json:"resources"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// SignCloudinaryURL generates a signed URL for private Cloudinary images
func SignCloudinaryURL(c *gin.Context) {
	var req CloudinarySignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Get Cloudinary credentials from CLOUDINARY_URL environment variable
	// Expected format: cloudinary://api_key:api_secret@cloud_name
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		response.InternalServerError(c, "CLOUDINARY_URL environment variable not set")
		return
	}

	// Parse the Cloudinary URL
	cloudName, _, apiSecret, err := parseCloudinaryToken(cloudinaryURL)
	if err != nil {
		response.InternalServerError(c, "Invalid CLOUDINARY_URL format: "+err.Error())
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

// parseCloudinaryToken parses the CLOUDINARY_TOKEN environment variable
// Expected format: cloudinary://api_key:api_secret@cloud_name
func parseCloudinaryToken(token string) (cloudName, apiKey, apiSecret string, err error) {
	// Remove the cloudinary:// prefix if present
	if strings.HasPrefix(token, "cloudinary://") {
		token = strings.TrimPrefix(token, "cloudinary://")
	}
	
	// Split by @ to separate credentials from cloud name
	parts := strings.Split(token, "@")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid token format: expected format cloudinary://api_key:api_secret@cloud_name")
	}
	
	cloudName = parts[1]
	credentials := parts[0]
	
	// Split credentials by : to separate api_key from api_secret
	credParts := strings.Split(credentials, ":")
	if len(credParts) != 2 {
		return "", "", "", fmt.Errorf("invalid credentials format: expected api_key:api_secret")
	}
	
	apiKey = credParts[0]
	apiSecret = credParts[1]
	
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", "", "", fmt.Errorf("missing required credentials: cloud_name, api_key, and api_secret are all required")
	}
	
	return cloudName, apiKey, apiSecret, nil
}

// ListCloudinaryImages fetches images from a specific folder or collection
func ListCloudinaryImages(c *gin.Context) {
	var req CloudinaryListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Validate that either folder or collection is provided
	if req.Folder == "" && req.Collection == "" {
		response.BadRequest(c, "Either 'folder' or 'collection' must be specified")
		return
	}

	// Get Cloudinary credentials from CLOUDINARY_URL environment variable
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		response.InternalServerError(c, "CLOUDINARY_URL environment variable not set")
		return
	}

	// Parse the Cloudinary URL
	cloudName, apiKey, apiSecret, err := parseCloudinaryToken(cloudinaryURL)
	if err != nil {
		response.InternalServerError(c, "Invalid CLOUDINARY_URL format: "+err.Error())
		return
	}

	// Set default max images if not specified
	maxImages := req.MaxImages
	if maxImages <= 0 || maxImages > 500 {
		maxImages = 100
	}

	// Call Cloudinary Admin API to list resources
	var images []CloudinaryImage
	var sourceType, sourceName string

	if req.Collection != "" {
		images, err = listCollectionImages(cloudName, apiKey, apiSecret, req.Collection, maxImages)
		sourceType = "collection"
		sourceName = req.Collection
	} else {
		images, err = listFolderImages(cloudName, apiKey, apiSecret, req.Folder, maxImages)
		sourceType = "folder"
		sourceName = req.Folder
	}

	if err != nil {
		response.InternalServerError(c, "Failed to fetch images from Cloudinary: "+err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"images":     images,
		"sourceType": sourceType,
		"sourceName": sourceName,
		"count":      len(images),
	})
}

// listFolderImages calls Cloudinary Admin API to list images in a folder
func listFolderImages(cloudName, apiKey, apiSecret, folder string, maxImages int) ([]CloudinaryImage, error) {
	// Build the Admin API URL
	apiURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/image", cloudName)
	
	// Prepare query parameters
	params := url.Values{}
	params.Set("type", "upload")
	params.Set("prefix", folder+"/") // folder prefix
	params.Set("max_results", strconv.Itoa(maxImages))
	
	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add basic auth
	req.SetBasicAuth(apiKey, apiSecret)
	
	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudinary API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var listResp CloudinaryListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to our format
	images := make([]CloudinaryImage, len(listResp.Resources))
	for i, resource := range listResp.Resources {
		images[i] = CloudinaryImage{
			PublicID:  resource.PublicID,
			Format:    resource.Format,
			Width:     resource.Width,
			Height:    resource.Height,
			CreatedAt: resource.CreatedAt,
			Tags:      resource.Tags,
		}
	}
	
	return images, nil
}

// listCollectionImages calls Cloudinary Admin API to list images in a collection
func listCollectionImages(cloudName, apiKey, apiSecret, collectionName string, maxImages int) ([]CloudinaryImage, error) {
	// Build the Admin API URL for searching resources
	// Use the search API which supports collections
	apiURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/search", cloudName)
	
	// Prepare search expression for collection
	searchExpression := fmt.Sprintf("collection=%s", collectionName)
	
	// Prepare query parameters
	params := url.Values{}
	params.Set("expression", searchExpression)
	params.Set("max_results", strconv.Itoa(maxImages))
	params.Set("with_field", "tags")
	
	fullURL := apiURL + "?" + params.Encode()
	fmt.Printf("Making request to Cloudinary collection API: %s\n", fullURL)
	fmt.Printf("Using credentials - Cloud Name: %s, API Key: %s, Collection: %s\n", cloudName, apiKey, collectionName)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add basic auth
	req.SetBasicAuth(apiKey, apiSecret)
	
	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Cloudinary API response status: %d\n", resp.StatusCode)
	fmt.Printf("Cloudinary API response body: %s\n", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudinary API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var listResp CloudinaryListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	fmt.Printf("Parsed %d resources from Cloudinary response\n", len(listResp.Resources))
	
	// Convert to our format
	images := make([]CloudinaryImage, len(listResp.Resources))
	for i, resource := range listResp.Resources {
		images[i] = CloudinaryImage{
			PublicID:  resource.PublicID,
			Format:    resource.Format,
			Width:     resource.Width,
			Height:    resource.Height,
			CreatedAt: resource.CreatedAt,
			Tags:      resource.Tags,
		}
	}
	
	return images, nil
}