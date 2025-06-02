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
    
    // Create marker content with simple circle design
    const root = ReactDOM.createRoot(el);
    root.render(
      <div 
        className="relative cursor-pointer"
        style={{ transform: 'translate(-50%, -50%)' }}
      >
        <svg
          width="32"
          height="32"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="drop-shadow-lg hover:scale-110 transition-transform"
        >
          <circle cx="12" cy="12" r="11" fill="white" stroke="#446b8e" strokeWidth="2"/>
          <circle cx="12" cy="12" r="4" fill="#446b8e"/>
          <circle cx="12" cy="12" r="8" fill="none" stroke="#446b8e" strokeWidth="1" strokeDasharray="2 2" opacity="0.5"/>
        </svg>
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