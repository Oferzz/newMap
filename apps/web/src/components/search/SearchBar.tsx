import React, { useState, useRef, useEffect } from 'react';
import { Search, X, Filter, Loader2 } from 'lucide-react';
import { useDebounce } from '../../hooks/useDebounce';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { searchAll } from '../../store/slices/searchSlice';

interface SearchBarProps {
  onSearch: (query: string) => void;
  placeholder?: string;
}

export const SearchBar: React.FC<SearchBarProps> = ({ 
  onSearch, 
  placeholder = "Search places, trips, or users..." 
}) => {
  const dispatch = useAppDispatch();
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState({
    type: 'all', // all, places, trips, users
    radius: 10, // km
    onlyMine: false,
  });

  const isSearching = useAppSelector((state) => state.search.isLoading);
  const debouncedQuery = useDebounce(query, 300);

  // Trigger search when debounced query changes
  useEffect(() => {
    if (debouncedQuery.trim()) {
      dispatch(searchAll({ 
        query: debouncedQuery, 
        filters 
      }));
      onSearch(debouncedQuery);
    } else {
      dispatch({ type: 'search/clearResults' });
    }
  }, [debouncedQuery, filters, dispatch, onSearch]);

  const handleClear = () => {
    setQuery('');
    inputRef.current?.focus();
    dispatch({ type: 'search/clearResults' });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleClear();
    }
  };

  return (
    <div className="relative w-full">
      <div className="relative flex items-center">
        {/* Search Icon */}
        <div className="absolute left-3 pointer-events-none">
          {isSearching ? (
            <Loader2 className="w-5 h-5 text-gray-400 animate-spin" />
          ) : (
            <Search className="w-5 h-5 text-gray-400" />
          )}
        </div>

        {/* Input */}
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="w-full pl-10 pr-20 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />

        {/* Action Buttons */}
        <div className="absolute right-2 flex items-center space-x-1">
          {query && (
            <button
              onClick={handleClear}
              className="p-1 hover:bg-gray-100 rounded"
            >
              <X className="w-4 h-4 text-gray-500" />
            </button>
          )}
          
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`p-1 hover:bg-gray-100 rounded ${
              showFilters ? 'bg-gray-100' : ''
            }`}
          >
            <Filter className="w-4 h-4 text-gray-500" />
          </button>
        </div>
      </div>

      {/* Filter Dropdown */}
      {showFilters && (
        <div className="absolute top-full mt-2 w-full bg-white rounded-lg shadow-lg border border-gray-200 p-4 z-50">
          <div className="space-y-4">
            {/* Search Type */}
            <div>
              <label className="text-sm font-medium text-gray-700">
                Search in
              </label>
              <div className="mt-2 grid grid-cols-4 gap-2">
                {['all', 'places', 'trips', 'users'].map((type) => (
                  <button
                    key={type}
                    onClick={() => setFilters({ ...filters, type })}
                    className={`px-3 py-1 text-sm rounded-md capitalize ${
                      filters.type === type
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {type}
                  </button>
                ))}
              </div>
            </div>

            {/* Radius Filter (for places) */}
            {(filters.type === 'all' || filters.type === 'places') && (
              <div>
                <label className="text-sm font-medium text-gray-700">
                  Search radius: {filters.radius} km
                </label>
                <input
                  type="range"
                  min="1"
                  max="50"
                  value={filters.radius}
                  onChange={(e) => 
                    setFilters({ ...filters, radius: parseInt(e.target.value) })
                  }
                  className="mt-2 w-full"
                />
              </div>
            )}

            {/* Only Mine Filter */}
            <div className="flex items-center">
              <input
                type="checkbox"
                id="onlyMine"
                checked={filters.onlyMine}
                onChange={(e) => 
                  setFilters({ ...filters, onlyMine: e.target.checked })
                }
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label 
                htmlFor="onlyMine" 
                className="ml-2 text-sm text-gray-700"
              >
                Only show my content
              </label>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};