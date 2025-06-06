import React from 'react';
import { MapIcon, List, Grid3x3 } from 'lucide-react';

export type ViewType = 'map' | 'list' | 'grid';

interface ViewTypeButtonsProps {
  activeView: ViewType;
  onViewChange: (view: ViewType) => void;
}

export const ViewTypeButtons: React.FC<ViewTypeButtonsProps> = ({ activeView, onViewChange }) => {
  const viewTypes: { type: ViewType; icon: React.ReactNode; label: string }[] = [
    { type: 'map', icon: <MapIcon className="w-4 h-4" />, label: 'Map' },
    { type: 'list', icon: <List className="w-4 h-4" />, label: 'List' },
    { type: 'grid', icon: <Grid3x3 className="w-4 h-4" />, label: 'Tiles' },
  ];

  return (
    <>
      {viewTypes.map(({ type, icon, label }) => (
        <button
          key={type}
          onClick={() => onViewChange(type)}
          className={`
            flex items-center gap-1 px-2 py-1 text-sm transition-all duration-200
            ${activeView === type 
              ? 'text-gray-900 font-bold' 
              : 'text-gray-500 font-normal hover:text-gray-700'
            }
          `}
          aria-label={`${label} view`}
        >
          {icon}
          <span className="hidden sm:inline">{label}</span>
        </button>
      ))}
    </>
  );
};