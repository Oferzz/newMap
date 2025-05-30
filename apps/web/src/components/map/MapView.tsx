import React, { useRef, useEffect, useCallback } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import './map.css';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { PlaceMarker } from './PlaceMarker';
import { TripRoute } from './TripRoute';
import { MapControls } from './MapControls';
import { SearchOverlay } from '../search/SearchOverlay';
import { DetailsPanel } from '../details/DetailsPanel';
import { TripPlanningPanel } from '../trips/TripPlanningPanel';
import { PlaceCreationPanel } from '../places/PlaceCreationPanel';
import { CollectionsPanel } from '../collections/CollectionsPanel';
import { CollaborativeCursors } from './CollaborativeCursors';
import { TemporaryMarker } from './TemporaryMarker';
import { MapContextMenu } from './MapContextMenu';
import { RouteCreationOverlay } from './RouteCreationOverlay';
import { useParams } from 'react-router-dom';
import { Place, Trip, SearchResult } from '../../types';
import { 
  addTemporaryMarker, 
  removeTemporaryMarker, 
  openContextMenu, 
  closeContextMenu,
  addNotification,
  setActivePanel,
  startRouteCreation,
  addRouteWaypoint,
  clearMapClickLocation,
  startAddToCollection,
  cancelAddToCollection
} from '../../store/slices/uiSlice';

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
  const temporaryMarkersRef = useRef<Map<string, TemporaryMarker>>(new Map());
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
  const temporaryMarkers = useAppSelector((state) => state.ui.temporaryMarkers);
  const contextMenuState = useAppSelector((state) => state.ui.contextMenuState);
  const routeCreationMode = useAppSelector((state) => state.ui.routeCreationMode);
  const mapClickLocation = useAppSelector((state) => state.ui.mapClickLocation);
  const collectionsMode = useAppSelector((state) => state.ui.collectionsMode);

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
      maxPitch: 85,
      projection: 'globe'
    });

    // Enable 3D terrain when map loads
    map.current.on('load', () => {
      if (!map.current) return;
      
      // Add terrain source
      map.current.addSource('mapbox-dem', {
        type: 'raster-dem',
        url: 'mapbox://mapbox.mapbox-terrain-dem-v1',
        tileSize: 512,
        maxzoom: 14
      });
      
      // Add the terrain layer
      map.current.setTerrain({ 
        source: 'mapbox-dem', 
        exaggeration: 1.5 
      });
      
      // Add sky layer for better 3D effect
      map.current.addLayer({
        id: 'sky',
        type: 'sky',
        paint: {
          'sky-type': 'atmosphere',
          'sky-atmosphere-sun': [0.0, 90.0],
          'sky-atmosphere-sun-intensity': 15
        }
      });
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

    // Left click handler for adding temporary markers
    map.current.on('click', (e) => {
      // Only handle clicks on the map itself, not on markers
      const features = map.current?.queryRenderedFeatures(e.point);
      if (features && features.length > 0) return;

      const coordinates: [number, number] = [e.lngLat.lng, e.lngLat.lat];

      // Store coordinates for later use in click handler effect
      dispatch({ type: 'ui/setMapClickLocation', payload: { coordinates } });
    });

    // Right click handler for context menu
    map.current.on('contextmenu', (e) => {
      e.preventDefault();
      
      dispatch(openContextMenu({
        coordinates: [e.lngLat.lng, e.lngLat.lat],
        position: {
          x: e.point.x,
          y: e.point.y,
        },
      }));
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

  // Handle map clicks based on current mode
  useEffect(() => {
    if (!mapClickLocation) return;

    if (routeCreationMode.isActive) {
      dispatch(addRouteWaypoint({ coordinates: mapClickLocation }));
    } else {
      dispatch(addTemporaryMarker({ coordinates: mapClickLocation }));
    }

    dispatch(clearMapClickLocation());
  }, [mapClickLocation, routeCreationMode.isActive, dispatch]);

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
          place: transformedPlace as any,
          map: map.current!,
          onClick: () => onPlaceSelect?.(transformedPlace as any),
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

  // Render temporary markers
  useEffect(() => {
    if (!map.current) return;

    // Remove markers that no longer exist
    temporaryMarkersRef.current.forEach((marker, id) => {
      if (!temporaryMarkers.find(m => m.id === id)) {
        marker.remove();
        temporaryMarkersRef.current.delete(id);
      }
    });

    // Add new markers
    temporaryMarkers.forEach((markerData) => {
      if (!temporaryMarkersRef.current.has(markerData.id)) {
        const marker = new TemporaryMarker({
          coordinates: markerData.coordinates,
          map: map.current!,
          onRemove: () => {
            dispatch(removeTemporaryMarker(markerData.id));
          },
        });
        temporaryMarkersRef.current.set(markerData.id, marker);
      }
    });
  }, [temporaryMarkers, dispatch]);

  // Render route waypoints
  useEffect(() => {
    if (!map.current || !routeCreationMode.isActive) return;

    // Clear existing route markers when not in route mode
    if (!routeCreationMode.isActive) {
      return;
    }

    // Add markers for each waypoint
    routeCreationMode.waypoints.forEach((waypoint, index) => {
      const el = document.createElement('div');
      el.className = 'route-waypoint-marker';
      el.innerHTML = `
        <div class="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white font-bold shadow-lg">
          ${index + 1}
        </div>
      `;

      new mapboxgl.Marker(el)
        .setLngLat(waypoint.coordinates)
        .addTo(map.current!);
    });

    // Draw line between waypoints if more than one
    if (routeCreationMode.waypoints.length > 1) {
      const coordinates = routeCreationMode.waypoints.map(wp => wp.coordinates);
      
      // Add or update the route line
      if (map.current.getSource('route-preview')) {
        (map.current.getSource('route-preview') as mapboxgl.GeoJSONSource).setData({
          type: 'Feature',
          properties: {},
          geometry: {
            type: 'LineString',
            coordinates: coordinates,
          },
        });
      } else {
        map.current.addSource('route-preview', {
          type: 'geojson',
          data: {
            type: 'Feature',
            properties: {},
            geometry: {
              type: 'LineString',
              coordinates: coordinates,
            },
          },
        });

        map.current.addLayer({
          id: 'route-preview',
          type: 'line',
          source: 'route-preview',
          layout: {
            'line-join': 'round',
            'line-cap': 'round',
          },
          paint: {
            'line-color': '#10b981',
            'line-width': 4,
            'line-dasharray': [2, 2],
          },
        });
      }
    }

    return () => {
      // Cleanup on unmount or when route mode changes
      if (map.current?.getLayer('route-preview')) {
        map.current.removeLayer('route-preview');
        map.current.removeSource('route-preview');
      }
    };
  }, [routeCreationMode, dispatch]);

  const handleSearchResultSelect = useCallback((result: SearchResult) => {
    dispatch({ type: 'ui/selectItem', payload: result });
    dispatch({ type: 'ui/setActivePanel', payload: 'details' });
    dispatch({ type: 'ui/clearSearch' });
  }, [dispatch]);

  const handleClosePanel = useCallback(() => {
    dispatch({ type: 'ui/setActivePanel', payload: 'none' });
    dispatch({ type: 'ui/clearSelectedItem' });
  }, [dispatch]);

  const handleSaveLocation = useCallback(() => {
    if (!contextMenuState.coordinates) return;
    
    // Open place creation panel
    dispatch(setActivePanel('place-creation'));
  }, [contextMenuState.coordinates, dispatch]);

  const handleCreateRoute = useCallback(() => {
    if (!contextMenuState.coordinates) return;
    
    dispatch(startRouteCreation({ 
      coordinates: contextMenuState.coordinates 
    }));
    
    dispatch(addNotification({
      type: 'info',
      message: 'Click on the map to add waypoints to your route',
    }));
  }, [contextMenuState.coordinates, dispatch]);

  const handleAddToCollection = useCallback(() => {
    if (!contextMenuState.coordinates) return;
    
    dispatch(startAddToCollection({ 
      coordinates: contextMenuState.coordinates 
    }));
  }, [contextMenuState.coordinates, dispatch]);

  return (
    <div className="absolute inset-0 top-16">
      {/* Map Container */}
      <div className="absolute inset-0">
        <div ref={mapContainer} className="w-full h-full" />
      </div>

      {/* Map Controls */}
      <MapControls map={map.current} />

      {/* Route Creation Overlay */}
      <RouteCreationOverlay />

      {/* Collaborative Cursors for trips */}
      {tripId && <CollaborativeCursors map={map.current} tripId={tripId} />}

      {/* Search Overlay */}
      {isSearching && searchResults && (
        <SearchOverlay
          results={searchResults}
          onSelect={handleSearchResultSelect}
          onClose={() => dispatch({ type: 'ui/clearSearch' })}
        />
      )}

      {/* Details Panel */}
      {activePanel === 'details' && selectedItem && (
        <DetailsPanel
          item={selectedItem as any}
          onClose={handleClosePanel}
        />
      )}

      {/* Trip Planning Panel */}
      <TripPlanningPanel
        isOpen={activePanel === 'trip-planning'}
        onClose={handleClosePanel}
      />

      {/* Place Creation Panel */}
      <PlaceCreationPanel
        isOpen={activePanel === 'place-creation'}
        onClose={handleClosePanel}
      />

      {/* Collections Panel */}
      <CollectionsPanel
        isOpen={activePanel === 'collections'}
        onClose={() => {
          handleClosePanel();
          dispatch(cancelAddToCollection());
        }}
        locationToAdd={collectionsMode.locationToAdd || undefined}
      />

      {/* Context Menu */}
      {contextMenuState.isOpen && contextMenuState.coordinates && contextMenuState.position && (
        <MapContextMenu
          coordinates={contextMenuState.coordinates}
          position={contextMenuState.position}
          onClose={() => dispatch(closeContextMenu())}
          onSaveLocation={handleSaveLocation}
          onCreateRoute={handleCreateRoute}
          onAddToCollection={handleAddToCollection}
        />
      )}
    </div>
  );
};