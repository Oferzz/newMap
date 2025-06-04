package nlp

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// QueryIntent represents the type of search query
type QueryIntent string

const (
	IntentActivity QueryIntent = "activity"
	IntentPlace    QueryIntent = "place"
	IntentMixed    QueryIntent = "mixed"
	IntentUnknown  QueryIntent = "unknown"
)

// ParsedQuery represents the structured output from natural language parsing
type ParsedQuery struct {
	Intent      QueryIntent            `json:"intent"`
	SearchText  string                 `json:"search_text"`
	Filters     map[string]interface{} `json:"filters"`
	Location    *LocationFilter        `json:"location,omitempty"`
	Spatial     *SpatialSearchContext  `json:"spatial,omitempty"`
	Confidence  float64                `json:"confidence"`
	Keywords    []string               `json:"keywords"`
	Explanation string                 `json:"explanation"`
}

// LocationFilter represents location-based search parameters
type LocationFilter struct {
	Name      string  `json:"name,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Radius    float64 `json:"radius,omitempty"` // in kilometers
}

// AreaFilter represents geometric area-based search parameters
type AreaFilter struct {
	Type        string      `json:"type"`        // "circle", "polygon", "bounds", "region"
	Coordinates interface{} `json:"coordinates"` // Format depends on type
	Radius      *float64    `json:"radius,omitempty"` // for circles
	Name        string      `json:"name,omitempty"`   // human-readable name
}

// SpatialSearchContext represents enhanced spatial search parameters
type SpatialSearchContext struct {
	Areas      []AreaFilter `json:"areas,omitempty"`      // Areas to search within
	Within     *AreaFilter  `json:"within,omitempty"`     // Must be completely within
	Intersects *AreaFilter  `json:"intersects,omitempty"` // Must intersect with
	Near       *AreaFilter  `json:"near,omitempty"`       // Within distance of
}

// Parser handles natural language query parsing
type Parser struct {
	// In a real implementation, this would contain LLM client
	// For now, we'll use rule-based parsing as a fallback
}

// NewParser creates a new NLP parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseQuery parses a natural language query into structured filters
func (p *Parser) ParseQuery(ctx context.Context, query string) (*ParsedQuery, error) {
	// Clean and normalize the query
	cleanQuery := strings.TrimSpace(strings.ToLower(query))
	if cleanQuery == "" {
		return &ParsedQuery{
			Intent:      IntentUnknown,
			SearchText:  query,
			Filters:     make(map[string]interface{}),
			Confidence:  0.0,
			Keywords:    []string{},
			Explanation: "Empty query provided",
		}, nil
	}

	// Try LLM parsing first (placeholder for now)
	// In production, this would call OpenAI/Anthropic API
	parsed, err := p.parseWithLLM(ctx, cleanQuery)
	if err != nil {
		// Fallback to rule-based parsing
		parsed = p.parseWithRules(cleanQuery)
	}

	// Enhance with keyword extraction
	parsed.Keywords = p.extractKeywords(cleanQuery)
	
	return parsed, nil
}

// parseWithLLM would call an LLM API to parse the query
// This is a placeholder implementation
func (p *Parser) parseWithLLM(ctx context.Context, query string) (*ParsedQuery, error) {
	// TODO: Implement actual LLM integration
	// For now, return an error to force fallback to rule-based parsing
	return nil, fmt.Errorf("LLM parsing not implemented yet")
}

// parseWithRules uses rule-based parsing as a fallback
func (p *Parser) parseWithRules(query string) *ParsedQuery {
	parsed := &ParsedQuery{
		Intent:      IntentUnknown,
		SearchText:  query,
		Filters:     make(map[string]interface{}),
		Confidence:  0.6, // Lower confidence for rule-based
		Explanation: "Parsed using rule-based system",
	}

	// Determine intent based on keywords
	activityKeywords := []string{
		"hiking", "biking", "climbing", "trail", "hike", "bike", "climb",
		"skiing", "snowboarding", "kayaking", "swimming", "running",
		"backpacking", "camping", "fishing", "activity", "activities",
		"route", "routes", "easy", "moderate", "hard", "difficult",
		"weekend", "day trip", "overnight",
	}

	placeKeywords := []string{
		"restaurant", "hotel", "coffee", "shop", "store", "museum",
		"park", "beach", "lake", "mountain", "city", "town",
		"attraction", "landmark", "place", "places", "spot", "spots",
		"near", "in", "around",
	}

	// Count keyword matches
	activityMatches := p.countKeywordMatches(query, activityKeywords)
	placeMatches := p.countKeywordMatches(query, placeKeywords)

	if activityMatches > placeMatches {
		parsed.Intent = IntentActivity
		parsed.Confidence += 0.2
	} else if placeMatches > activityMatches {
		parsed.Intent = IntentPlace
		parsed.Confidence += 0.2
	} else if activityMatches > 0 && placeMatches > 0 {
		parsed.Intent = IntentMixed
	}

	// Parse activity-specific filters
	if parsed.Intent == IntentActivity || parsed.Intent == IntentMixed {
		p.parseActivityFilters(query, parsed)
	}

	// Parse location information
	location := p.parseLocation(query)
	if location != nil {
		parsed.Location = location
		parsed.Confidence += 0.1
	}

	// Parse spatial/area information
	spatial := p.parseSpatialContext(query)
	if spatial != nil {
		parsed.Spatial = spatial
		parsed.Confidence += 0.15
	}

	// Parse duration and distance
	p.parseDurationAndDistance(query, parsed)

	return parsed
}

// countKeywordMatches counts how many keywords from the list appear in the query
func (p *Parser) countKeywordMatches(query string, keywords []string) int {
	count := 0
	for _, keyword := range keywords {
		if strings.Contains(query, keyword) {
			count++
		}
	}
	return count
}

// parseActivityFilters extracts activity-specific filters
func (p *Parser) parseActivityFilters(query string, parsed *ParsedQuery) {
	// Activity types
	activityTypes := map[string]string{
		"hiking":      "hiking",
		"hike":        "hiking",
		"trail":       "hiking",
		"walk":        "walking",
		"biking":      "biking",
		"bike":        "biking",
		"cycling":     "biking",
		"climbing":    "climbing",
		"climb":       "climbing",
		"skiing":      "skiing",
		"ski":         "skiing",
		"snowboard":   "snowboarding",
		"kayaking":    "kayaking",
		"kayak":       "kayaking",
		"swimming":    "swimming",
		"swim":        "swimming",
		"running":     "running",
		"run":         "running",
		"backpacking": "backpacking",
		"camping":     "camping",
		"camp":        "camping",
		"fishing":     "fishing",
		"fish":        "fishing",
	}

	for keyword, activityType := range activityTypes {
		if strings.Contains(query, keyword) {
			if parsed.Filters["activity_types"] == nil {
				parsed.Filters["activity_types"] = []string{}
			}
			types := parsed.Filters["activity_types"].([]string)
			// Check if already added
			found := false
			for _, t := range types {
				if t == activityType {
					found = true
					break
				}
			}
			if !found {
				parsed.Filters["activity_types"] = append(types, activityType)
			}
		}
	}

	// Difficulty levels
	difficultyMap := map[string]string{
		"easy":        "easy",
		"beginner":    "easy",
		"simple":      "easy",
		"moderate":    "moderate",
		"medium":      "moderate",
		"intermediate": "moderate",
		"hard":        "hard",
		"difficult":   "hard",
		"challenging": "hard",
		"expert":      "expert",
		"advanced":    "expert",
	}

	for keyword, difficulty := range difficultyMap {
		if strings.Contains(query, keyword) {
			if parsed.Filters["difficulty_levels"] == nil {
				parsed.Filters["difficulty_levels"] = []string{}
			}
			levels := parsed.Filters["difficulty_levels"].([]string)
			// Check if already added
			found := false
			for _, l := range levels {
				if l == difficulty {
					found = true
					break
				}
			}
			if !found {
				parsed.Filters["difficulty_levels"] = append(levels, difficulty)
			}
		}
	}

	// Water features
	waterFeatures := []string{
		"waterfall", "waterfalls", "river", "rivers", "lake", "lakes",
		"pond", "ponds", "creek", "creeks", "stream", "streams",
		"swimming", "swim", "water",
	}

	for _, feature := range waterFeatures {
		if strings.Contains(query, feature) {
			if parsed.Filters["water_features"] == nil {
				parsed.Filters["water_features"] = []string{}
			}
			features := parsed.Filters["water_features"].([]string)
			// Add generic water feature
			found := false
			for _, f := range features {
				if f == "water" {
					found = true
					break
				}
			}
			if !found {
				parsed.Filters["water_features"] = append(features, "water")
			}
			break
		}
	}
}

// parseLocation extracts location information from the query
func (p *Parser) parseLocation(query string) *LocationFilter {
	// Look for location patterns
	locationPatterns := []string{
		`near\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
		`in\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
		`around\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
		`([a-zA-Z\s]+)\s+area`,
	}

	for _, pattern := range locationPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 1 {
			locationName := strings.TrimSpace(matches[1])
			if len(locationName) > 2 { // Avoid single letters
				return &LocationFilter{
					Name:   locationName,
					Radius: 50, // Default 50km radius
				}
			}
		}
	}

	return nil
}

// parseSpatialContext extracts area-based spatial search parameters
func (p *Parser) parseSpatialContext(query string) *SpatialSearchContext {
	spatial := &SpatialSearchContext{}
	hasFilters := false

	// Within patterns - activities/places that must be completely inside an area
	withinPatterns := []string{
		`within\s+([a-zA-Z\s]+?)(?:\s+area|region|bounds)?(?:\s|$|,)`,
		`inside\s+([a-zA-Z\s]+?)(?:\s+area|region|bounds)?(?:\s|$|,)`,
		`in\s+the\s+([a-zA-Z\s]+?)\s+(?:area|region|zone|district)(?:\s|$|,)`,
		`([a-zA-Z\s]+?)\s+city\s+limits`,
		`([a-zA-Z\s]+?)\s+county(?:\s|$|,)`,
		`([a-zA-Z\s]+?)\s+state\s+park(?:\s|$|,)`,
		`([a-zA-Z\s]+?)\s+national\s+park(?:\s|$|,)`,
	}

	for _, pattern := range withinPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 1 {
			areaName := strings.TrimSpace(matches[1])
			if len(areaName) > 2 {
				spatial.Within = &AreaFilter{
					Type: "region",
					Name: areaName,
				}
				hasFilters = true
				break
			}
		}
	}

	// Distance-based area patterns - creates circular areas
	distancePatterns := []string{
		`within\s+(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)\s+of\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
		`(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)\s+(?:from|of|around)\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
		`near\s+([a-zA-Z\s]+?)\s+within\s+(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)(?:\s|$|,)`,
		`around\s+([a-zA-Z\s]+?)\s+(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)(?:\s|$|,)`,
	}

	for _, pattern := range distancePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) >= 3 {
			var distance float64
			var locationName string
			var err error

			// Handle different pattern groups
			if strings.Contains(pattern, "within.*of") {
				// Pattern: "within X miles of location"
				distance, err = strconv.ParseFloat(matches[1], 64)
				locationName = strings.TrimSpace(matches[2])
			} else if strings.Contains(pattern, "near.*within") {
				// Pattern: "near location within X miles"
				locationName = strings.TrimSpace(matches[1])
				distance, err = strconv.ParseFloat(matches[2], 64)
			} else {
				// Pattern: "X miles from location" or "around location X miles"
				if strings.Contains(pattern, "around") {
					locationName = strings.TrimSpace(matches[1])
					distance, err = strconv.ParseFloat(matches[2], 64)
				} else {
					distance, err = strconv.ParseFloat(matches[1], 64)
					locationName = strings.TrimSpace(matches[2])
				}
			}

			if err == nil && len(locationName) > 2 {
				// Convert miles to km if needed
				if strings.Contains(matches[0], "mile") || strings.Contains(matches[0], "mi") {
					distance = distance * 1.60934
				}

				spatial.Near = &AreaFilter{
					Type:   "circle",
					Name:   locationName,
					Radius: &distance,
				}
				hasFilters = true
				break
			}
		}
	}

	// Geographic region patterns
	regionPatterns := []string{
		`in\s+(?:the\s+)?([a-zA-Z\s]+?)\s+(?:mountains?|hills?|valleys?)(?:\s|$|,)`,
		`(?:along|near)\s+(?:the\s+)?([a-zA-Z\s]+?)\s+(?:coast|coastline|shore|shoreline)(?:\s|$|,)`,
		`in\s+(?:the\s+)?([a-zA-Z\s]+?)\s+(?:desert|wilderness|forest)(?:\s|$|,)`,
		`(?:around|near)\s+(?:the\s+)?([a-zA-Z\s]+?)\s+(?:river|lake|bay|peninsula)(?:\s|$|,)`,
		`in\s+(?:the\s+)?([a-zA-Z\s]+?)\s+(?:area|region|zone)(?:\s|$|,)`,
		`(?:north|south|east|west|northern|southern|eastern|western)\s+([a-zA-Z\s]+?)(?:\s|$|,)`,
	}

	for _, pattern := range regionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 1 {
			regionName := strings.TrimSpace(matches[1])
			if len(regionName) > 2 {
				area := AreaFilter{
					Type: "region",
					Name: regionName,
				}
				
				// Determine if this should be "within" or just an area filter
				if strings.Contains(pattern, "mountains") || strings.Contains(pattern, "coast") || 
				   strings.Contains(pattern, "desert") || strings.Contains(pattern, "forest") {
					spatial.Within = &area
				} else {
					spatial.Areas = append(spatial.Areas, area)
				}
				hasFilters = true
				break
			}
		}
	}

	// Elevation-based areas
	elevationPatterns := []string{
		`(?:above|over)\s+(\d+)\s*(?:feet|ft|meters?|m)\s*(?:elevation)?`,
		`(?:below|under)\s+(\d+)\s*(?:feet|ft|meters?|m)\s*(?:elevation)?`,
		`(?:at|around)\s+(\d+)\s*(?:feet|ft|meters?|m)\s*(?:elevation)?`,
		`(?:high|low)\s+elevation`,
		`(?:sea\s+level|low\s+altitude)`,
	}

	for _, pattern := range elevationPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 0 {
			// Add elevation filter to main filters rather than spatial context
			// This will be handled in the repository layer
			if len(matches) > 1 {
				if elevation, err := strconv.ParseFloat(matches[1], 64); err == nil {
					// Convert feet to meters if needed
					if strings.Contains(matches[0], "feet") || strings.Contains(matches[0], "ft") {
						elevation = elevation * 0.3048
					}
					
					if strings.Contains(pattern, "above") || strings.Contains(pattern, "over") {
						// This would be added to parsed.Filters in the calling function
					} else if strings.Contains(pattern, "below") || strings.Contains(pattern, "under") {
						// This would be added to parsed.Filters in the calling function
					}
				}
			}
			// Add semantic elevation filters
			if strings.Contains(matches[0], "high elevation") {
				// High elevation areas
			} else if strings.Contains(matches[0], "sea level") || strings.Contains(matches[0], "low altitude") {
				// Low elevation areas
			}
			hasFilters = true
			break
		}
	}

	if !hasFilters {
		return nil
	}

	return spatial
}

// parseDurationAndDistance extracts duration and distance information
func (p *Parser) parseDurationAndDistance(query string, parsed *ParsedQuery) {
	// Duration patterns
	durationPatterns := []string{
		`(\d+(?:\.\d+)?)\s*hours?`,
		`(\d+(?:\.\d+)?)\s*hrs?`,
		`(\d+(?:\.\d+)?)\s*h`,
		`(\d+)\s*day\s*trip`,
		`(\d+)\s*days?`,
		`weekend`,
		`half\s*day`,
		`full\s*day`,
	}

	for _, pattern := range durationPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 1 {
			if pattern == `weekend` {
				parsed.Filters["max_duration"] = 48.0 // 2 days
			} else if pattern == `half\s*day` {
				parsed.Filters["max_duration"] = 4.0
			} else if pattern == `full\s*day` {
				parsed.Filters["max_duration"] = 8.0
			} else if strings.Contains(pattern, "day") {
				if duration, err := strconv.ParseFloat(matches[1], 64); err == nil {
					parsed.Filters["max_duration"] = duration * 24 // Convert days to hours
				}
			} else {
				if duration, err := strconv.ParseFloat(matches[1], 64); err == nil {
					parsed.Filters["max_duration"] = duration
				}
			}
			break
		}
	}

	// Distance patterns
	distancePatterns := []string{
		`(\d+(?:\.\d+)?)\s*miles?`,
		`(\d+(?:\.\d+)?)\s*mi`,
		`(\d+(?:\.\d+)?)\s*kilometers?`,
		`(\d+(?:\.\d+)?)\s*km`,
		`under\s+(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)`,
		`less\s+than\s+(\d+(?:\.\d+)?)\s*(?:miles?|mi|kilometers?|km)`,
	}

	for _, pattern := range distancePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(query)
		if len(matches) > 1 {
			if distance, err := strconv.ParseFloat(matches[1], 64); err == nil {
				// Convert miles to km if needed
				if strings.Contains(pattern, "mile") || strings.Contains(pattern, "mi") {
					distance = distance * 1.60934
				}
				if strings.Contains(pattern, "under") || strings.Contains(pattern, "less") {
					parsed.Filters["max_distance"] = distance
				} else {
					parsed.Filters["max_distance"] = distance
				}
			}
			break
		}
	}
}

// extractKeywords extracts relevant keywords from the query
func (p *Parser) extractKeywords(query string) []string {
	// Simple keyword extraction - split and filter
	words := strings.Fields(query)
	var keywords []string

	// Filter out common stop words
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "were": true, "will": true,
		"with": true, "me": true, "my": true, "i": true, "you": true, "your": true,
		"we": true, "our": true, "they": true, "their": true, "them": true,
	}

	for _, word := range words {
		cleaned := strings.ToLower(strings.Trim(word, ".,!?;:"))
		if len(cleaned) > 2 && !stopWords[cleaned] {
			keywords = append(keywords, cleaned)
		}
	}

	return keywords
}

// GenerateExplanation creates a human-readable explanation of the parsed query
func (p *Parser) GenerateExplanation(parsed *ParsedQuery) string {
	parts := []string{}

	// Intent
	switch parsed.Intent {
	case IntentActivity:
		parts = append(parts, "Looking for activities")
	case IntentPlace:
		parts = append(parts, "Looking for places")
	case IntentMixed:
		parts = append(parts, "Looking for activities and places")
	default:
		parts = append(parts, "General search")
	}

	// Activity types
	if activityTypes, ok := parsed.Filters["activity_types"].([]string); ok && len(activityTypes) > 0 {
		parts = append(parts, fmt.Sprintf("Activity types: %s", strings.Join(activityTypes, ", ")))
	}

	// Difficulty
	if difficultyLevels, ok := parsed.Filters["difficulty_levels"].([]string); ok && len(difficultyLevels) > 0 {
		parts = append(parts, fmt.Sprintf("Difficulty: %s", strings.Join(difficultyLevels, ", ")))
	}

	// Location
	if parsed.Location != nil && parsed.Location.Name != "" {
		parts = append(parts, fmt.Sprintf("Near %s", parsed.Location.Name))
	}

	// Spatial context
	if parsed.Spatial != nil {
		if parsed.Spatial.Within != nil {
			parts = append(parts, fmt.Sprintf("Within %s", parsed.Spatial.Within.Name))
		}
		if parsed.Spatial.Near != nil {
			if parsed.Spatial.Near.Radius != nil {
				parts = append(parts, fmt.Sprintf("Within %.1f km of %s", *parsed.Spatial.Near.Radius, parsed.Spatial.Near.Name))
			} else {
				parts = append(parts, fmt.Sprintf("Near %s", parsed.Spatial.Near.Name))
			}
		}
		if len(parsed.Spatial.Areas) > 0 {
			areaNames := make([]string, len(parsed.Spatial.Areas))
			for i, area := range parsed.Spatial.Areas {
				areaNames[i] = area.Name
			}
			parts = append(parts, fmt.Sprintf("In areas: %s", strings.Join(areaNames, ", ")))
		}
	}

	// Duration
	if maxDuration, ok := parsed.Filters["max_duration"].(float64); ok {
		if maxDuration < 24 {
			parts = append(parts, fmt.Sprintf("Up to %.1f hours", maxDuration))
		} else {
			parts = append(parts, fmt.Sprintf("Up to %.1f days", maxDuration/24))
		}
	}

	// Distance
	if maxDistance, ok := parsed.Filters["max_distance"].(float64); ok {
		parts = append(parts, fmt.Sprintf("Up to %.1f km", maxDistance))
	}

	if len(parts) == 0 {
		return "General search"
	}

	return strings.Join(parts, " • ")
}