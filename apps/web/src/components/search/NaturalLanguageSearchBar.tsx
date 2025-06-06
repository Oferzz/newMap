import React, { useState, useRef, useEffect } from 'react';
import { Search, X, Loader2, Sparkles, HelpCircle } from 'lucide-react';
import { useDebounce } from '../../hooks/useDebounce';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { clearSearch } from '../../store/slices/uiSlice';
import { SearchOverlay } from './SearchOverlay';
import { SearchResult } from '../../types';
import { nlpSearchService } from '../../services/nlpSearch.service';
import { setSearchResults, setIsSearching } from '../../store/slices/uiSlice';
import { setLoading, setError } from '../../store/slices/searchSlice';

interface NaturalLanguageSearchBarProps {
  onSearch: (query: string, filters?: any) => void;
  placeholder?: string;
  onResultSelect?: (result: SearchResult) => void;
}

export const NaturalLanguageSearchBar: React.FC<NaturalLanguageSearchBarProps> = ({ 
  onSearch, 
  placeholder = "Search places, trips, or activities...",
  onResultSelect
}) => {
  const dispatch = useAppDispatch();
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [queryUnderstanding, setQueryUnderstanding] = useState<any>(null);
  const [showHelp, setShowHelp] = useState(false);

  const isSearching = useAppSelector((state) => state.search.isLoading);
  const searchResults = useAppSelector((state) => state.ui.searchResults);
  const isSearchActive = useAppSelector((state) => state.ui.isSearching);
  const debouncedQuery = useDebounce(query, 300);
  const debouncedSuggestionQuery = useDebounce(query, 150);

  // Get search suggestions as user types
  useEffect(() => {
    if (debouncedSuggestionQuery.trim() && debouncedSuggestionQuery.length > 2) {
      nlpSearchService.getSuggestions(debouncedSuggestionQuery, 5)
        .then(response => {
          if (response.success && response.data) {
            setSuggestions(response.data);
            setShowSuggestions(true);
          }
        })
        .catch(error => {
          console.error('Error fetching suggestions:', error);
          setSuggestions([]);
        });
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  }, [debouncedSuggestionQuery]);

  // Trigger NLP search when debounced query changes
  useEffect(() => {
    if (debouncedQuery.trim()) {
      performNLPSearch(debouncedQuery);
    } else {
      dispatch(clearSearch());
      setQueryUnderstanding(null);
    }
  }, [debouncedQuery, dispatch]);

  const performNLPSearch = async (searchQuery: string) => {
    try {
      dispatch(setLoading(true));
      dispatch(setIsSearching(true));
      dispatch(setError(null));

      // Parse the query to show understanding
      const parseResponse = await nlpSearchService.parseQuery(searchQuery);
      if (parseResponse.success && parseResponse.data) {
        setQueryUnderstanding(parseResponse.data);
      }

      // Perform the actual search
      const searchResponse = await nlpSearchService.search({
        query: searchQuery,
        limit: 20
      });

      if (searchResponse.success && searchResponse.data) {
        // Transform NLP results to legacy format for compatibility
        const transformedResults = nlpSearchService.transformResultsToLegacyFormat(
          searchResponse.data.results
        );
        
        // Convert to proper types for SearchResults
        const places = transformedResults
          .filter(r => r.type === 'place')
          .map(r => ({
            id: r.id,
            name: r.name,
            description: r.description || '',
            category: 'general',
            location: r.coordinates ? {
              type: 'Point',
              coordinates: [r.coordinates.lng, r.coordinates.lat] as [number, number]
            } : undefined,
            address: '',
            city: '',
            state: '',
            country: '',
            postalCode: '',
            tags: [],
            images: [],
            createdBy: '',
            createdAt: new Date(),
            updatedAt: new Date(),
            type: 'poi' as const
          }));

        const trips = transformedResults
          .filter(r => r.type === 'trip')
          .map(r => ({
            id: r.id,
            title: r.name,
            description: r.description || '',
            owner_id: '',
            cover_image: '',
            privacy: 'public',
            status: 'active',
            timezone: '',
            tags: [],
            view_count: 0,
            share_count: 0,
            suggestion_count: 0
          }));

        const users = transformedResults.filter(r => r.type === 'user');
        
        dispatch(setSearchResults({
          places,
          trips,
          users
        }));

        // Call the legacy onSearch for backward compatibility
        onSearch(searchQuery, {
          nlp: true,
          understanding: parseResponse.data
        });
      }
    } catch (error: any) {
      console.error('NLP Search error:', error);
      dispatch(setError(error.message || 'Search failed'));
    } finally {
      dispatch(setLoading(false));
    }
  };

  const handleClear = () => {
    setQuery('');
    setSuggestions([]);
    setShowSuggestions(false);
    setQueryUnderstanding(null);
    inputRef.current?.focus();
    dispatch(clearSearch());
  };

  const handleResultSelect = (result: SearchResult) => {
    setQuery('');
    setSuggestions([]);
    setShowSuggestions(false);
    setQueryUnderstanding(null);
    dispatch(clearSearch());
    onResultSelect?.(result);
  };

  const handleCloseResults = () => {
    dispatch(clearSearch());
    setQueryUnderstanding(null);
  };

  const handleSuggestionSelect = (suggestion: string) => {
    setQuery(suggestion);
    setShowSuggestions(false);
    inputRef.current?.focus();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      if (showSuggestions) {
        setShowSuggestions(false);
      } else {
        handleClear();
      }
    }
  };

  const handleInputFocus = () => {
    if (suggestions.length > 0) {
      setShowSuggestions(true);
    }
  };

  const handleInputBlur = () => {
    // Delay hiding suggestions to allow clicking on them
    setTimeout(() => setShowSuggestions(false), 200);
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
          onFocus={handleInputFocus}
          onBlur={handleInputBlur}
          placeholder={placeholder}
          className="w-full pl-10 pr-16 py-2.5 bg-white border border-gray-300 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-400 text-gray-800 transition-all duration-200 shadow-sm hover:shadow-md"
        />

        {/* Action Buttons */}
        <div className="absolute right-3 flex items-center space-x-1">
          {query && (
            <button
              onClick={handleClear}
              className="p-1 hover:bg-gray-100 rounded-full transition-colors"
              title="Clear search"
            >
              <X className="w-4 h-4 text-gray-500" />
            </button>
          )}
          
          <button
            onClick={() => setShowHelp(!showHelp)}
            className={`p-1 hover:bg-gray-100 rounded-full transition-colors ${
              showHelp ? 'bg-gray-100' : ''
            }`}
            title="Search help"
          >
            <HelpCircle className="w-4 h-4 text-gray-500" />
          </button>
        </div>
      </div>

      {/* Query Understanding Display */}
      {queryUnderstanding && query.trim() && (
        <div className="absolute top-full mt-1 w-full bg-blue-50 border border-blue-200 rounded-lg p-3 z-40 shadow-sm">
          <div className="flex items-start space-x-2">
            <Sparkles className="w-4 h-4 text-blue-600 mt-0.5 flex-shrink-0" />
            <div className="flex-1 text-sm">
              <div className="text-blue-800 font-medium">
                I understand: {queryUnderstanding.intent}
              </div>
              {queryUnderstanding.explanation && (
                <div className="text-blue-600 mt-1">
                  {queryUnderstanding.explanation}
                </div>
              )}
              {queryUnderstanding.filters && Object.keys(queryUnderstanding.filters).length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {Object.entries(queryUnderstanding.filters).map(([key, value]) => (
                    <span 
                      key={key}
                      className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
                    >
                      {key}: {String(value)}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Search Suggestions */}
      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute top-full mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg z-50 max-h-60 overflow-y-auto">
          {suggestions.map((suggestion, index) => (
            <button
              key={index}
              onClick={() => handleSuggestionSelect(suggestion)}
              className="w-full px-4 py-3 text-left hover:bg-gray-50 flex items-center space-x-3 transition-colors"
            >
              <Search className="w-4 h-4 text-gray-400 flex-shrink-0" />
              <span className="text-gray-700">{suggestion}</span>
            </button>
          ))}
        </div>
      )}

      {/* Help Overlay */}
      {showHelp && (
        <div className="absolute top-full mt-2 w-full bg-white border border-gray-200 rounded-lg shadow-lg p-4 z-50">
          <h3 className="font-semibold text-gray-800 mb-3">Natural Language Search Help</h3>
          <div className="space-y-2 text-sm text-gray-600">
            <div>
              <strong>Examples:</strong>
            </div>
            <ul className="space-y-1 pl-4">
              <li>• "Easy hiking trails near San Francisco"</li>
              <li>• "Day trips with waterfalls in California"</li>
              <li>• "Camping spots for families with kids"</li>
              <li>• "Difficult mountain climbs in Colorado"</li>
              <li>• "Beach activities in Southern California"</li>
            </ul>
            <div className="mt-3 pt-3 border-t border-gray-200">
              <div className="text-xs text-gray-500">
                Powered by natural language processing - just describe what you're looking for!
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Search Results Overlay */}
      {isSearchActive && searchResults && query.trim() && !showSuggestions && (
        <SearchOverlay
          results={searchResults}
          onSelect={handleResultSelect}
          onClose={handleCloseResults}
        />
      )}
    </div>
  );
};