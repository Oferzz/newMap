package search

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/elasticsearch"
	"github.com/Oferzz/newMap/apps/api/internal/nlp"
)

// Service handles unified search across activities and places
type Service struct {
	esClient  *elasticsearch.Client
	nlpParser *nlp.Parser
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query     string `json:"query" binding:"required"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	UserID    string `json:"-"` // Set from auth context
	SessionID string `json:"session_id,omitempty"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query       *nlp.ParsedQuery              `json:"query"`
	Results     []elasticsearch.SearchResult  `json:"results"`
	Total       int64                         `json:"total"`
	Took        int                           `json:"took"`
	Suggestions []string                      `json:"suggestions,omitempty"`
}

// NewService creates a new search service
func NewService(esClient *elasticsearch.Client, nlpParser *nlp.Parser) *Service {
	return &Service{
		esClient:  esClient,
		nlpParser: nlpParser,
	}
}

// Search performs a unified natural language search
func (s *Service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Parse the natural language query
	parsedQuery, err := s.nlpParser.ParseQuery(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Add user-specific filters for visibility
	s.addVisibilityFilters(parsedQuery, req.UserID)

	// Build Elasticsearch query
	esQuery := s.buildElasticsearchQuery(parsedQuery, req.Limit, req.Offset)

	// Execute search based on intent
	var esResponse *elasticsearch.SearchResponse
	if s.esClient.IsAvailable() {
		switch parsedQuery.Intent {
		case nlp.IntentActivity:
			esResponse, err = s.esClient.SearchActivities(ctx, esQuery)
		case nlp.IntentPlace:
			esResponse, err = s.esClient.SearchPlaces(ctx, esQuery)
		default:
			// Mixed or unknown - search both
			esResponse, err = s.esClient.SearchUnified(ctx, esQuery)
		}

		if err != nil {
			log.Printf("Elasticsearch search failed: %v", err)
			// Fallback to database search
			esResponse = s.fallbackSearch(ctx, parsedQuery, req)
		}
	} else {
		// Elasticsearch not available - use database fallback
		esResponse = s.fallbackSearch(ctx, parsedQuery, req)
	}

	// Generate search suggestions
	suggestions := s.generateSuggestions(parsedQuery, esResponse)

	// Log the search for analytics (async)
	go s.logSearch(context.Background(), req, parsedQuery, esResponse)

	return &SearchResponse{
		Query:       parsedQuery,
		Results:     esResponse.Results,
		Total:       esResponse.Total,
		Took:        esResponse.Took,
		Suggestions: suggestions,
	}, nil
}

// addVisibilityFilters adds user-specific visibility filters
func (s *Service) addVisibilityFilters(parsedQuery *nlp.ParsedQuery, userID string) {
	if userID != "" {
		// Authenticated user - can see public + their private content
		parsedQuery.Filters["visibility_filter"] = map[string]interface{}{
			"user_id": userID,
		}
	} else {
		// Guest user - only public content
		parsedQuery.Filters["visibility"] = "public"
	}
}

// buildElasticsearchQuery converts parsed query to Elasticsearch query
func (s *Service) buildElasticsearchQuery(parsedQuery *nlp.ParsedQuery, limit, offset int) map[string]interface{} {
	query := elasticsearch.BuildQuery(parsedQuery.SearchText, parsedQuery.Filters, limit, offset)

	// Add location-based search if present
	if parsedQuery.Location != nil {
		if parsedQuery.Location.Latitude != 0 && parsedQuery.Location.Longitude != 0 {
			// Use exact coordinates
			parsedQuery.Filters["location"] = map[string]interface{}{
				"lat":    parsedQuery.Location.Latitude,
				"lng":    parsedQuery.Location.Longitude,
				"radius": parsedQuery.Location.Radius,
			}
		} else if parsedQuery.Location.Name != "" {
			// Add location name to search text
			if parsedQuery.SearchText == "" {
				parsedQuery.SearchText = parsedQuery.Location.Name
			} else {
				parsedQuery.SearchText += " " + parsedQuery.Location.Name
			}
		}

		// Rebuild query with location
		query = elasticsearch.BuildQuery(parsedQuery.SearchText, parsedQuery.Filters, limit, offset)
	}

	// Add enhanced spatial search if present
	if parsedQuery.Spatial != nil {
		s.addSpatialFilters(query, parsedQuery.Spatial)
	}

	// Add visibility filters
	if visibilityFilter, ok := parsedQuery.Filters["visibility_filter"].(map[string]interface{}); ok {
		userID := visibilityFilter["user_id"].(string)
		
		// Build complex visibility query: public OR (private AND owned by user)
		if boolQuery, ok := query["query"].(map[string]interface{})["bool"].(map[string]interface{}); ok {
			if filters, ok := boolQuery["filter"].([]map[string]interface{}); ok {
				visibilityClause := map[string]interface{}{
					"bool": map[string]interface{}{
						"should": []map[string]interface{}{
							{
								"term": map[string]interface{}{
									"visibility": "public",
								},
							},
							{
								"bool": map[string]interface{}{
									"must": []map[string]interface{}{
										{
											"term": map[string]interface{}{
												"visibility": "private",
											},
										},
										{
											"term": map[string]interface{}{
												"owner_id": userID,
											},
										},
									},
								},
							},
						},
						"minimum_should_match": 1,
					},
				}
				
				boolQuery["filter"] = append(filters, visibilityClause)
			}
		}
	}

	return query
}

// fallbackSearch provides database-based search when Elasticsearch is unavailable
func (s *Service) fallbackSearch(ctx context.Context, parsedQuery *nlp.ParsedQuery, req *SearchRequest) *elasticsearch.SearchResponse {
	// This would integrate with the existing database search
	// For now, return empty results
	log.Printf("Using fallback search for query: %s", req.Query)
	
	return &elasticsearch.SearchResponse{
		Total:   0,
		Results: []elasticsearch.SearchResult{},
		Took:    10, // Placeholder timing
	}
}

// generateSuggestions creates search suggestions based on the query and results
func (s *Service) generateSuggestions(parsedQuery *nlp.ParsedQuery, results *elasticsearch.SearchResponse) []string {
	suggestions := []string{}

	// If no results, suggest similar queries
	if results.Total == 0 {
		switch parsedQuery.Intent {
		case nlp.IntentActivity:
			suggestions = append(suggestions, 
				"Try broadening your search area",
				"Consider different difficulty levels",
				"Look for similar activity types",
			)
		case nlp.IntentPlace:
			suggestions = append(suggestions,
				"Try searching in nearby cities",
				"Look for similar types of places",
				"Expand your search radius",
			)
		default:
			suggestions = append(suggestions,
				"Try more specific keywords",
				"Include location information",
				"Use activity or place names",
			)
		}
	} else if results.Total < 5 {
		// Few results - suggest expanding search
		suggestions = append(suggestions,
			"Expand search area for more results",
			"Try different keywords",
		)
	}

	// Add intent-specific suggestions based on filters
	if parsedQuery.Intent == nlp.IntentActivity {
		if _, hasActivity := parsedQuery.Filters["activity_types"]; !hasActivity {
			suggestions = append(suggestions, "Try specifying an activity type (hiking, biking, etc.)")
		}
		if _, hasDifficulty := parsedQuery.Filters["difficulty_levels"]; !hasDifficulty {
			suggestions = append(suggestions, "Specify difficulty level (easy, moderate, hard)")
		}
	}

	return suggestions
}

// logSearch logs the search query for analytics
func (s *Service) logSearch(ctx context.Context, req *SearchRequest, parsedQuery *nlp.ParsedQuery, results *elasticsearch.SearchResponse) {
	if !s.esClient.IsAvailable() {
		return
	}

	queryLog := map[string]interface{}{
		"query":            req.Query,
		"interpreted_type": string(parsedQuery.Intent),
		"filters":          parsedQuery.Filters,
		"results_count":    results.Total,
		"user_id":          req.UserID,
		"session_id":       req.SessionID,
		"confidence":       parsedQuery.Confidence,
		"timestamp":        time.Now(),
		"took_ms":          results.Took,
	}

	if err := s.esClient.LogQuery(ctx, queryLog); err != nil {
		log.Printf("Failed to log search query: %v", err)
	}
}

// GetSearchSuggestions provides autocomplete suggestions
func (s *Service) GetSearchSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	// This would implement autocomplete functionality
	// For now, return some static suggestions based on common patterns
	
	commonSuggestions := []string{
		"hiking trails near me",
		"easy bike routes",
		"waterfall hikes",
		"weekend camping spots",
		"moderate difficulty trails",
		"mountain climbing routes",
		"family-friendly activities",
		"dog-friendly hikes",
		"scenic bike paths",
		"swimming holes",
	}

	var suggestions []string
	for _, suggestion := range commonSuggestions {
		if len(suggestions) >= limit {
			break
		}
		if prefix == "" || containsIgnoreCase(suggestion, prefix) {
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}

// containsIgnoreCase checks if a string contains another string (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (substr == "" || 
		    len(s) > 0 && 
		    (s[0:1] == substr[0:1] || 
		     (s[0] >= 'A' && s[0] <= 'Z' && s[0]+32 == substr[0]) ||
		     (s[0] >= 'a' && s[0] <= 'z' && s[0]-32 == substr[0])))
}

// IndexActivity indexes an activity for search
func (s *Service) IndexActivity(ctx context.Context, activityID string, activity map[string]interface{}) error {
	if !s.esClient.IsAvailable() {
		log.Printf("Elasticsearch not available, skipping activity indexing: %s", activityID)
		return nil
	}

	return s.esClient.IndexActivity(ctx, activityID, activity)
}

// IndexPlace indexes a place for search
func (s *Service) IndexPlace(ctx context.Context, placeID string, place map[string]interface{}) error {
	if !s.esClient.IsAvailable() {
		log.Printf("Elasticsearch not available, skipping place indexing: %s", placeID)
		return nil
	}

	return s.esClient.IndexPlace(ctx, placeID, place)
}

// DeleteFromIndex removes a document from the search index
func (s *Service) DeleteFromIndex(ctx context.Context, docType, documentID string) error {
	if !s.esClient.IsAvailable() {
		return nil
	}

	index := "activities"
	if docType == "place" {
		index = "places"
	}

	return s.esClient.DeleteDocument(ctx, index, documentID)
}
// addSpatialFilters adds enhanced spatial search filters to Elasticsearch query
func (s *Service) addSpatialFilters(query map[string]interface{}, spatial *nlp.SpatialSearchContext) {
	if spatial == nil {
		return
	}

	// Get the bool query to add filters to
	boolQuery, ok := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
	if !ok {
		// Create bool query if it doesn't exist
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []interface{}{},
				"filter": []interface{}{},
			},
		}
		boolQuery = query["query"].(map[string]interface{})["bool"].(map[string]interface{})
	}

	// Ensure filter array exists
	if _, ok := boolQuery["filter"]; !ok {
		boolQuery["filter"] = []interface{}{}
	}
	filters := boolQuery["filter"].([]interface{})

	// Within area - document must be completely within the specified area
	if spatial.Within != nil {
		spatialFilter := s.buildSpatialFilter("within", spatial.Within)
		if spatialFilter != nil {
			filters = append(filters, spatialFilter)
		}
	}

	// Near area - document must be within distance of the specified area
	if spatial.Near != nil {
		spatialFilter := s.buildSpatialFilter("distance", spatial.Near)
		if spatialFilter != nil {
			filters = append(filters, spatialFilter)
		}
	}

	// Intersects area - document must intersect with the specified area
	if spatial.Intersects != nil {
		spatialFilter := s.buildSpatialFilter("intersects", spatial.Intersects)
		if spatialFilter != nil {
			filters = append(filters, spatialFilter)
		}
	}

	// Multiple areas - document must be in one of the specified areas
	if len(spatial.Areas) > 0 {
		areaFilters := []interface{}{}
		for _, area := range spatial.Areas {
			spatialFilter := s.buildSpatialFilter("within", &area)
			if spatialFilter != nil {
				areaFilters = append(areaFilters, spatialFilter)
			}
		}
		
		if len(areaFilters) > 0 {
			// Use "should" query for OR logic between areas
			filters = append(filters, map[string]interface{}{
				"bool": map[string]interface{}{
					"should": areaFilters,
					"minimum_should_match": 1,
				},
			})
		}
	}

	// Update the filters array
	boolQuery["filter"] = filters
}

// buildSpatialFilter creates Elasticsearch spatial filter for a specific area
func (s *Service) buildSpatialFilter(operation string, area *nlp.AreaFilter) map[string]interface{} {
	if area == nil {
		return nil
	}

	switch area.Type {
	case "circle":
		// Circular area search using geo_distance
		if coords, ok := area.Coordinates.([]interface{}); ok && len(coords) >= 2 && area.Radius != nil {
			lat, latOk := coords[1].(float64)
			lng, lngOk := coords[0].(float64)
			if latOk && lngOk {
				return map[string]interface{}{
					"geo_distance": map[string]interface{}{
						"distance": fmt.Sprintf("%.0fkm", *area.Radius),
						"location": map[string]float64{
							"lat": lat,
							"lon": lng,
						},
					},
				}
			}
		}

	case "polygon":
		// Polygon area search using geo_polygon
		if coords, ok := area.Coordinates.([]interface{}); ok {
			// Convert coordinates to Elasticsearch format
			points := []map[string]float64{}
			for _, coord := range coords {
				if coordPair, ok := coord.([]interface{}); ok && len(coordPair) >= 2 {
					if lng, lngOk := coordPair[0].(float64); lngOk {
						if lat, latOk := coordPair[1].(float64); latOk {
							points = append(points, map[string]float64{
								"lat": lat,
								"lon": lng,
							})
						}
					}
				}
			}

			if len(points) >= 3 { // Minimum points for a polygon
				return map[string]interface{}{
					"geo_polygon": map[string]interface{}{
						"location": map[string]interface{}{
							"points": points,
						},
					},
				}
			}
		}

	case "bounds":
		// Rectangular bounds search using geo_bounding_box
		if coords, ok := area.Coordinates.([]interface{}); ok && len(coords) >= 4 {
			minLng, _ := coords[0].(float64)
			minLat, _ := coords[1].(float64)
			maxLng, _ := coords[2].(float64)
			maxLat, _ := coords[3].(float64)

			return map[string]interface{}{
				"geo_bounding_box": map[string]interface{}{
					"location": map[string]interface{}{
						"top_left": map[string]float64{
							"lat": maxLat,
							"lon": minLng,
						},
						"bottom_right": map[string]float64{
							"lat": minLat,
							"lon": maxLng,
						},
					},
				},
			}
		}

	case "region":
		// Named region search - add to text search
		if area.Name != "" {
			return map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  area.Name,
					"fields": []string{"city", "state", "country", "region", "location_name"},
					"type":   "best_fields",
					"boost":  0.5, // Lower boost than main content
				},
			}
		}
	}

	return nil
}
