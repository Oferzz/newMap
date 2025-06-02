import React, { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  Calendar, 
  MapPin, 
  Users, 
  Globe,
  Lock,
  Clock,
  Star,
  Compass,
  User,
  Grid,
  List
} from 'lucide-react';
import { getUserTripsThunk } from '../store/thunks/trips.thunks';
import { Trip as ReduxTrip } from '../store/slices/tripsSlice';
import { Place, Trip as APITrip } from '../types';
import { format, differenceInDays } from 'date-fns';
import { api } from '../services/api';
import { clearSearch as clearSearchAction } from '../store/slices/searchSlice';
import { clearSearch as clearUISearch } from '../store/slices/uiSlice';

type ContentType = 'all' | 'trips' | 'places';
type FilterType = 'all' | 'saved';
type ViewMode = 'grid' | 'list';

interface ExplorePageProps {
  contentType?: ContentType;
  onContentTypeChange?: (type: ContentType) => void;
}

export const ExplorePage: React.FC<ExplorePageProps> = ({ 
  contentType: propContentType = 'all'
}) => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { items: userTrips } = useAppSelector(state => state.trips);
  const { items: userPlaces } = useAppSelector(state => state.places);
  const { isAuthenticated } = useAppSelector(state => state.auth);
  
  // Get search query from Redux search state
  const searchQuery = useAppSelector(state => state.search.query);
  const searchResults = useAppSelector(state => state.ui.searchResults);
  
  // Clear search when leaving the explore page
  useEffect(() => {
    return () => {
      if (location.pathname !== '/explore') {
        // Clear search when navigating away from explore page
        dispatch(clearSearchAction());
        dispatch(clearUISearch());
      }
    };
  }, [location.pathname, dispatch]);

  // Use contentType from props instead of local state
  const contentType = propContentType;
  const [filterType, setFilterType] = useState<FilterType>('all');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [isLoading, setIsLoading] = useState(false);
  const [publicTrips, setPublicTrips] = useState<APITrip[]>([]);
  const [publicPlaces, setPublicPlaces] = useState<Place[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  // Load data on component mount and when filters change
  useEffect(() => {
    loadData();
  }, [contentType, filterType]);

  // Load user's saved content if authenticated
  useEffect(() => {
    if (isAuthenticated && filterType === 'saved') {
      dispatch(getUserTripsThunk());
    }
  }, [isAuthenticated, filterType, dispatch]);

  // Use search results from Redux when available
  useEffect(() => {
    if (searchResults && searchQuery) {
      // Search results from header search are already filtered
      setPublicTrips(searchResults.trips || []);
      setPublicPlaces(searchResults.places || []);
    } else if (!searchQuery) {
      // Load default content when no search query
      loadData();
    }
  }, [searchResults, searchQuery]);

  const loadData = async () => {
    setIsLoading(true);
    setPage(1);
    
    try {
      if (filterType === 'all') {
        // Load public content
        if (contentType === 'trips' || contentType === 'all') {
          await loadPublicTrips(1, true);
        }
        if (contentType === 'places' || contentType === 'all') {
          await loadPublicPlaces(1, true);
        }
      }
      // For 'saved' filter, we rely on Redux state which is loaded in useEffect above
    } catch (error) {
      console.error('Failed to load explore data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadPublicTrips = async (pageNum: number, reset: boolean = false) => {
    try {
      const params = new URLSearchParams({
        page: pageNum.toString(),
        limit: '20',
        privacy: 'public',
      });
      
      if (searchQuery) {
        params.append('q', searchQuery);
      }

      const response = await api.get<{data: APITrip[], meta: {hasMore: boolean}}>(`/trips?${params.toString()}`, { skipAuth: true });
      
      if (reset) {
        setPublicTrips(response.data || []);
      } else {
        setPublicTrips(prev => [...prev, ...(response.data || [])]);
      }
      
      setHasMore(response.meta?.hasMore || false);
    } catch (error) {
      console.error('Failed to load public trips:', error);
      setPublicTrips([]);
    }
  };

  const loadPublicPlaces = async (pageNum: number, reset: boolean = false) => {
    try {
      if (!searchQuery && filterType === 'all') {
        // For public places without search, we need a search query
        setPublicPlaces([]);
        return;
      }

      const params = new URLSearchParams({
        page: pageNum.toString(),
        limit: '20',
      });
      
      if (searchQuery) {
        params.append('q', searchQuery);
      }

      const response = await api.get<{data: Place[], meta: {hasMore: boolean}}>(`/places/search?${params.toString()}`, { skipAuth: true });
      
      if (reset) {
        setPublicPlaces(response.data || []);
      } else {
        setPublicPlaces(prev => [...prev, ...(response.data || [])]);
      }
      
      setHasMore(response.meta?.hasMore || false);
    } catch (error) {
      console.error('Failed to load public places:', error);
      setPublicPlaces([]);
    }
  };

  const loadMore = async () => {
    if (!hasMore || isLoading) return;
    
    const nextPage = page + 1;
    setPage(nextPage);
    
    if (filterType === 'all') {
      if (contentType === 'trips' || contentType === 'all') {
        await loadPublicTrips(nextPage, false);
      }
      if (contentType === 'places' || contentType === 'all') {
        await loadPublicPlaces(nextPage, false);
      }
    }
  };

  // Get filtered data based on current settings
  const getFilteredData = () => {
    let trips: (ReduxTrip | APITrip)[] = [];
    let places: Place[] = [];

    if (filterType === 'saved') {
      trips = userTrips;
      places = userPlaces;
      
      // Apply search filter for saved content if search query exists
      if (searchQuery) {
        trips = trips.filter(trip => 
          trip.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
          trip.description?.toLowerCase().includes(searchQuery.toLowerCase())
        );
        places = places.filter(place => 
          place.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
          place.description?.toLowerCase().includes(searchQuery.toLowerCase())
        );
      }
    } else {
      // For public content, use search results if available, otherwise use loaded public content
      trips = searchResults && searchQuery ? (searchResults.trips || []) : publicTrips;
      places = searchResults && searchQuery ? (searchResults.places || []) : publicPlaces;
    }

    return { trips, places };
  };

  const { trips, places } = getFilteredData();

  const getTripIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <MapPin className="w-4 h-4 text-green-600" />;
      case 'completed':
        return <Calendar className="w-4 h-4 text-blue-600" />;
      case 'planning':
      default:
        return <Clock className="w-4 h-4 text-orange-600" />;
    }
  };

  const getPrivacyIcon = (privacy: string) => {
    switch (privacy) {
      case 'public':
        return <Globe className="w-4 h-4 text-gray-500" />;
      case 'friends':
        return <Users className="w-4 h-4 text-gray-500" />;
      case 'private':
      default:
        return <Lock className="w-4 h-4 text-gray-500" />;
    }
  };

  const getTripDuration = (trip: ReduxTrip | APITrip) => {
    const startDate = trip.startDate || (trip as any).start_date;
    const endDate = trip.endDate || (trip as any).end_date;
    
    if (!startDate || !endDate) return 'Duration unknown';
    
    const days = differenceInDays(new Date(endDate), new Date(startDate)) + 1;
    return `${days} day${days > 1 ? 's' : ''}`;
  };

  const handleTripClick = (trip: ReduxTrip | APITrip) => {
    navigate(`/trips/${trip.id}`);
  };

  const handlePlaceClick = (place: Place) => {
    navigate(`/places/${place.id}`);
  };

  const renderTripCard = (trip: ReduxTrip | APITrip) => (
    <div
      key={trip.id}
      className={`bg-white rounded-lg shadow-md overflow-hidden cursor-pointer hover:shadow-lg transition-shadow ${
        viewMode === 'list' ? 'flex' : ''
      }`}
      onClick={() => handleTripClick(trip)}
    >
      {(trip as any).cover_image || (trip as any).coverImage ? (
        <div className={`${viewMode === 'list' ? 'w-32 h-24' : 'h-48'} bg-gray-200 relative`}>
          <img
            src={(trip as any).cover_image || (trip as any).coverImage}
            alt={trip.title}
            className="w-full h-full object-cover"
          />
          <div className="absolute top-2 right-2 flex gap-1">
            {getTripIcon(trip.status)}
            {getPrivacyIcon(trip.privacy)}
          </div>
        </div>
      ) : (
        <div className={`${viewMode === 'list' ? 'w-32 h-24' : 'h-48'} bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center relative`}>
          <MapPin className="w-12 h-12 text-white" />
          <div className="absolute top-2 right-2 flex gap-1">
            {getTripIcon(trip.status)}
            {getPrivacyIcon(trip.privacy)}
          </div>
        </div>
      )}
      
      <div className="p-4 flex-1">
        <h3 className="font-semibold text-gray-900 mb-2 line-clamp-1">
          {trip.title}
        </h3>
        <p className="text-sm text-gray-600 mb-3 line-clamp-2">
          {trip.description}
        </p>
        
        <div className="flex items-center gap-4 text-xs text-gray-500">
          {(trip as any).start_date || trip.startDate ? (
            <div className="flex items-center gap-1">
              <Calendar className="w-3 h-3" />
              <span>{format(new Date((trip as any).start_date || trip.startDate), 'MMM d, yyyy')}</span>
            </div>
          ) : null}
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            <span>{getTripDuration(trip)}</span>
          </div>
          {trip.waypoints && trip.waypoints.length > 0 && (
            <div className="flex items-center gap-1">
              <MapPin className="w-3 h-3" />
              <span>{trip.waypoints.length} places</span>
            </div>
          )}
        </div>

        {trip.collaborators && trip.collaborators.length > 1 && (
          <div className="flex items-center gap-1 mt-2">
            <Users className="w-3 h-3 text-gray-400" />
            <span className="text-xs text-gray-500">
              {trip.collaborators.length} collaborators
            </span>
          </div>
        )}
      </div>
    </div>
  );

  const renderPlaceCard = (place: Place) => (
    <div
      key={place.id}
      className={`bg-white rounded-lg shadow-md overflow-hidden cursor-pointer hover:shadow-lg transition-shadow ${
        viewMode === 'list' ? 'flex' : ''
      }`}
      onClick={() => handlePlaceClick(place)}
    >
      {place.photos && place.photos.length > 0 ? (
        <div className={`${viewMode === 'list' ? 'w-32 h-24' : 'h-48'} bg-gray-200`}>
          <img
            src={place.photos[0]}
            alt={place.name}
            className="w-full h-full object-cover"
          />
        </div>
      ) : (
        <div className={`${viewMode === 'list' ? 'w-32 h-24' : 'h-48'} bg-gradient-to-br from-green-500 to-blue-600 flex items-center justify-center`}>
          <MapPin className="w-12 h-12 text-white" />
        </div>
      )}
      
      <div className="p-4 flex-1">
        <h3 className="font-semibold text-gray-900 mb-2 line-clamp-1">
          {place.name}
        </h3>
        <p className="text-sm text-gray-600 mb-3 line-clamp-2">
          {place.description}
        </p>
        
        <div className="flex items-center gap-4 text-xs text-gray-500">
          {place.category && (
            <div className="flex items-center gap-1">
              <MapPin className="w-3 h-3" />
              <span className="capitalize">
                {Array.isArray(place.category) ? place.category[0] : place.category}
              </span>
            </div>
          )}
          {place.average_rating && (
            <div className="flex items-center gap-1">
              <Star className="w-3 h-3 text-yellow-500" />
              <span>{place.average_rating.toFixed(1)}</span>
            </div>
          )}
          {place.city && (
            <span>{place.city}</span>
          )}
        </div>

        {place.tags && place.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mt-2">
            {place.tags.slice(0, 3).map((tag, index) => (
              <span
                key={index}
                className="px-2 py-1 bg-gray-100 text-xs text-gray-600 rounded"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );

  const shouldShowContent = (type: ContentType) => {
    return contentType === 'all' || contentType === type;
  };

  return (
    <div className="min-h-screen bg-gray-50 pt-24">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-4">
            <Compass className="w-8 h-8 text-indigo-600" />
            <h1 className="text-3xl font-bold text-gray-900">Explore</h1>
          </div>
          <p className="text-gray-600 max-w-2xl">
            Discover amazing trips and places from our community. Find inspiration for your next adventure.
          </p>
        </div>

        {/* Controls */}
        <div className="bg-white rounded-lg shadow-sm border p-6 mb-8">
          <div className="flex flex-wrap gap-4 items-center">
            {/* Filter Toggle */}
            <div className="flex bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setFilterType('all')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  filterType === 'all'
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'text-gray-700 hover:text-gray-900'
                }`}
              >
                <Globe className="w-4 h-4 inline mr-2" />
                Public
              </button>
              {isAuthenticated && (
                <button
                  onClick={() => setFilterType('saved')}
                  className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                    filterType === 'saved'
                      ? 'bg-indigo-600 text-white shadow-sm'
                      : 'text-gray-700 hover:text-gray-900'
                  }`}
                >
                  <User className="w-4 h-4 inline mr-2" />
                  My Saved
                </button>
              )}
            </div>

            {/* Search instruction */}
            <div className="flex-1 text-center">
              <p className="text-sm text-gray-500">
                {searchQuery ? `Showing results for "${searchQuery}"` : 'Use the search bar above to find specific content'}
              </p>
            </div>

            {/* View Mode Toggle */}
            <div className="flex bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-md transition-colors ${
                  viewMode === 'grid'
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'text-gray-700 hover:text-gray-900'
                }`}
              >
                <Grid className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`p-2 rounded-md transition-colors ${
                  viewMode === 'list'
                    ? 'bg-indigo-600 text-white shadow-sm'
                    : 'text-gray-700 hover:text-gray-900'
                }`}
              >
                <List className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
          </div>
        ) : (
          <div className="space-y-8">
            {/* Trips Section */}
            {shouldShowContent('trips') && trips.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                  <MapPin className="w-5 h-5 text-indigo-600" />
                  Trips ({trips.length})
                </h2>
                <div className={`grid gap-6 ${
                  viewMode === 'grid' 
                    ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4' 
                    : 'grid-cols-1'
                }`}>
                  {trips.map(renderTripCard)}
                </div>
              </div>
            )}

            {/* Places Section */}
            {shouldShowContent('places') && places.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold text-gray-900 mb-4 flex items-center gap-2">
                  <MapPin className="w-5 h-5 text-green-600" />
                  Places ({places.length})
                </h2>
                <div className={`grid gap-6 ${
                  viewMode === 'grid' 
                    ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4' 
                    : 'grid-cols-1'
                }`}>
                  {places.map(renderPlaceCard)}
                </div>
              </div>
            )}

            {/* Empty States */}
            {!isLoading && trips.length === 0 && places.length === 0 && (
              <div className="text-center py-16">
                <Compass className="w-16 h-16 mx-auto mb-4 text-gray-300" />
                <h3 className="text-xl font-semibold text-gray-700 mb-2">
                  {searchQuery ? 'No results found' : 'Nothing to explore yet'}
                </h3>
                <p className="text-gray-500 mb-6">
                  {searchQuery 
                    ? 'Try adjusting your search query or filters'
                    : filterType === 'saved' 
                      ? "You haven't saved any content yet. Start exploring to save your favorites!"
                      : 'No public content available at the moment.'
                  }
                </p>
                {filterType === 'saved' && (
                  <button
                    onClick={() => setFilterType('all')}
                    className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
                  >
                    Explore Public Content
                  </button>
                )}
              </div>
            )}

            {/* Load More */}
            {hasMore && (trips.length > 0 || places.length > 0) && (
              <div className="text-center py-8">
                <button
                  onClick={loadMore}
                  disabled={isLoading}
                  className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isLoading ? 'Loading...' : 'Load More'}
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};