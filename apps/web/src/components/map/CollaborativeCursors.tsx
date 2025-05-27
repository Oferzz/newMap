import React, { useEffect, useState, useRef } from 'react';
import mapboxgl from 'mapbox-gl';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useAppSelector } from '../../hooks/redux';

interface CollaborativeCursorsProps {
  map: mapboxgl.Map | null;
  tripId?: string;
}

interface UserCursor {
  userId: string;
  userName: string;
  position: { lat: number; lng: number };
  color: string;
  lastUpdate: number;
}

const CURSOR_COLORS = [
  '#EF4444', // red
  '#F59E0B', // amber
  '#10B981', // emerald
  '#3B82F6', // blue
  '#8B5CF6', // violet
  '#EC4899', // pink
  '#14B8A6', // teal
  '#F97316', // orange
];

export const CollaborativeCursors: React.FC<CollaborativeCursorsProps> = ({ map, tripId }) => {
  const [cursors, setCursors] = useState<Map<string, UserCursor>>(new Map());
  const markersRef = useRef<Map<string, mapboxgl.Marker>>(new Map());
  const colorIndexRef = useRef(0);
  const currentUser = useAppSelector(state => state.auth.user);
  const { on, broadcastCursor } = useWebSocket({ room: tripId ? `trip:${tripId}` : undefined });

  // Handle cursor updates
  useEffect(() => {
    if (!tripId) return;

    const unsubscribe = on('user:cursor:moved', (data) => {
      if (data.userId === currentUser?.id) return; // Don't show own cursor

      setCursors(prev => {
        const newCursors = new Map(prev);
        const existingCursor = newCursors.get(data.userId);
        
        newCursors.set(data.userId, {
          userId: data.userId,
          userName: data.userName || 'Anonymous',
          position: data.position,
          color: existingCursor?.color || CURSOR_COLORS[colorIndexRef.current++ % CURSOR_COLORS.length],
          lastUpdate: Date.now(),
        });
        
        return newCursors;
      });
    });

    return unsubscribe;
  }, [tripId, currentUser?.id, on]);

  // Broadcast own cursor movements
  useEffect(() => {
    if (!map || !tripId) return;

    let lastBroadcast = 0;
    const BROADCAST_THROTTLE = 100; // ms

    const handleMouseMove = (e: mapboxgl.MapMouseEvent) => {
      const now = Date.now();
      if (now - lastBroadcast < BROADCAST_THROTTLE) return;

      lastBroadcast = now;
      broadcastCursor({
        lat: e.lngLat.lat,
        lng: e.lngLat.lng,
      });
    };

    map.on('mousemove', handleMouseMove);

    return () => {
      map.off('mousemove', handleMouseMove);
    };
  }, [map, tripId, broadcastCursor]);

  // Update markers on map
  useEffect(() => {
    if (!map) return;

    // Update or create markers
    cursors.forEach((cursor, userId) => {
      let marker = markersRef.current.get(userId);

      if (!marker) {
        // Create cursor element
        const el = document.createElement('div');
        el.className = 'collaborative-cursor';
        el.style.cssText = `
          width: 24px;
          height: 24px;
          background-color: ${cursor.color};
          border: 2px solid white;
          border-radius: 50%;
          box-shadow: 0 2px 4px rgba(0,0,0,0.3);
          position: relative;
          transition: transform 0.1s ease-out;
          z-index: 100;
        `;

        // Add name label
        const label = document.createElement('div');
        label.textContent = cursor.userName;
        label.style.cssText = `
          position: absolute;
          bottom: 100%;
          left: 50%;
          transform: translateX(-50%);
          background-color: ${cursor.color};
          color: white;
          padding: 2px 8px;
          border-radius: 4px;
          font-size: 12px;
          white-space: nowrap;
          margin-bottom: 4px;
          box-shadow: 0 2px 4px rgba(0,0,0,0.2);
        `;
        el.appendChild(label);

        // Create marker
        marker = new mapboxgl.Marker(el)
          .setLngLat([cursor.position.lng, cursor.position.lat])
          .addTo(map);

        markersRef.current.set(userId, marker);
      } else {
        // Update existing marker position
        marker.setLngLat([cursor.position.lng, cursor.position.lat]);
      }
    });

    // Remove stale markers
    markersRef.current.forEach((marker, userId) => {
      if (!cursors.has(userId)) {
        marker.remove();
        markersRef.current.delete(userId);
      }
    });
  }, [map, cursors]);

  // Remove stale cursors (no update for 5 seconds)
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      const STALE_THRESHOLD = 5000; // 5 seconds

      setCursors(prev => {
        const newCursors = new Map(prev);
        let changed = false;

        prev.forEach((cursor, userId) => {
          if (now - cursor.lastUpdate > STALE_THRESHOLD) {
            newCursors.delete(userId);
            changed = true;
          }
        });

        return changed ? newCursors : prev;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Cleanup markers on unmount
  useEffect(() => {
    return () => {
      markersRef.current.forEach(marker => marker.remove());
      markersRef.current.clear();
    };
  }, []);

  return null; // This component doesn't render anything directly
};