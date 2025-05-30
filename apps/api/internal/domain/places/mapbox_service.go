package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const mapboxGeocodingAPI = "https://api.mapbox.com/geocoding/v5/mapbox.places"

type MapboxService struct {
	apiKey     string
	httpClient *http.Client
}

func NewMapboxService(apiKey string) *MapboxService {
	return &MapboxService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type MapboxFeature struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	PlaceName  string                 `json:"place_name"`
	Properties map[string]interface{} `json:"properties"`
	Text       string                 `json:"text"`
	Center     []float64              `json:"center"` // [longitude, latitude]
	Geometry   struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Context []struct {
		ID    string `json:"id"`
		Text  string `json:"text"`
		Short string `json:"short_code,omitempty"`
	} `json:"context"`
}

type MapboxResponse struct {
	Type     string          `json:"type"`
	Features []MapboxFeature `json:"features"`
}

func (s *MapboxService) SearchPlaces(ctx context.Context, query string, limit int) ([]*Place, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("mapbox API key not configured")
	}

	// Build the request URL
	u, err := url.Parse(fmt.Sprintf("%s/%s.json", mapboxGeocodingAPI, url.QueryEscape(query)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("access_token", s.apiKey)
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("types", "poi,address,place,locality,neighborhood")
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mapbox API returned status %d", resp.StatusCode)
	}

	// Parse response
	var mapboxResp MapboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&mapboxResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Mapbox features to our Place model
	places := make([]*Place, 0, len(mapboxResp.Features))
	for _, feature := range mapboxResp.Features {
		place := s.featureToPlace(feature)
		places = append(places, place)
	}

	return places, nil
}

func (s *MapboxService) featureToPlace(feature MapboxFeature) *Place {
	place := &Place{
		ID:          feature.ID,
		Name:        feature.Text,
		Description: feature.PlaceName,
		Type:        "poi", // Default to POI
		Privacy:     "public",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set location coordinates
	if len(feature.Center) >= 2 {
		place.Location = &GeoPoint{
			Type:        "Point",
			Coordinates: feature.Center, // [longitude, latitude]
		}
	}

	// Extract address components from context
	for _, ctx := range feature.Context {
		if containsString(ctx.ID, "place") {
			place.City = ctx.Text
		} else if containsString(ctx.ID, "region") || containsString(ctx.ID, "state") {
			place.State = ctx.Text
		} else if containsString(ctx.ID, "country") {
			place.Country = ctx.Text
		} else if containsString(ctx.ID, "postcode") {
			place.PostalCode = ctx.Text
		}
	}

	// Set category based on feature type
	if feature.Type == "Feature" {
		place.Category = []string{extractCategory(feature.ID)}
	}

	// If it's an address, update the type
	if containsString(feature.ID, "address") {
		place.Type = "address"
		// Extract street address from the text
		place.StreetAddress = feature.Text
	}

	return place
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func extractCategory(id string) string {
	// Mapbox IDs are like "poi.1234567890"
	for idx := 0; idx < len(id); idx++ {
		if id[idx] == '.' {
			return id[:idx]
		}
	}
	return "place"
}