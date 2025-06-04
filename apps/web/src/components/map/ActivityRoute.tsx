// @ts-ignore
import mapboxgl from 'mapbox-gl';
import { Activity } from '../../types/activity.types';

interface ActivityRouteOptions {
  activity: Activity;
  map: mapboxgl.Map;
  onClick?: () => void;
  isSelected?: boolean;
  isHovered?: boolean;
}

export class ActivityRoute {
  private map: mapboxgl.Map;
  private sourceId: string;
  private layerId: string;
  private routeAreaLayerId?: string;
  private startMarkerId?: string;
  private endMarkerId?: string;

  constructor(options: ActivityRouteOptions) {
    const { activity, map, onClick, isSelected, isHovered } = options;
    
    this.map = map;
    this.sourceId = `activity-route-${activity.id}`;
    this.layerId = `activity-route-layer-${activity.id}`;

    // Extract coordinates from route waypoints
    const coordinates = (activity.route?.waypoints || [])
      .map(w => [w.lng, w.lat]);

    if (coordinates.length < 2) return;

    // Add route data source
    this.map.addSource(this.sourceId, {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: {
          activity_type: activity.activity_type,
          difficulty: activity.metadata?.difficulty,
          title: activity.title,
        },
        geometry: {
          type: 'LineString',
          coordinates,
        },
      },
    });

    // Get styling based on activity type and difficulty
    const routeStyle = this.getActivityRouteStyle(
      activity.activity_type,
      activity.metadata?.difficulty,
      isSelected,
      isHovered
    );

    // Add route area/buffer for certain activity types
    if (routeStyle.showArea && coordinates.length >= 2) {
      this.addActivityArea(coordinates, routeStyle);
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

    // Add start/end markers for certain route types
    if (activity.route?.type === 'point-to-point') {
      this.addStartEndMarkers(coordinates, routeStyle);
    }

    // Add interaction handlers
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

  private getActivityRouteStyle(
    activityType: string,
    difficulty?: string,
    isSelected?: boolean,
    isHovered?: boolean
  ) {
    const baseStyle = {
      width: 3,
      opacity: 0.8,
      dashArray: undefined as number[] | undefined,
      showArea: false,
      areaColor: '#6366f1',
      areaOpacity: 0.1,
    };

    // Adjust for selection/hover states
    if (isSelected) {
      baseStyle.width = 6;
      baseStyle.opacity = 1.0;
    } else if (isHovered) {
      baseStyle.width = 4;
      baseStyle.opacity = 0.9;
    }

    // Activity type specific colors and styles
    let color = '#6366f1'; // Default blue
    
    switch (activityType?.toLowerCase()) {
      case 'hiking':
        color = '#16a34a'; // Green
        baseStyle.showArea = true;
        baseStyle.areaColor = '#16a34a';
        break;
      case 'backpacking':
        color = '#15803d'; // Darker green
        baseStyle.showArea = true;
        baseStyle.areaColor = '#15803d';
        baseStyle.width += 1;
        break;
      case 'trail-running':
        color = '#ea580c'; // Orange
        baseStyle.dashArray = [8, 4];
        break;
      case 'biking':
      case 'mountain-biking':
        color = '#dc2626'; // Red
        baseStyle.width += 1;
        break;
      case 'road-cycling':
        color = '#b91c1c'; // Darker red
        baseStyle.dashArray = [12, 4];
        break;
      case 'climbing':
      case 'rock-climbing':
        color = '#92400e'; // Brown
        baseStyle.dashArray = [6, 6];
        break;
      case 'skiing':
        color = '#1d4ed8'; // Blue
        baseStyle.dashArray = [10, 5];
        break;
      case 'snowboarding':
        color = '#1e40af'; // Darker blue
        baseStyle.dashArray = [8, 8];
        break;
      case 'kayaking':
        color = '#0891b2'; // Cyan
        baseStyle.showArea = true;
        baseStyle.areaColor = '#0891b2';
        break;
      case 'canoeing':
        color = '#0e7490'; // Darker cyan
        baseStyle.showArea = true;
        baseStyle.areaColor = '#0e7490';
        break;
      case 'fishing':
        color = '#0d9488'; // Teal
        baseStyle.dashArray = [4, 4];
        break;
      case 'camping':
        color = '#65a30d'; // Lime
        baseStyle.showArea = true;
        baseStyle.areaColor = '#65a30d';
        baseStyle.areaOpacity = 0.2;
        break;
      case 'photography':
        color = '#7c3aed'; // Purple
        baseStyle.dashArray = [12, 6];
        break;
      case 'wildlife-viewing':
        color = '#8b5cf6'; // Lighter purple
        baseStyle.dashArray = [6, 3];
        break;
      default:
        color = '#6366f1'; // Default indigo
    }

    // Difficulty level modifications
    switch (difficulty?.toLowerCase()) {
      case 'easy':
        baseStyle.opacity *= 0.9;
        color = this.lightenColor(color, 0.1);
        break;
      case 'moderate':
        // No change to base style
        break;
      case 'hard':
        baseStyle.width += 1;
        color = this.darkenColor(color, 0.1);
        break;
      case 'expert':
        baseStyle.width += 2;
        color = this.darkenColor(color, 0.2);
        // Add subtle glow effect for expert routes
        break;
    }

    return {
      color,
      ...baseStyle,
    };
  }

  private addActivityArea(coordinates: number[][], style: any) {
    // Create activity area representation
    const buffered = this.createActivityBuffer(coordinates, style);
    
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

    this.map.addLayer(
      {
        id: this.routeAreaLayerId,
        type: 'fill',
        source: areaSourceId,
        paint: {
          'fill-color': style.areaColor,
          'fill-opacity': style.areaOpacity,
        },
      },
      this.layerId // Insert before the route line
    );
  }

  private addStartEndMarkers(coordinates: number[][], style: any) {
    const startCoord = coordinates[0];
    const endCoord = coordinates[coordinates.length - 1];

    // Start marker
    this.startMarkerId = `${this.layerId}-start`;
    this.map.addSource(this.startMarkerId, {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: { type: 'start' },
        geometry: {
          type: 'Point',
          coordinates: startCoord,
        },
      },
    });

    this.map.addLayer({
      id: this.startMarkerId,
      type: 'circle',
      source: this.startMarkerId,
      paint: {
        'circle-radius': 6,
        'circle-color': style.color,
        'circle-stroke-width': 2,
        'circle-stroke-color': '#ffffff',
      },
    });

    // End marker
    this.endMarkerId = `${this.layerId}-end`;
    this.map.addSource(this.endMarkerId, {
      type: 'geojson',
      data: {
        type: 'Feature',
        properties: { type: 'end' },
        geometry: {
          type: 'Point',
          coordinates: endCoord,
        },
      },
    });

    this.map.addLayer({
      id: this.endMarkerId,
      type: 'circle',
      source: this.endMarkerId,
      paint: {
        'circle-radius': 6,
        'circle-color': '#ffffff',
        'circle-stroke-width': 3,
        'circle-stroke-color': style.color,
      },
    });
  }

  private createActivityBuffer(coordinates: number[][], style: any): number[][] {
    // Create appropriate buffer based on activity type
    const bufferDistance = this.getBufferDistance(style);
    
    // Simple buffer implementation - in production, use turf.js
    const buffered: number[][] = [];
    
    coordinates.forEach((coord, index) => {
      const [lng, lat] = coord;
      const offset = bufferDistance;
      
      // Create buffer points around the route
      if (index === 0) {
        // Start of route
        buffered.push([lng - offset, lat - offset]);
        buffered.push([lng - offset, lat + offset]);
      }
      
      buffered.push([lng + offset, lat + offset]);
      
      if (index === coordinates.length - 1) {
        // End of route
        buffered.push([lng + offset, lat - offset]);
        buffered.push([lng - offset, lat - offset]);
      }
    });
    
    // Close the polygon
    if (buffered.length > 0) {
      buffered.push(buffered[0]);
    }
    
    return buffered;
  }

  private getBufferDistance(style: any): number {
    // Different buffer sizes for different activity types
    switch (style.activityType) {
      case 'camping':
        return 0.001; // Larger area for campgrounds
      case 'kayaking':
      case 'canoeing':
        return 0.0005; // Water activity area
      case 'hiking':
      case 'backpacking':
        return 0.0003; // Trail corridor
      default:
        return 0.0002; // Default small buffer
    }
  }

  private lightenColor(color: string, amount: number): string {
    const hex = color.replace('#', '');
    const r = Math.min(255, parseInt(hex.substr(0, 2), 16) + (255 * amount));
    const g = Math.min(255, parseInt(hex.substr(2, 2), 16) + (255 * amount));
    const b = Math.min(255, parseInt(hex.substr(4, 2), 16) + (255 * amount));
    
    return `#${Math.round(r).toString(16).padStart(2, '0')}${Math.round(g).toString(16).padStart(2, '0')}${Math.round(b).toString(16).padStart(2, '0')}`;
  }

  private darkenColor(color: string, amount: number): string {
    const hex = color.replace('#', '');
    const r = Math.max(0, parseInt(hex.substr(0, 2), 16) * (1 - amount));
    const g = Math.max(0, parseInt(hex.substr(2, 2), 16) * (1 - amount));
    const b = Math.max(0, parseInt(hex.substr(4, 2), 16) * (1 - amount));
    
    return `#${Math.round(r).toString(16).padStart(2, '0')}${Math.round(g).toString(16).padStart(2, '0')}${Math.round(b).toString(16).padStart(2, '0')}`;
  }

  // Update the route appearance (e.g., for selection changes)
  updateStyle(isSelected?: boolean, isHovered?: boolean) {
    if (!this.map.getLayer(this.layerId)) return;

    // Get the activity data from the source
    const source = this.map.getSource(this.sourceId) as mapboxgl.GeoJSONSource;
    if (!source) return;

    // Update the route style
    const data = (source as any)._data;
    const activityType = data?.properties?.activity_type;
    const difficulty = data?.properties?.difficulty;
    
    const routeStyle = this.getActivityRouteStyle(
      activityType,
      difficulty,
      isSelected,
      isHovered
    );

    this.map.setPaintProperty(this.layerId, 'line-color', routeStyle.color);
    this.map.setPaintProperty(this.layerId, 'line-width', routeStyle.width);
    this.map.setPaintProperty(this.layerId, 'line-opacity', routeStyle.opacity);
  }

  remove() {
    // Remove main route
    if (this.map.getLayer(this.layerId)) {
      this.map.removeLayer(this.layerId);
    }
    if (this.map.getSource(this.sourceId)) {
      this.map.removeSource(this.sourceId);
    }
    
    // Remove area layer
    if (this.routeAreaLayerId && this.map.getLayer(this.routeAreaLayerId)) {
      this.map.removeLayer(this.routeAreaLayerId);
    }
    const areaSourceId = `${this.sourceId}-area`;
    if (this.map.getSource(areaSourceId)) {
      this.map.removeSource(areaSourceId);
    }

    // Remove start/end markers
    if (this.startMarkerId && this.map.getLayer(this.startMarkerId)) {
      this.map.removeLayer(this.startMarkerId);
      this.map.removeSource(this.startMarkerId);
    }
    if (this.endMarkerId && this.map.getLayer(this.endMarkerId)) {
      this.map.removeLayer(this.endMarkerId);
      this.map.removeSource(this.endMarkerId);
    }
  }
}