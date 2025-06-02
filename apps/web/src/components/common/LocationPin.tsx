import React from 'react';

interface LocationPinProps {
  size?: number;
  color?: string;
  className?: string;
  onClick?: () => void;
}

export const LocationPin: React.FC<LocationPinProps> = ({ 
  size = 24, 
  color = '#5e4c41', // Brown color matching the theme
  className = '',
  onClick
}) => {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      className={className}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Outer circle */}
      <circle
        cx="12"
        cy="12"
        r="11"
        fill="white"
        stroke={color}
        strokeWidth="2"
      />
      
      {/* Inner dot */}
      <circle
        cx="12"
        cy="12"
        r="4"
        fill={color}
      />
    </svg>
  );
};

// Alternative simplified version (closer to your screenshot)
export const SimpleLocationPin: React.FC<LocationPinProps> = ({ 
  size = 24, 
  color = '#EA4335',
  className = '',
  onClick
}) => {
  return (
    <svg
      width={size}
      height={size * 1.33}
      viewBox="0 0 18 24"
      fill="none"
      className={className}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Main pin shape */}
      <path
        d="M9 0C4.03 0 0 4.03 0 9c0 6.75 9 15 9 15s9-8.25 9-15c0-4.97-4.03-9-9-9z"
        fill={color}
      />
      
      {/* White center circle */}
      <circle
        cx="9"
        cy="9"
        r="4"
        fill="white"
      />
    </svg>
  );
};

// Define color constants
export const LocationPinColors = {
  default: '#5e4c41',     // Trail brown (matching theme)
  restaurant: '#d4b5a0',  // Terrain
  hotel: '#6a9ec9',       // Water blue  
  attraction: '#8fbe7e',  // Forest green
  shopping: '#fda328',    // Road orange
  transport: '#927b67',   // Trail light
  selected: '#446b8e',    // Water dark
} as const;