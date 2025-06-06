import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch } from '../hooks/redux';
import { Header } from '../components/layout/Header';
import { HeroLanding } from '../components/hero/HeroLanding';
import { searchAllThunk } from '../store/thunks/search.thunks';
import { SearchResult } from '../types';

export const LandingPage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const handleSearch = (query: string, filters?: any) => {
    // Navigate to map view when search is performed
    navigate('/map');
    // Perform the search
    dispatch(searchAllThunk({ query, filters }));
  };

  const handleSearchResultSelect = (_result: SearchResult) => {
    // Navigate to map view and select result
    navigate('/map');
    // Handle result selection (this will be managed by the map view)
  };

  return (
    <div className="relative h-screen w-full overflow-hidden">
      {/* Header with hero mode styling and search handlers */}
      <Header 
        heroMode={true}
        onSearch={handleSearch}
        onSearchResultSelect={handleSearchResultSelect}
      />
      
      {/* Hero landing with full-screen photo */}
      <HeroLanding />
      
      {/* Optional: Subtle scroll indicator */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2 z-20">
        <div className="flex flex-col items-center text-white/70">
          <span className="text-sm font-medium mb-2">Start exploring</span>
          <div className="w-6 h-10 border-2 border-white/30 rounded-full flex justify-center">
            <div className="w-1 h-3 bg-white/50 rounded-full mt-2 animate-bounce"></div>
          </div>
        </div>
      </div>
    </div>
  );
};