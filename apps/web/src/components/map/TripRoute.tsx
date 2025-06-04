// @ts-ignore
import mapboxgl from 'mapbox-gl';
import { Trip } from '../../types';

interface TripRouteOptions {
  trip: Trip;
  map: mapboxgl.Map;
  onClick?: () => void;
  activityType?: string;
  difficulty?: string;
  isSelected?: boolean;
}

export class TripRoute {
  private map: mapboxgl.Map;
  private sourceId: string;
  private layerId: string;
  private routeAreaLayerId?: string;

  constructor(options: TripRouteOptions) {
    const { trip, map, onClick, activityType, difficulty, isSelected } = options;
    
    this.map = map;
    this.sourceId = `trip-route-${trip.id}`;
    this.layerId = `trip-route-layer-${trip.id}`;

    // Extract coordinates from waypoints
    const coordinates = (trip.waypoints || [])
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

    // Get styling based on activity type and difficulty
    const routeStyle = this.getRouteStyle(activityType, difficulty, isSelected);

    // Add route area/buffer if applicable
    if (routeStyle.showArea && coordinates.length >= 2) {
      this.addRouteArea(coordinates, routeStyle);
    }

    // Add main route line
    this.map.addLayer({
      id: this.layerId,
      type: 'line',
      source: this.sourceId,
      layout: {
        'line-join': 'round',
        'line-cap': 'round',
      },
      paint: {
        'line-color': routeStyle.color,
        'line-width': routeStyle.width,
        'line-opacity': routeStyle.opacity,
        'line-dasharray': routeStyle.dashArray,
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

  private getRouteStyle(activityType?: string, difficulty?: string, isSelected?: boolean) {
    const baseStyle = {
      width: isSelected ? 5 : 3,
      opacity: isSelected ? 1.0 : 0.8,
      dashArray: undefined as number[] | undefined,
      showArea: false,
      areaColor: '#6366f1',
      areaOpacity: 0.1,
    };

    // Activity type specific colors and styles
    let color = '#6366f1'; // Default blue
    
    switch (activityType?.toLowerCase()) {
      case 'hiking':
      case 'backpacking':
        color = '#059669'; // Green
        baseStyle.showArea = true;
        baseStyle.areaColor = '#059669';
        break;
      case 'biking':
      case 'mountain-biking':
      case 'road-cycling':
        color = '#dc2626'; // Red
        baseStyle.width += 1;
        break;
      case 'trail-running':
        color = '#ea580c'; // Orange
        baseStyle.dashArray = [2, 1];
        break;
      case 'climbing':
      case 'rock-climbing':
        color = '#7c2d12'; // Brown
        baseStyle.dashArray = [3, 2];
        break;
      case 'skiing':
      case 'snowboarding':
        color = '#1e40af'; // Dark blue
        baseStyle.dashArray = [4, 2];
        break;
      case 'kayaking':
      case 'canoeing':
        color = '#0891b2'; // Cyan
        baseStyle.showArea = true;
        baseStyle.areaColor = '#0891b2';
        break;
      case 'fishing':
        color = '#0d9488'; // Teal
        baseStyle.dashArray = [1, 1];
        break;
      case 'camping':
        color = '#65a30d'; // Lime
        break;
      case 'photography':
      case 'wildlife-viewing':
        color = '#7c3aed'; // Purple
        baseStyle.dashArray = [5, 3];
        break;
      default:
        color = '#6366f1'; // Default blue
    }

    // Difficulty modifications
    switch (difficulty?.toLowerCase()) {
      case 'easy':
        baseStyle.opacity *= 0.9;
        break;
      case 'moderate':
        // No change
        break;
      case 'hard':
        baseStyle.width += 1;
        break;
      case 'expert':
        baseStyle.width += 2;
        color = this.darkenColor(color, 0.2);
        break;
    }

    return {
      color,
      ...baseStyle,
    };
  }

  private addRouteArea(coordinates: number[][], style: any) {
    // Create a buffer around the route for area-based activities
    const buffered = this.createRouteBuffer(coordinates, 0.0005); // ~50m buffer
    
    this.routeAreaLayerId = `${this.layerId}-area`;
    const areaSourceId = `${this.sourceId}-area`;

    this.map.addSource(areaSourceId, {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Polygon',
          coordinates: [buffered],
        },
      },
    });

    this.map.addLayer({
      id: this.routeAreaLayerId,
      type: 'fill',
      source: areaSourceId,
      paint: {
        'fill-color': style.areaColor,
        'fill-opacity': style.areaOpacity,
      },
    });
  }

  private createRouteBuffer(coordinates: number[][], buffer: number): number[][] {
    // Simple buffer creation - in production, use turf.js for proper buffering
    const buffered: number[][] = [];
    
    // Add points around each coordinate
    coordinates.forEach(coord => {
      const [lng, lat] = coord;
      // Create a simple square buffer around each point
      buffered.push(
        [lng - buffer, lat - buffer],
        [lng + buffer, lat - buffer],
        [lng + buffer, lat + buffer],
        [lng - buffer, lat + buffer]
      );
    });
    
    // Close the polygon
    if (buffered.length > 0) {
      buffered.push(buffered[0]);
    }
    
    return buffered;
  }

  private darkenColor(color: string, amount: number): string {
    // Simple color darkening - convert hex to RGB, darken, convert back
    const hex = color.replace('#', '');
    const r = Math.max(0, parseInt(hex.substr(0, 2), 16) * (1 - amount));
    const g = Math.max(0, parseInt(hex.substr(2, 2), 16) * (1 - amount));
    const b = Math.max(0, parseInt(hex.substr(4, 2), 16) * (1 - amount));
    
    return `#${Math.round(r).toString(16).padStart(2, '0')}${Math.round(g).toString(16).padStart(2, '0')}${Math.round(b).toString(16).padStart(2, '0')}`;
  }

  remove() {
    if (this.map.getLayer(this.layerId)) {
      this.map.removeLayer(this.layerId);
    }
    if (this.map.getSource(this.sourceId)) {
      this.map.removeSource(this.sourceId);
    }
    
    // Remove area layer if it exists
    if (this.routeAreaLayerId && this.map.getLayer(this.routeAreaLayerId)) {
      this.map.removeLayer(this.routeAreaLayerId);
    }
    const areaSourceId = `${this.sourceId}-area`;
    if (this.map.getSource(areaSourceId)) {
      this.map.removeSource(areaSourceId);
    }
  }
}