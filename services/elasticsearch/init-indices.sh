#!/bin/bash

# Wait for Elasticsearch to be ready
until curl -s http://localhost:9200/_cluster/health | grep -q '"status":"green"\|"status":"yellow"'; do
  echo "Waiting for Elasticsearch to be ready..."
  sleep 5
done

echo "Elasticsearch is ready. Creating indices..."

# Create activities index with mapping
curl -X PUT "http://localhost:9200/activities" -H 'Content-Type: application/json' -d '
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "activity_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "asciifolding", "stop", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "title": { 
        "type": "text",
        "analyzer": "activity_analyzer",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { 
        "type": "text",
        "analyzer": "activity_analyzer"
      },
      "activity_type": { "type": "keyword" },
      "difficulty_level": { "type": "keyword" },
      "duration_hours": { "type": "float" },
      "distance_km": { "type": "float" },
      "elevation_gain_m": { "type": "integer" },
      "location": { "type": "geo_point" },
      "route": { "type": "geo_shape" },
      "water_features": { "type": "keyword" },
      "terrain_types": { "type": "keyword" },
      "best_seasons": { "type": "keyword" },
      "visibility": { "type": "keyword" },
      "owner_id": { "type": "keyword" },
      "tags": { "type": "keyword" },
      "average_rating": { "type": "float" },
      "completion_count": { "type": "integer" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}'

# Create places index with mapping
curl -X PUT "http://localhost:9200/places" -H 'Content-Type: application/json' -d '
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "place_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "asciifolding", "stop", "snowball"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "name": { 
        "type": "text",
        "analyzer": "place_analyzer",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { 
        "type": "text",
        "analyzer": "place_analyzer"
      },
      "type": { "type": "keyword" },
      "location": { "type": "geo_point" },
      "category": { "type": "keyword" },
      "tags": { "type": "keyword" },
      "city": { "type": "keyword" },
      "state": { "type": "keyword" },
      "country": { "type": "keyword" },
      "average_rating": { "type": "float" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}'

# Create search_queries index for analytics
curl -X PUT "http://localhost:9200/search_queries" -H 'Content-Type: application/json' -d '
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "query": { "type": "text" },
      "interpreted_type": { "type": "keyword" },
      "filters": { "type": "object" },
      "results_count": { "type": "integer" },
      "user_id": { "type": "keyword" },
      "timestamp": { "type": "date" }
    }
  }
}'

echo "Indices created successfully!"