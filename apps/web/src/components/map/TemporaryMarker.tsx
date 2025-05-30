import mapboxgl from 'mapbox-gl';
import ReactDOM from 'react-dom/client';

interface TemporaryMarkerProps {
  coordinates: [number, number];
  map: mapboxgl.Map;
  onRemove?: () => void;
}

export class TemporaryMarker {
  private marker: mapboxgl.Marker;
  private coordinates: [number, number];

  constructor({ coordinates, map, onRemove }: TemporaryMarkerProps) {
    this.coordinates = coordinates;

    // Create marker element
    const el = document.createElement('div');
    el.className = 'temporary-marker';
    el.setAttribute('data-testid', 'temporary-marker');
    
    // Create marker content
    const root = ReactDOM.createRoot(el);
    root.render(
      <div className="relative">
        <div className="absolute -top-8 -left-4 w-8 h-8 bg-blue-500 rounded-full shadow-lg cursor-pointer hover:bg-blue-600 transition-colors">
          <div className="absolute inset-2 bg-white rounded-full"></div>
        </div>
        <div className="absolute -top-1 -left-0.5 w-1 h-8 bg-blue-500"></div>
      </div>
    );

    // Create the marker
    this.marker = new mapboxgl.Marker(el, {
      offset: [0, 0]
    })
      .setLngLat(coordinates)
      .addTo(map);

    // Add click handler to remove
    el.addEventListener('click', (e) => {
      e.stopPropagation();
      this.remove();
      onRemove?.();
    });
  }

  remove() {
    this.marker.remove();
  }

  getCoordinates() {
    return this.coordinates;
  }
}