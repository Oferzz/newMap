import React from 'react';
// @ts-ignore
import mapboxgl from 'mapbox-gl';
import { Layers } from 'lucide-react';

interface MapControlsProps {
  map: mapboxgl.Map | null;
}

export const MapControls: React.FC<MapControlsProps> = ({ map }) => {
  const mapStyles = [
    { id: 'streets-v12', name: 'Streets' },
    { id: 'outdoors-v12', name: 'Outdoors' },
    { id: 'light-v11', name: 'Light' },
    { id: 'dark-v11', name: 'Dark' },
    { id: 'satellite-streets-v12', name: 'Satellite' },
  ];

  const [isStyleMenuOpen, setIsStyleMenuOpen] = React.useState(false);
  const [currentStyle, setCurrentStyle] = React.useState('outdoors-v12');

  const handleStyleChange = (styleId: string) => {
    if (!map) return;
    
    map.setStyle(`mapbox://styles/mapbox/${styleId}`);
    setCurrentStyle(styleId);
    setIsStyleMenuOpen(false);
  };

  return (
    <div className="absolute bottom-10 left-4 z-10">
      <div className="relative">
        <button
          onClick={() => setIsStyleMenuOpen(!isStyleMenuOpen)}
          className="bg-terrain-50 p-3 rounded-map border border-terrain-300 shadow-map-control hover:shadow-medium transition-all hover:bg-terrain-100"
          title="Change map style"
        >
          <Layers className="w-5 h-5 text-trail-700" />
        </button>

        {isStyleMenuOpen && (
          <div className="absolute bottom-full left-0 mb-2 bg-terrain-50 rounded-lg shadow-medium border border-terrain-300 py-2 min-w-[150px]">
            {mapStyles.map((style) => (
              <button
                key={style.id}
                onClick={() => handleStyleChange(style.id)}
                className={`w-full text-left px-4 py-2 hover:bg-terrain-200 transition-colors ${
                  currentStyle === style.id ? 'bg-forest-100 text-forest-700 font-medium' : 'text-trail-700'
                }`}
              >
                {style.name}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};