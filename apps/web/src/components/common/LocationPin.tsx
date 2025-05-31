import React from 'react';

interface LocationPinProps {
  size?: number;
  color?: string;
  className?: string;
  onClick?: () => void;
}

export const LocationPin: React.FC<LocationPinProps> = ({ 
  size = 24, 
  color = '#EA4335', // Google's red color similar to the screenshot
  className = '',
  onClick
}) => {
  return (
    <svg
      width={size}
      height={size * 1.2} // Make it slightly taller for the pin shape
      viewBox="0 0 24 29"
      fill="none"
      className={className}
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Drop shadow */}
      <ellipse
        cx="12"
        cy="26"
        rx="3"
        ry="1.5"
        fill="rgba(0,0,0,0.2)"
      />
      
      {/* Main pin body */}
      <path
        d="M12 0C7.03 0 3 4.03 3 9c0 6.75 9 20 9 20s9-13.25 9-20c0-4.97-4.03-9-9-9z"
        fill={color}
        stroke="#fff"
        strokeWidth="0.5"
      />
      
      {/* Inner white circle */}
      <circle
        cx="12"
        cy="9"
        r="3.5"
        fill="white"
      />
      
      {/* Small center dot */}
      <circle
        cx="12"
        cy="9"
        r="1.5"
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
  default: '#EA4335',     // Google red
  restaurant: '#FF6B6B',  // Red
  hotel: '#4ECDC4',       // Teal  
  attraction: '#45B7D1',  // Blue
  shopping: '#96CEB4',    // Green
  transport: '#FECA57',   // Yellow
  selected: '#8E44AD',    // Purple
} as const;