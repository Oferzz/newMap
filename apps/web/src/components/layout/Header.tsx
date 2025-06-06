import React from 'react';
import { Search, Plus, User, Menu, MapPin, LogIn, LogOut } from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { NaturalLanguageSearchBar } from '../search/NaturalLanguageSearchBar';
import { MobileMenu } from './MobileMenu';
import { ViewTypeButtons } from '../common';
import { searchAllThunk } from '../../store/thunks/search.thunks';
import { setActivePanel, setViewType } from '../../store/slices/uiSlice';
import { SearchResult } from '../../types';
import { logout } from '../../store/slices/authSlice';

type ContentType = 'all' | 'trips' | 'places';

interface HeaderProps {
  showContentTypeButtons?: boolean;
  contentType?: ContentType;
  onContentTypeChange?: (type: ContentType) => void;
  heroMode?: boolean; // New prop for hero landing page styling
  onSearch?: (query: string, filters?: any) => void; // Optional search handler
  onSearchResultSelect?: (result: SearchResult) => void; // Optional result selection handler
}

export const Header: React.FC<HeaderProps> = ({ 
  showContentTypeButtons = false,
  contentType = 'all',
  onContentTypeChange,
  heroMode = false,
  onSearch,
  onSearchResultSelect
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const viewType = useAppSelector((state) => state.ui.viewType);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = React.useState(false);
  const [showProfileMenu, setShowProfileMenu] = React.useState(false);
  const profileMenuRef = React.useRef<HTMLDivElement>(null);
  
  // Show view type buttons on home page or explore page (but not in hero mode)
  const showViewTypeButtons = (location.pathname === '/' || location.pathname === '/explore') && !heroMode;

  // Close profile menu when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setShowProfileMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleCreateTrip = () => {
    // Allow trip creation in guest mode with local storage
    navigate('/trips/new');
  };

  const handleOpenTrips = () => {
    dispatch(setActivePanel('trips'));
  };

  const handleSearch = (query: string, filters?: any) => {
    if (onSearch) {
      onSearch(query, filters);
    } else {
      dispatch(searchAllThunk({ query, filters }));
    }
  };

  const handleSearchResultSelect = (result: SearchResult) => {
    if (onSearchResultSelect) {
      onSearchResultSelect(result);
    } else {
      dispatch({ type: 'ui/selectItem', payload: result });
      dispatch({ type: 'ui/setActivePanel', payload: 'details' });
    }
  };

  const handleLogout = () => {
    dispatch(logout());
    navigate('/');
    setShowProfileMenu(false);
  };

  return (
    <header className={`fixed top-0 left-0 right-0 z-50 ${
      heroMode 
        ? 'bg-black/20 backdrop-blur-md' 
        : 'bg-white/60 backdrop-blur-lg border-b border-gray-200/50'
    }`}>
      <div className="h-16 px-4 flex items-center justify-between relative">
        {/* Logo */}
        <div className="flex items-center z-10">
          <button
            className={`md:hidden p-2 mr-2 ${
              heroMode 
                ? 'text-white hover:text-gray-200' 
                : 'text-gray-700 hover:text-gray-800'
            }`}
            onClick={() => setIsMobileMenuOpen(true)}
          >
            <Menu className="w-5 h-5" />
          </button>
          <a href="/" className="flex items-center">
            <img 
              src="/logo.svg" 
              alt="newMap" 
              className="h-12 w-auto drop-shadow-lg filter contrast-125 brightness-90"
              style={{ filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.3))' }}
            />
          </a>
        </div>

        {/* Search Bar - Absolutely centered */}
        <div className="hidden md:block absolute left-1/2 top-1/2 transform -translate-x-1/2 -translate-y-1/2 w-full max-w-2xl px-8">
          <NaturalLanguageSearchBar 
            onSearch={handleSearch}
            onResultSelect={handleSearchResultSelect}
          />
        </div>

        {/* Mobile Search Icon */}
        <button className={`md:hidden p-2 ${
          heroMode 
            ? 'text-white hover:text-gray-200' 
            : 'text-gray-700 hover:text-gray-800'
        }`}>
          <Search className="w-5 h-5" />
        </button>

        {/* Desktop Actions */}
        <div className="hidden md:flex items-center space-x-3 flex-shrink-0 z-10">
          {isAuthenticated ? (
            <>
              <button
                onClick={handleOpenTrips}
                className={`flex items-center px-3 py-2 rounded-lg transition-colors ${
                  heroMode 
                    ? 'text-white hover:text-gray-200 hover:bg-white/20' 
                    : 'text-gray-700 hover:text-gray-800 hover:bg-gray-200'
                }`}
              >
                <MapPin className="w-4 h-4 mr-2" />
                My Trips
              </button>
              
              <button
                onClick={handleCreateTrip}
                className={`flex items-center px-4 py-2 rounded-lg transition-colors ${
                  heroMode 
                    ? 'text-white hover:text-gray-200 hover:bg-white/20' 
                    : 'text-gray-800 hover:bg-gray-200'
                }`}
              >
                <Plus className="w-4 h-4 mr-2" />
                New Trip
              </button>
              
              {/* User Profile Dropdown */}
              <div className="relative" ref={profileMenuRef}>
                <button 
                  onClick={() => setShowProfileMenu(!showProfileMenu)}
                  className={`p-2 rounded-lg transition-colors ${
                    heroMode 
                      ? 'hover:bg-white/20' 
                      : 'hover:bg-gray-200'
                  }`}
                >
                  {user?.avatarUrl ? (
                    <img 
                      src={user.avatarUrl} 
                      alt={user.displayName}
                      className="w-8 h-8 rounded-full"
                    />
                  ) : (
                    <User className={`w-6 h-6 ${heroMode ? 'text-white' : 'text-gray-700'}`} />
                  )}
                </button>
                
                {showProfileMenu && (
                  <div className="absolute right-0 mt-2 w-48 bg-white/95 backdrop-blur-sm rounded-lg shadow-lg border border-gray-300 py-1">
                    <div className="px-4 py-2 border-b border-gray-300">
                      <p className="text-sm font-medium text-gray-800">{user?.displayName}</p>
                      <p className="text-xs text-gray-600">{user?.email}</p>
                    </div>
                    <button
                      onClick={() => {
                        navigate('/profile');
                        setShowProfileMenu(false);
                      }}
                      className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-200 transition-colors"
                    >
                      <User className="w-4 h-4 inline mr-2" />
                      Profile
                    </button>
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-200 transition-colors"
                    >
                      <LogOut className="w-4 h-4 inline mr-2" />
                      Logout
                    </button>
                  </div>
                )}
              </div>
            </>
          ) : (
            <button
              onClick={() => navigate('/login')}
              className={`flex items-center px-4 py-2 rounded-lg transition-colors ${
                heroMode 
                  ? 'text-white hover:text-gray-200 hover:bg-white/20' 
                  : 'text-gray-800 hover:bg-gray-200'
              }`}
            >
              <LogIn className="w-4 h-4 mr-2" />
              Login
            </button>
          )}
        </div>

        {/* Mobile User Avatar */}
        <div className="md:hidden">
          {user ? (
            <button className="p-1">
              <img 
                src={user.avatarUrl || '/default-avatar.png'} 
                alt={user.displayName}
                className="w-8 h-8 rounded-full"
              />
            </button>
          ) : (
            <button 
              onClick={() => navigate('/login')}
              className="p-2"
            >
              <User className={`w-5 h-5 ${heroMode ? 'text-white' : 'text-gray-700'}`} />
            </button>
          )}
        </div>
      </div>

      {/* View Type Buttons - Below search bar */}
      {showViewTypeButtons && (
        <div className="hidden md:flex justify-center gap-4 pb-1">
          <ViewTypeButtons
            activeView={viewType}
            onViewChange={(view) => dispatch(setViewType(view))}
          />
        </div>
      )}

      {/* Content Type Toggles - Part of header for explore page */}
      {showContentTypeButtons && (
        <div className="bg-white/60 backdrop-blur-lg -mt-1">
          <div className="flex justify-center py-0.5">
            <div className="flex items-center gap-6">
              {(['all', 'trips', 'places'] as ContentType[]).map((type) => (
                <button
                  key={type}
                  onClick={() => onContentTypeChange?.(type)}
                  className={`px-2 py-1 font-medium rounded-md transition-colors capitalize ${
                    contentType === type
                      ? 'text-gray-800 bg-gray-100'
                      : 'text-gray-600 hover:text-gray-800 hover:bg-gray-100'
                  }`}
                >
                  {type}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Mobile Menu */}
      <MobileMenu 
        isOpen={isMobileMenuOpen} 
        onClose={() => setIsMobileMenuOpen(false)} 
      />
    </header>
  );
};