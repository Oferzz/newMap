import React, { useRef, useEffect, useCallback } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { PlaceMarker } from './PlaceMarker';
import { TripRoute } from './TripRoute';
import { MapControls } from './MapControls';
import { SearchOverlay } from '../search/SearchOverlay';
import { DetailsPanel } from '../details/DetailsPanel';
import { TripPlanningPanel } from '../trips/TripPlanningPanel';
import { CollaborativeCursors } from './CollaborativeCursors';
import { useParams } from 'react-router-dom';
import { Place, Trip, SearchResult } from '../../types';

// Initialize Mapbox
mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

// Disable Mapbox telemetry to prevent CORS errors
(mapboxgl as any).config = {
  ...(mapboxgl as any).config,
  EVENTS_URL: ''
};

interface MapViewProps {
  onPlaceSelect?: (place: Place) => void;
  onTripSelect?: (trip: Trip) => void;
}

export const MapView: React.FC<MapViewProps> = ({ 
  onPlaceSelect, 
  onTripSelect 
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const dispatch = useAppDispatch();
  const { id: tripId } = useParams();

  // Redux state
  const places = useAppSelector((state) => state.places.items);
  const trips = useAppSelector((state) => state.trips.items);
  const selectedItem = useAppSelector((state) => state.ui.selectedItem);
  const activePanel = useAppSelector((state) => state.ui.activePanel);
  const mapViewState = useAppSelector((state) => state.ui.mapView);
  const searchResults = useAppSelector((state) => state.ui.searchResults);
  const isSearching = useAppSelector((state) => state.ui.isSearching);

  // Initialize map
  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: mapViewState.style || 'mapbox://styles/mapbox/outdoors-v12',
      center: mapViewState.center || [-74.5, 40],
      zoom: mapViewState.zoom || 9,
      pitch: 0,
      bearing: 0,
    });

    // Add navigation controls
    map.current.addControl(
      new mapboxgl.NavigationControl({
        visualizePitch: true,
      }),
      'top-right'
    );

    // Add geolocation control
    map.current.addControl(
      new mapboxgl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true,
        },
        trackUserLocation: true,
        showUserHeading: true,
      }),
      'top-right'
    );

    // Add scale control
    map.current.addControl(
      new mapboxgl.ScaleControl({
        maxWidth: 200,
        unit: 'metric',
      }),
      'bottom-left'
    );

    // Map event handlers
    map.current.on('moveend', () => {
      if (!map.current) return;
      const center = map.current.getCenter();
      const zoom = map.current.getZoom();
      
      dispatch({
        type: 'ui/updateMapView',
        payload: {
          center: [center.lng, center.lat],
          zoom,
        },
      });
    });

    // Click handler for adding new places
    map.current.on('click', (e) => {
      // Only handle clicks on the map itself, not on markers
      const features = map.current?.queryRenderedFeatures(e.point);
      if (features && features.length > 0) return;

      dispatch({
        type: 'ui/setMapClickLocation',
        payload: {
          coordinates: [e.lngLat.lng, e.lngLat.lat],
        },
      });
    });

    return () => {
      map.current?.remove();
    };
  }, []);

  // Update map style
  useEffect(() => {
    if (map.current && mapViewState.style) {
      map.current.setStyle(mapViewState.style);
    }
  }, [mapViewState.style]);

  // Handle selected item
  useEffect(() => {
    if (!map.current || !selectedItem) return;

    if ('location' in selectedItem && selectedItem.location) {
      // It's a place
      map.current.flyTo({
        center: selectedItem.location.coordinates,
        zoom: 15,
        duration: 1000,
      });
    } else if ('waypoints' in selectedItem && selectedItem.waypoints.length > 0) {
      // It's a trip - fit bounds to show all waypoints
      const bounds = new mapboxgl.LngLatBounds();
      selectedItem.waypoints.forEach((waypoint) => {
        const place = waypoint.place as any;
        if (place?.location) {
          bounds.extend(place.location.coordinates);
        } else if (place?.coordinates) {
          bounds.extend([place.coordinates.lng, place.coordinates.lat]);
        }
      });
      
      map.current.fitBounds(bounds, {
        padding: { top: 100, bottom: 100, left: 400, right: 400 },
        duration: 1000,
      });
    }
  }, [selectedItem]);

  // Render places on map
  useEffect(() => {
    if (!map.current) return;

    places.forEach((place) => {
      if (place.location || (place as any).coordinates) {
        // Transform place if it has coordinates instead of location
        const transformedPlace = place.location ? place : {
          ...place,
          location: {
            coordinates: [(place as any).coordinates.lng, (place as any).coordinates.lat]
          }
        };
        
        new PlaceMarker({
          place: transformedPlace as Place,
          map: map.current!,
          onClick: () => onPlaceSelect?.(transformedPlace as Place),
        });
      }
    });
  }, [places, onPlaceSelect]);

  // Render trips on map
  useEffect(() => {
    if (!map.current) return;

    trips.forEach((trip) => {
      if (trip.waypoints && trip.waypoints.length > 1) {
        new TripRoute({
          trip: trip as unknown as Trip,
          map: map.current!,
          onClick: () => onTripSelect?.(trip as unknown as Trip),
        });
      }
    });
  }, [trips, onTripSelect]);

  const handleSearchResultSelect = useCallback((result: SearchResult) => {
    dispatch({ type: 'ui/selectItem', payload: result });
    dispatch({ type: 'ui/setActivePanel', payload: 'details' });
    dispatch({ type: 'ui/clearSearch' });
  }, [dispatch]);

  const handleClosePanel = useCallback(() => {
    dispatch({ type: 'ui/setActivePanel', payload: 'none' });
    dispatch({ type: 'ui/clearSelectedItem' });
  }, [dispatch]);

  return (
    <div className="absolute inset-0 top-16">
      {/* Map Container */}
      <div ref={mapContainer} className="absolute inset-0" />

      {/* Map Controls */}
      <MapControls map={map.current} />

      {/* Collaborative Cursors for trips */}
      {tripId && <CollaborativeCursors map={map.current} tripId={tripId} />}

      {/* Search Overlay */}
      {isSearching && searchResults && (
        <SearchOverlay
          results={{ places: [], trips: [], users: [] }}
          onSelect={handleSearchResultSelect}
          onClose={() => dispatch({ type: 'ui/clearSearch' })}
        />
      )}

      {/* Details Panel */}
      {activePanel === 'details' && selectedItem && (
        <DetailsPanel
          item={selectedItem as Place | Trip}
          onClose={handleClosePanel}
        />
      )}

      {/* Trip Planning Panel */}
      <TripPlanningPanel
        isOpen={activePanel === 'trip-planning'}
        onClose={handleClosePanel}
      />
    </div>
  );
};