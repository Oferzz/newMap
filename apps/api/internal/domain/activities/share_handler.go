package activities

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

// ShareLink represents a shareable link for an activity
type ShareLink struct {
	ID         string    `json:"id"`
	ActivityID string    `json:"activity_id"`
	Token      string    `json:"token"`
	URL        string    `json:"url"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
	ViewCount  int       `json:"view_count"`
	Settings   ShareSettings `json:"settings"`
}

// ShareSettings represents permissions for a share link
type ShareSettings struct {
	AllowComments   bool `json:"allow_comments"`
	AllowDownloads  bool `json:"allow_downloads"`
	RequirePassword bool `json:"require_password"`
	Password        string `json:"password,omitempty"`
	MaxViews        *int `json:"max_views,omitempty"`
}

// GenerateShareLink generates a shareable link for an activity
// @Summary Generate share link for activity
// @Description Create a shareable link for an activity with custom permissions
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param body body ShareSettings true "Share settings"
// @Success 200 {object} response.Response{data=ShareLink}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/activities/{id}/share [post]
func (h *Handler) GenerateShareLink(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")
	
	var settings ShareSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "Invalid share settings")
		return
	}

	// Verify the user owns the activity or has permission to share
	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	if activity.CreatedBy != userID.(string) {
		response.Forbidden(c, "You don't have permission to share this activity")
		return
	}

	// Generate secure token
	token := generateSecureToken()
	
	// Create share link
	shareLink := &ShareLink{
		ID:         generateID(),
		ActivityID: activityID,
		Token:      token,
		URL:        fmt.Sprintf("%s/activities/shared/%s", getBaseURL(c), token),
		CreatedAt:  time.Now(),
		CreatedBy:  userID.(string),
		ViewCount:  0,
		Settings:   settings,
	}

	// In real implementation, save to database
	// For now, return the generated link
	response.Success(c, shareLink)
}

// GetSharedActivity retrieves an activity via share link
// @Summary Get activity via share link
// @Description Access an activity using a share token
// @Tags activities
// @Accept json
// @Produce json
// @Param token path string true "Share token"
// @Param password query string false "Password if required"
// @Success 200 {object} response.Response{data=Activity}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/shared/{token} [get]
func (h *Handler) GetSharedActivity(c *gin.Context) {
	token := c.Param("token")
	password := c.Query("password")

	// In real implementation, look up share link by token
	// Verify it's not expired, check password if required, increment view count
	
	// For now, return a mock response
	response.Success(c, map[string]interface{}{
		"message": "Shared activity access",
		"token":   token,
		"note":    "This would return the activity data based on share permissions",
	})
}

// RevokeShareLink revokes a share link
// @Summary Revoke share link
// @Description Revoke an existing share link for an activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param linkId path string true "Share link ID"
// @Success 204
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/{id}/share/{linkId} [delete]
func (h *Handler) RevokeShareLink(c *gin.Context) {
	activityID := c.Param("id")
	linkID := c.Param("linkId")
	userID, _ := c.Get("user_id")

	// Verify ownership
	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	if activity.CreatedBy != userID.(string) {
		response.Forbidden(c, "You don't have permission to manage share links")
		return
	}

	// In real implementation, delete the share link from database
	_ = linkID // Use the linkID

	response.NoContent(c)
}

// ListShareLinks lists all share links for an activity
// @Summary List share links
// @Description Get all share links for an activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} response.Response{data=[]ShareLink}
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/{id}/share [get]
func (h *Handler) ListShareLinks(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify ownership
	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	if activity.CreatedBy != userID.(string) {
		response.Forbidden(c, "You don't have permission to view share links")
		return
	}

	// In real implementation, fetch from database
	shareLinks := []ShareLink{
		{
			ID:         "link1",
			ActivityID: activityID,
			Token:      "sample-token",
			URL:        fmt.Sprintf("%s/activities/shared/sample-token", getBaseURL(c)),
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			CreatedBy:  userID.(string),
			ViewCount:  42,
			Settings: ShareSettings{
				AllowComments:  true,
				AllowDownloads: false,
			},
		},
	}

	response.Success(c, shareLinks)
}

// Helper functions
func generateSecureToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getBaseURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}