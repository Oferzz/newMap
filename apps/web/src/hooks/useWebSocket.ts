import { useEffect, useRef, useCallback } from 'react';
import { useAppSelector } from './redux';
import { websocketService } from '../services/websocket.service';

interface UseWebSocketOptions {
  autoConnect?: boolean;
  room?: string;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { autoConnect = true, room } = options;
  const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated);
  const isConnectedRef = useRef(false);

  useEffect(() => {
    if (autoConnect && isAuthenticated && !isConnectedRef.current) {
      websocketService.connect();
      isConnectedRef.current = true;
    }

    return () => {
      if (isConnectedRef.current) {
        websocketService.disconnect();
        isConnectedRef.current = false;
      }
    };
  }, [autoConnect, isAuthenticated]);

  useEffect(() => {
    if (room && websocketService.isConnected()) {
      websocketService.joinRoom(room);

      return () => {
        websocketService.leaveRoom(room);
      };
    }
  }, [room]);

  const emit = useCallback((event: string, data?: any) => {
    websocketService.emit(event, data);
  }, []);

  const on = useCallback((event: any, callback: (data: any) => void) => {
    return websocketService.on(event, callback);
  }, []);

  const broadcastCursor = useCallback((position: { lat: number; lng: number }) => {
    websocketService.broadcastCursor(position);
  }, []);

  const broadcastTyping = useCallback((isTyping: boolean, context: string) => {
    websocketService.broadcastTyping(isTyping, context);
  }, []);

  return {
    isConnected: websocketService.isConnected(),
    emit,
    on,
    broadcastCursor,
    broadcastTyping,
  };
}