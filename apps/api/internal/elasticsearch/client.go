package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Client struct {
	es *elasticsearch.Client
}

type SearchResult struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // "activity" or "place"
	Source map[string]interface{} `json:"source"`
	Score  float64                `json:"score"`
}

type SearchResponse struct {
	Total   int64          `json:"total"`
	Results []SearchResult `json:"results"`
	Took    int            `json:"took"`
}

// NewClient creates a new Elasticsearch client
func NewClient() (*Client, error) {
	var esURL string
	if url := os.Getenv("ELASTICSEARCH_URL"); url != "" {
		esURL = url
	} else {
		esURL = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	client := &Client{es: es}

	// Test connection
	if err := client.ping(); err != nil {
		log.Printf("Warning: Elasticsearch not available: %v", err)
		// Don't fail - return client anyway for graceful degradation
	}

	return client, nil
}

// ping tests the Elasticsearch connection
func (c *Client) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.es.Ping(c.es.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.Status())
	}

	return nil
}

// IndexActivity indexes an activity document
func (c *Client) IndexActivity(ctx context.Context, activityID string, activity map[string]interface{}) error {
	body, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      "activities",
		DocumentID: activityID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index activity: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.Status())
	}

	return nil
}

// IndexPlace indexes a place document
func (c *Client) IndexPlace(ctx context.Context, placeID string, place map[string]interface{}) error {
	body, err := json.Marshal(place)
	if err != nil {
		return fmt.Errorf("failed to marshal place: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      "places",
		DocumentID: placeID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to index place: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.Status())
	}

	return nil
}

// SearchUnified performs a unified search across activities and places
func (c *Client) SearchUnified(ctx context.Context, query map[string]interface{}) (*SearchResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex("activities", "places"),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s - %s", res.Status(), string(body))
	}

	var response struct {
		Took int `json:"took"`
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Index  string                 `json:"_index"`
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]SearchResult, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		docType := "activity"
		if hit.Index == "places" {
			docType = "place"
		}
		
		results[i] = SearchResult{
			ID:     hit.ID,
			Type:   docType,
			Source: hit.Source,
			Score:  hit.Score,
		}
	}

	return &SearchResponse{
		Total:   response.Hits.Total.Value,
		Results: results,
		Took:    response.Took,
	}, nil
}

// SearchActivities searches only activities
func (c *Client) SearchActivities(ctx context.Context, query map[string]interface{}) (*SearchResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex("activities"),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s - %s", res.Status(), string(body))
	}

	return c.parseSearchResponse(res.Body, "activity")
}

// SearchPlaces searches only places
func (c *Client) SearchPlaces(ctx context.Context, query map[string]interface{}) (*SearchResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex("places"),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s - %s", res.Status(), string(body))
	}

	return c.parseSearchResponse(res.Body, "place")
}

// parseSearchResponse parses the Elasticsearch response
func (c *Client) parseSearchResponse(body io.Reader, docType string) (*SearchResponse, error) {
	var response struct {
		Took int `json:"took"`
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]SearchResult, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		results[i] = SearchResult{
			ID:     hit.ID,
			Type:   docType,
			Source: hit.Source,
			Score:  hit.Score,
		}
	}

	return &SearchResponse{
		Total:   response.Hits.Total.Value,
		Results: results,
		Took:    response.Took,
	}, nil
}

// DeleteDocument deletes a document from an index
func (c *Client) DeleteDocument(ctx context.Context, index, documentID string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete failed: %s", res.Status())
	}

	return nil
}

// BuildQuery builds an Elasticsearch query from search parameters
func BuildQuery(searchText string, filters map[string]interface{}, limit, offset int) map[string]interface{} {
	query := map[string]interface{}{
		"size": limit,
		"from": offset,
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	// Build the main query
	var queryClause map[string]interface{}

	if searchText != "" {
		// Multi-match query for text search
		queryClause = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  searchText,
							"fields": []string{"title^3", "description^2", "name^3"},
							"type":   "best_fields",
							"fuzziness": "AUTO",
						},
					},
				},
				"filter": buildFilters(filters),
			},
		}
	} else {
		// Match all with filters
		queryClause = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"match_all": map[string]interface{}{}},
				},
				"filter": buildFilters(filters),
			},
		}
	}

	query["query"] = queryClause
	return query
}

// buildFilters builds the filter clauses for the query
func buildFilters(filters map[string]interface{}) []map[string]interface{} {
	var filterClauses []map[string]interface{}

	for key, value := range filters {
		switch key {
		case "activity_types":
			if types, ok := value.([]string); ok && len(types) > 0 {
				filterClauses = append(filterClauses, map[string]interface{}{
					"terms": map[string]interface{}{
						"activity_type": types,
					},
				})
			}
		case "difficulty_levels":
			if levels, ok := value.([]string); ok && len(levels) > 0 {
				filterClauses = append(filterClauses, map[string]interface{}{
					"terms": map[string]interface{}{
						"difficulty_level": levels,
					},
				})
			}
		case "water_features":
			if features, ok := value.([]string); ok && len(features) > 0 {
				filterClauses = append(filterClauses, map[string]interface{}{
					"terms": map[string]interface{}{
						"water_features": features,
					},
				})
			}
		case "visibility":
			if vis, ok := value.(string); ok && vis != "" {
				filterClauses = append(filterClauses, map[string]interface{}{
					"term": map[string]interface{}{
						"visibility": vis,
					},
				})
			}
		case "location":
			if location, ok := value.(map[string]interface{}); ok {
				if lat, latOk := location["lat"].(float64); latOk {
					if lng, lngOk := location["lng"].(float64); lngOk {
						if radius, radiusOk := location["radius"].(float64); radiusOk {
							filterClauses = append(filterClauses, map[string]interface{}{
								"geo_distance": map[string]interface{}{
									"distance": fmt.Sprintf("%.0fkm", radius),
									"location": map[string]float64{
										"lat": lat,
										"lon": lng,
									},
								},
							})
						}
					}
				}
			}
		}
	}

	return filterClauses
}

// LogQuery logs a search query for analytics
func (c *Client) LogQuery(ctx context.Context, queryLog map[string]interface{}) error {
	body, err := json.Marshal(queryLog)
	if err != nil {
		return fmt.Errorf("failed to marshal query log: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   "search_queries",
		Body:    bytes.NewReader(body),
		Refresh: "false", // Don't need immediate refresh for logs
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		// Log but don't fail the main operation
		log.Printf("Failed to log search query: %v", err)
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Query logging failed: %s", res.Status())
	}

	return nil
}

// IsAvailable checks if Elasticsearch is available
func (c *Client) IsAvailable() bool {
	return c.ping() == nil
}