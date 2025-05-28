// @ts-ignore
import mapboxgl from 'mapbox-gl';
import { Trip } from '../../types';

interface TripRouteOptions {
  trip: Trip;
  map: mapboxgl.Map;
  onClick?: () => void;
}

export class TripRoute {
  private map: mapboxgl.Map;
  private sourceId: string;
  private layerId: string;

  constructor(options: TripRouteOptions) {
    const { trip, map, onClick } = options;
    
    this.map = map;
    this.sourceId = `trip-route-${trip.id}`;
    this.layerId = `trip-route-layer-${trip.id}`;

    // Extract coordinates from waypoints
    const coordinates = trip.waypoints
      .filter(w => w.place?.coordinates)
      .map(w => [w.place.coordinates.lng, w.place.coordinates.lat]);

    if (coordinates.length < 2) return;

    // Add source
    this.map.addSource(this.sourceId, {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'LineString',
          coordinates,
        },
      },
    });

    // Add layer
    this.map.addLayer({
      id: this.layerId,
      type: 'line',
      source: this.sourceId,
      layout: {
        'line-join': 'round',
        'line-cap': 'round',
      },
      paint: {
        'line-color': '#6366f1',
        'line-width': 3,
        'line-opacity': 0.8,
      },
    });

    // Add click handler
    if (onClick) {
      this.map.on('click', this.layerId, onClick);
      this.map.on('mouseenter', this.layerId, () => {
        this.map.getCanvas().style.cursor = 'pointer';
      });
      this.map.on('mouseleave', this.layerId, () => {
        this.map.getCanvas().style.cursor = '';
      });
    }
  }

  remove() {
    if (this.map.getLayer(this.layerId)) {
      this.map.removeLayer(this.layerId);
    }
    if (this.map.getSource(this.sourceId)) {
      this.map.removeSource(this.sourceId);
    }
  }
}