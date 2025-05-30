import React, { useEffect } from 'react';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { removeNotification } from '../../store/slices/uiSlice';

export const Notifications: React.FC = () => {
  const notifications = useAppSelector((state) => state.ui.notifications);
  const dispatch = useAppDispatch();

  useEffect(() => {
    // Auto-remove notifications after 5 seconds
    const timers = notifications.map((notification) => {
      return setTimeout(() => {
        dispatch(removeNotification(notification.id));
      }, 5000);
    });

    return () => {
      timers.forEach(clearTimeout);
    };
  }, [notifications, dispatch]);

  if (notifications.length === 0) return null;

  return (
    <div className="fixed top-20 right-4 z-50 space-y-2">
      {notifications.map((notification) => (
        <div
          key={notification.id}
          className={`
            p-4 rounded-lg shadow-lg flex items-center justify-between min-w-[300px]
            ${notification.type === 'success' ? 'bg-green-500 text-white' : ''}
            ${notification.type === 'error' ? 'bg-red-500 text-white' : ''}
            ${notification.type === 'info' ? 'bg-blue-500 text-white' : ''}
            ${notification.type === 'warning' ? 'bg-yellow-500 text-white' : ''}
          `}
        >
          <span className="mr-4">{notification.message}</span>
          <button
            onClick={() => dispatch(removeNotification(notification.id))}
            className="text-white hover:text-gray-200"
          >
            ✕
          </button>
        </div>
      ))}
    </div>
  );
};