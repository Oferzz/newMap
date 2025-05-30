import React from 'react';
// @ts-ignore
import mapboxgl from 'mapbox-gl';
import { Layers, Mountain, Box } from 'lucide-react';

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
  const [isTerrainEnabled, setIsTerrainEnabled] = React.useState(true);
  const [is3DView, setIs3DView] = React.useState(false);

  // Listen for pitch changes to update 3D view state
  React.useEffect(() => {
    if (!map) return;

    const updatePitchState = () => {
      const pitch = map.getPitch();
      setIs3DView(pitch > 30);
    };

    map.on('pitch', updatePitchState);
    
    return () => {
      map.off('pitch', updatePitchState);
    };
  }, [map]);

  const handleStyleChange = (styleId: string) => {
    if (!map) return;
    
    map.setStyle(`mapbox://styles/mapbox/${styleId}`);
    setCurrentStyle(styleId);
    setIsStyleMenuOpen(false);
    
    // Re-apply terrain after style change
    map.once('style.load', () => {
      if (isTerrainEnabled) {
        enableTerrain();
      }
    });
  };

  const enableTerrain = () => {
    if (!map) return;
    
    // Add terrain source if not exists
    if (!map.getSource('mapbox-dem')) {
      map.addSource('mapbox-dem', {
        type: 'raster-dem',
        url: 'mapbox://mapbox.mapbox-terrain-dem-v1',
        tileSize: 512,
        maxzoom: 14
      });
    }
    
    // Set terrain
    map.setTerrain({ 
      source: 'mapbox-dem', 
      exaggeration: 1.5 
    });
    
    // Add sky layer if not exists
    if (!map.getLayer('sky')) {
      map.addLayer({
        id: 'sky',
        type: 'sky',
        paint: {
          'sky-type': 'atmosphere',
          'sky-atmosphere-sun': [0.0, 90.0],
          'sky-atmosphere-sun-intensity': 15
        }
      });
    }
  };

  const toggleTerrain = () => {
    if (!map) return;
    
    if (isTerrainEnabled) {
      // Disable terrain
      map.setTerrain(null);
      // Remove sky layer
      if (map.getLayer('sky')) {
        map.removeLayer('sky');
      }
    } else {
      // Enable terrain
      enableTerrain();
    }
    
    setIsTerrainEnabled(!isTerrainEnabled);
  };

  const toggle3DView = () => {
    if (!map) return;
    
    if (is3DView) {
      // Return to 2D view
      map.easeTo({
        pitch: 0,
        bearing: 0,
        duration: 1000
      });
    } else {
      // Enable 3D view
      // First ensure terrain is enabled
      if (!isTerrainEnabled) {
        enableTerrain();
        setIsTerrainEnabled(true);
      }
      
      // Animate to 3D perspective
      map.easeTo({
        pitch: 60,
        bearing: -20,
        duration: 1000
      });
    }
    
    setIs3DView(!is3DView);
  };

  return (
    <div className="absolute bottom-10 left-4 z-10">
      <div className="flex flex-col gap-2">
        {/* 3D View Toggle Button */}
        <button
          onClick={toggle3DView}
          className={`p-3 rounded-map border shadow-map-control hover:shadow-medium transition-all ${
            is3DView 
              ? 'bg-forest-100 border-forest-300 hover:bg-forest-200' 
              : 'bg-terrain-100 border-terrain-300 hover:bg-terrain-200'
          }`}
          title={is3DView ? 'Return to 2D view' : 'Switch to 3D view'}
        >
          <Box className={`w-5 h-5 ${
            is3DView ? 'text-forest-700' : 'text-trail-700'
          }`} />
        </button>

        {/* Terrain Toggle Button */}
        <button
          onClick={toggleTerrain}
          className={`p-3 rounded-map border shadow-map-control hover:shadow-medium transition-all ${
            isTerrainEnabled 
              ? 'bg-forest-100 border-forest-300 hover:bg-forest-200' 
              : 'bg-terrain-100 border-terrain-300 hover:bg-terrain-200'
          }`}
          title={isTerrainEnabled ? 'Disable terrain' : 'Enable terrain'}
        >
          <Mountain className={`w-5 h-5 ${
            isTerrainEnabled ? 'text-forest-700' : 'text-trail-700'
          }`} />
        </button>

        {/* Style Selector */}
        <div className="relative">
          <button
            onClick={() => setIsStyleMenuOpen(!isStyleMenuOpen)}
            className="bg-terrain-100 p-3 rounded-map border border-terrain-300 shadow-map-control hover:shadow-medium transition-all hover:bg-terrain-200"
            title="Change map style"
          >
            <Layers className="w-5 h-5 text-trail-700" />
          </button>

          {isStyleMenuOpen && (
            <div 
              className="absolute bottom-full left-0 mb-2 bg-terrain-100 rounded-lg shadow-xl border border-terrain-300 py-2 min-w-[150px]"
              style={{ backgroundColor: '#faf8f5' }}>
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
    </div>
  );
};