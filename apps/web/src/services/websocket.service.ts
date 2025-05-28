import { io, Socket } from 'socket.io-client';
import { store } from '../store';
import { 
  addNotification,
  setActivePanel,
  selectItem
} from '../store/slices/uiSlice';
import {
  updateTrip,
  addWaypoint,
  updateWaypoint,
  removeWaypoint,
  reorderWaypoints
} from '../store/slices/tripsSlice';
import {
  updatePlace,
  addPlace,
  deletePlace
} from '../store/slices/placesSlice';

type EventCallback = (data: any) => void;

interface WebSocketEvents {
  // Connection events
  connect: () => void;
  disconnect: () => void;
  error: (error: Error) => void;
  
  // Trip events
  'trip:created': (data: any) => void;
  'trip:updated': (data: any) => void;
  'trip:deleted': (data: { tripId: string }) => void;
  'trip:collaborator:added': (data: any) => void;
  'trip:collaborator:removed': (data: any) => void;
  'trip:waypoint:added': (data: any) => void;
  'trip:waypoint:updated': (data: any) => void;
  'trip:waypoint:removed': (data: any) => void;
  'trip:waypoints:reordered': (data: any) => void;
  
  // Place events
  'place:created': (data: any) => void;
  'place:updated': (data: any) => void;
  'place:deleted': (data: { placeId: string }) => void;
  'place:media:added': (data: any) => void;
  'place:media:removed': (data: any) => void;
  
  // Suggestion events
  'suggestion:created': (data: any) => void;
  'suggestion:approved': (data: any) => void;
  'suggestion:rejected': (data: any) => void;
  
  // Real-time collaboration
  'user:joined': (data: { userId: string; userName: string; tripId?: string }) => void;
  'user:left': (data: { userId: string; userName: string; tripId?: string }) => void;
  'user:cursor:moved': (data: { userId: string; position: { lat: number; lng: number } }) => void;
  'user:typing': (data: { userId: string; isTyping: boolean; context: string }) => void;
}

class WebSocketService {
  private socket: Socket | null = null;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Map<string, Set<EventCallback>> = new Map();
  private isIntentionalDisconnect = false;

  connect(): void {
    const state = store.getState();
    const token = state.auth.accessToken;
    
    if (!token) {
      console.warn('Cannot connect to WebSocket without authentication token');
      return;
    }

    // Don't connect if already connected
    if (this.socket?.connected) {
      return;
    }

    const wsUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    
    this.socket = io(wsUrl, {
      auth: {
        token,
      },
      transports: ['websocket', 'polling'],
      reconnection: true,
      reconnectionAttempts: this.maxReconnectAttempts,
      reconnectionDelay: this.reconnectDelay,
    });

    this.setupEventHandlers();
    this.isIntentionalDisconnect = false;
  }

  disconnect(): void {
    this.isIntentionalDisconnect = true;
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
    this.listeners.clear();
  }

  private setupEventHandlers(): void {
    if (!this.socket) return;

    // Connection events
    this.socket.on('connect', () => {
      console.log('WebSocket connected');
      
      store.dispatch(addNotification({
        type: 'success',
        message: 'Real-time updates connected',
      }));
    });

    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      
      if (!this.isIntentionalDisconnect) {
        store.dispatch(addNotification({
          type: 'warning',
          message: 'Real-time updates disconnected. Reconnecting...',
        }));
      }
    });

    this.socket.on('error', (error) => {
      console.error('WebSocket error:', error);
      
      store.dispatch(addNotification({
        type: 'error',
        message: 'Connection error. Some features may not update in real-time.',
      }));
    });

    // Trip events
    this.socket.on('trip:updated', (data) => {
      store.dispatch(updateTrip(data.trip));
      this.notifyListeners('trip:updated', data);
    });

    this.socket.on('trip:deleted', (data) => {
      // If viewing the deleted trip, redirect
      const state = store.getState();
      if (state.trips.currentTrip?.id === data.tripId) {
        store.dispatch(setActivePanel('none'));
        store.dispatch(selectItem(null));
        window.location.href = '/';
      }
      this.notifyListeners('trip:deleted', data);
    });

    this.socket.on('trip:collaborator:added', (data) => {
      store.dispatch(addNotification({
        type: 'info',
        message: `${data.userName} was added to the trip`,
      }));
      this.notifyListeners('trip:collaborator:added', data);
    });

    this.socket.on('trip:waypoint:added', (data) => {
      store.dispatch(addWaypoint({
        tripId: data.tripId,
        waypoint: data.waypoint,
      }));
      this.notifyListeners('trip:waypoint:added', data);
    });

    this.socket.on('trip:waypoint:updated', (data) => {
      store.dispatch(updateWaypoint({
        tripId: data.tripId,
        waypoint: data.waypoint,
      }));
      this.notifyListeners('trip:waypoint:updated', data);
    });

    this.socket.on('trip:waypoint:removed', (data) => {
      store.dispatch(removeWaypoint({
        tripId: data.tripId,
        waypointId: data.waypointId,
      }));
      this.notifyListeners('trip:waypoint:removed', data);
    });

    this.socket.on('trip:waypoints:reordered', (data) => {
      store.dispatch(reorderWaypoints({
        tripId: data.tripId,
        waypoints: data.waypoints,
      }));
      this.notifyListeners('trip:waypoints:reordered', data);
    });

    // Place events
    this.socket.on('place:created', (data) => {
      store.dispatch(addPlace(data.place));
      this.notifyListeners('place:created', data);
    });

    this.socket.on('place:updated', (data) => {
      store.dispatch(updatePlace(data.place));
      this.notifyListeners('place:updated', data);
    });

    this.socket.on('place:deleted', (data) => {
      store.dispatch(deletePlace(data.placeId));
      this.notifyListeners('place:deleted', data);
    });

    // Real-time collaboration events
    this.socket.on('user:joined', (data) => {
      store.dispatch(addNotification({
        type: 'info',
        message: `${data.userName} joined`,
      }));
      this.notifyListeners('user:joined', data);
    });

    this.socket.on('user:left', (data) => {
      store.dispatch(addNotification({
        type: 'info',
        message: `${data.userName} left`,
      }));
      this.notifyListeners('user:left', data);
    });

    this.socket.on('user:cursor:moved', (data) => {
      this.notifyListeners('user:cursor:moved', data);
    });

    this.socket.on('user:typing', (data) => {
      this.notifyListeners('user:typing', data);
    });
  }

  // Subscribe to events
  on<K extends keyof WebSocketEvents>(event: K, callback: WebSocketEvents[K]): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    
    const callbacks = this.listeners.get(event)!;
    callbacks.add(callback as EventCallback);
    
    // Return unsubscribe function
    return () => {
      callbacks.delete(callback as EventCallback);
      if (callbacks.size === 0) {
        this.listeners.delete(event);
      }
    };
  }

  // Emit events
  emit(event: string, data?: any): void {
    if (!this.socket?.connected) {
      console.warn('Cannot emit event: WebSocket not connected');
      return;
    }
    
    this.socket.emit(event, data);
  }

  // Join a room (e.g., for trip collaboration)
  joinRoom(room: string): void {
    this.emit('join:room', { room });
  }

  // Leave a room
  leaveRoom(room: string): void {
    this.emit('leave:room', { room });
  }

  // Broadcast cursor position
  broadcastCursor(position: { lat: number; lng: number }): void {
    this.emit('cursor:move', position);
  }

  // Broadcast typing status
  broadcastTyping(isTyping: boolean, context: string): void {
    this.emit('typing:status', { isTyping, context });
  }

  private notifyListeners(event: string, data: any): void {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in WebSocket event listener for ${event}:`, error);
        }
      });
    }
  }

  // Check connection status
  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  // Get socket instance (for advanced usage)
  getSocket(): Socket | null {
    return this.socket;
  }
}

// Export singleton instance
export const websocketService = new WebSocketService();