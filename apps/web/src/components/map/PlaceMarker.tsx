import mapboxgl from 'mapbox-gl';
import { Place } from '../../types';

interface PlaceMarkerOptions {
  place: Place;
  map: mapboxgl.Map;
  onClick?: () => void;
}

export class PlaceMarker {
  private marker: mapboxgl.Marker | null = null;

  constructor(options: PlaceMarkerOptions) {
    const { place, map, onClick } = options;

    if (!place.location) {
      this.marker = null;
      return;
    }

    // Create a custom marker element
    const el = document.createElement('div');
    el.className = 'custom-marker';
    el.style.width = '32px';
    el.style.height = '32px';
    el.style.backgroundImage = `url('/marker-icon.png')`;
    el.style.backgroundSize = 'cover';
    el.style.cursor = 'pointer';

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