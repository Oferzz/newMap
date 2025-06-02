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
        height="${size}"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style="transform: translate(-50%, -50%); filter: drop-shadow(0 1px 2px rgba(0,0,0,0.3));"
      >
        <circle cx="12" cy="12" r="11" fill="white" stroke="${color}" stroke-width="2"/>
        <circle cx="12" cy="12" r="4" fill="${color}"/>
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