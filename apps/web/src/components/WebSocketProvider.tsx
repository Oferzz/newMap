import { useEffect } from 'react';
import { useAppSelector } from '../hooks/redux';
import { websocketService } from '../services/websocket.service';

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated);
  const accessToken = useAppSelector(state => state.auth.accessToken);

  useEffect(() => {
    if (isAuthenticated && accessToken) {
      websocketService.connect();
    } else {
      websocketService.disconnect();
    }

    return () => {
      websocketService.disconnect();
    };
  }, [isAuthenticated, accessToken]);

  return <>{children}</>;
};