import mapboxgl from 'mapbox-gl';
import { Place } from '../../types';
import { LocationPinColors } from '../common/LocationPin';

interface PlaceMarkerOptions {
  place: Place;
  map: mapboxgl.Map;
  onClick?: () => void;
  isSelected?: boolean;
}

export class PlaceMarker {
  private marker: mapboxgl.Marker | null = null;

  constructor(options: PlaceMarkerOptions) {
    const { place, map, onClick, isSelected = false } = options;

    if (!place.location) {
      this.marker = null;
      return;
    }

    // Create a custom marker element with SVG
    const el = document.createElement('div');
    el.className = 'custom-marker';
    el.style.cursor = 'pointer';
    
    // Create the SVG location pin
    const size = isSelected ? 36 : 28;
    const color = isSelected ? LocationPinColors.selected : LocationPinColors.default;
    
    el.innerHTML = `
      <svg
        width="${size}"
        height="${size * 1.2}"
        viewBox="0 0 24 29"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style="transform: translate(-50%, -100%); filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));"
      >
        <ellipse cx="12" cy="26" rx="3" ry="1.5" fill="rgba(0,0,0,0.2)" />
        <path d="M12 0C7.03 0 3 4.03 3 9c0 6.75 9 20 9 20s9-13.25 9-20c0-4.97-4.03-9-9-9z" fill="${color}" stroke="#fff" stroke-width="0.5" />
        <circle cx="12" cy="9" r="3.5" fill="white" />
        <circle cx="12" cy="9" r="1.5" fill="${color}" />
      </svg>
    `;

    // Create the marker
    this.marker = new mapboxgl.Marker(el)
      .setLngLat(place.location.coordinates)
      .setPopup(
        new mapboxgl.Popup({ offset: 25 })
          .setHTML(`
            <div class="p-2">
              <h3 class="font-bold">${place.name}</h3>
              <p class="text-sm text-gray-600">${place.address}</p>
            </div>
          `)
      )
      .addTo(map);

    // Add click handler
    if (onClick) {
      el.addEventListener('click', onClick);
    }
  }

  remove() {
    if (this.marker) {
      this.marker.remove();
    }
  }
}