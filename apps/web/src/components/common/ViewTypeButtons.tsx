import React from 'react';
import { MapIcon, List, Grid3x3, Calendar } from 'lucide-react';

export type ViewType = 'map' | 'list' | 'grid' | 'calendar';

interface ViewTypeButtonsProps {
  activeView: ViewType;
  onViewChange: (view: ViewType) => void;
}

export const ViewTypeButtons: React.FC<ViewTypeButtonsProps> = ({ activeView, onViewChange }) => {
  const viewTypes: { type: ViewType; icon: React.ReactNode; label: string }[] = [
    { type: 'map', icon: <MapIcon className="w-4 h-4" />, label: 'Map' },
    { type: 'list', icon: <List className="w-4 h-4" />, label: 'List' },
    { type: 'grid', icon: <Grid3x3 className="w-4 h-4" />, label: 'Grid' },
    { type: 'calendar', icon: <Calendar className="w-4 h-4" />, label: 'Calendar' },
  ];

  return (
    <div className="flex items-center justify-center gap-2 p-2 bg-white/80 backdrop-blur-sm rounded-full shadow-sm">
      {viewTypes.map(({ type, icon, label }) => (
        <button
          key={type}
          onClick={() => onViewChange(type)}
          className={`
            flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-200
            ${activeView === type 
              ? 'bg-blue-500 text-white shadow-md' 
              : 'text-gray-600 hover:text-gray-800 hover:bg-gray-100'
            }
          `}
          aria-label={`${label} view`}
        >
          {icon}
          <span className="hidden sm:inline">{label}</span>
        </button>
      ))}
    </div>
  );
};