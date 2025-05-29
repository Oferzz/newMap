/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Primary colors inspired by outdoors map
        terrain: {
          50: '#fdfcfb',
          100: '#faf8f5',
          200: '#f8f4f0', // Main background like map
          300: '#f0e8df',
          400: '#e6d4c1',
          500: '#d4b5a0',
          600: '#b89478',
          700: '#9a7560',
          800: '#7d5d4b',
          900: '#654a3b',
        },
        // Water colors from map
        water: {
          50: '#f0f7fc',
          100: '#e0eff8',
          200: '#c2dff1',
          300: '#aec9e3', // Map water color
          400: '#8bb5d8',
          500: '#6a9ec9',
          600: '#5385b0',
          700: '#446b8e',
          800: '#385670',
          900: '#2f455a',
        },
        // Forest/park greens from map
        forest: {
          50: '#f4f9f3',
          100: '#e8f3e6',
          200: '#d1e7cd',
          300: '#c8e3c5', // Light parks
          400: '#acd19d', // Deeper forests
          500: '#8fbe7e',
          600: '#70a55e',
          700: '#5a8a4a',
          800: '#476d3b',
          900: '#3a5831',
        },
        // Trail/path browns
        trail: {
          50: '#faf9f7',
          100: '#f3f0eb',
          200: '#e7dfd6',
          300: '#d5c7b8', // Contour lines
          400: '#c7b8a8',
          500: '#b09a86',
          600: '#927b67',
          700: '#755f50',
          800: '#5e4c41',
          900: '#4d3e36',
        },
        // Road colors
        road: {
          primary: '#fda328', // Orange highways
          secondary: '#ffffff', // White roads
          tertiary: '#e6d4c1', // Tan local roads
        }
      },
      fontFamily: {
        'map': ['"DIN Pro"', '"Arial Unicode MS"', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        'map-control': '0 0 0 2px rgba(0,0,0,.1)',
        'soft': '0 1px 4px rgba(0,0,0,.08)',
        'medium': '0 2px 8px rgba(0,0,0,.12)',
      },
      borderRadius: {
        'map': '4px',
      }
    },
  },
  plugins: [],
}