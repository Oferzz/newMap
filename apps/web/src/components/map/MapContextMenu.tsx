import React, { useEffect, useRef } from 'react';

interface MapContextMenuProps {
  coordinates: [number, number];
  position: { x: number; y: number };
  onClose: () => void;
  onSaveLocation: () => void;
  onCreateRoute: () => void;
  onAddToCollection: () => void;
  onDropPin: () => void;
}

export const MapContextMenu: React.FC<MapContextMenuProps> = ({
  coordinates,
  position,
  onClose,
  onSaveLocation,
  onCreateRoute,
  onAddToCollection,
  onDropPin,
}) => {
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [onClose]);

  const menuItems = [
    {
      icon: '📌',
      label: 'Drop Pin',
      action: onDropPin,
    },
    {
      icon: '📍',
      label: 'Save Location',
      action: onSaveLocation,
    },
    {
      icon: '🗺️',
      label: 'Create Route',
      action: onCreateRoute,
    },
    {
      icon: '📁',
      label: 'Add to Collection',
      action: onAddToCollection,
    },
  ];

  return (
    <div
      ref={menuRef}
      className="absolute z-50 bg-white rounded-lg shadow-xl border border-gray-200 py-1 min-w-[200px]"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`,
        transform: 'translate(-50%, 0)',
      }}
    >
      <div className="px-3 py-2 border-b border-gray-100 text-xs text-gray-500">
        {coordinates[0].toFixed(6)}, {coordinates[1].toFixed(6)}
      </div>
      {menuItems.map((item, index) => (
        <button
          key={index}
          className="w-full text-left px-3 py-2 hover:bg-gray-100 transition-colors flex items-center gap-2"
          onClick={() => {
            item.action();
            onClose();
          }}
        >
          <span className="text-lg">{item.icon}</span>
          <span className="text-sm">{item.label}</span>
        </button>
      ))}
    </div>
  );
};