import React from 'react';
import { Search, Plus, User, Menu, MapPin, LogIn, LogOut, Compass } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { SearchBar } from '../search/SearchBar';
import { MobileMenu } from './MobileMenu';
import { searchAllThunk } from '../../store/thunks/search.thunks';
import { setActivePanel } from '../../store/slices/uiSlice';
import { SearchResult } from '../../types';
import { logout } from '../../store/slices/authSlice';

type ContentType = 'all' | 'trips' | 'places';

interface HeaderProps {
  showContentTypeButtons?: boolean;
  contentType?: ContentType;
  onContentTypeChange?: (type: ContentType) => void;
}

export const Header: React.FC<HeaderProps> = ({ 
  showContentTypeButtons = false,
  contentType = 'all',
  onContentTypeChange 
}) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = React.useState(false);
  const [showProfileMenu, setShowProfileMenu] = React.useState(false);
  const profileMenuRef = React.useRef<HTMLDivElement>(null);

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
    dispatch(searchAllThunk({ query, filters }));
  };

  const handleSearchResultSelect = (result: SearchResult) => {
    dispatch({ type: 'ui/selectItem', payload: result });
    dispatch({ type: 'ui/setActivePanel', payload: 'details' });
  };

  const handleLogout = () => {
    dispatch(logout());
    navigate('/');
    setShowProfileMenu(false);
  };

  return (
    <header className={`fixed top-0 left-0 right-0 bg-terrain-100 z-50 shadow-soft border-b border-terrain-300 ${showContentTypeButtons ? '' : 'h-16'}`}>
      <div className="h-16 px-4 flex items-center justify-between relative">
        {/* Logo */}
        <div className="flex items-center z-10">
          <button
            className="md:hidden p-2 mr-2 text-trail-700 hover:text-trail-800"
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
        <div className="hidden md:block absolute left-1/2 top-1/2 transform -translate-x-1/2 -translate-y-1/2 w-full max-w-xl px-4">
          <SearchBar 
            onSearch={handleSearch}
            placeholder="Search places, trips, or users..."
            onResultSelect={handleSearchResultSelect}
          />
        </div>

        {/* Mobile Search Icon */}
        <button className="md:hidden p-2 text-trail-700 hover:text-trail-800">
          <Search className="w-5 h-5" />
        </button>

        {/* Desktop Actions */}
        <div className="hidden md:flex items-center space-x-3 flex-shrink-0 z-10">
          <button
            onClick={() => navigate('/explore')}
            className="flex items-center px-3 py-2 text-trail-700 hover:text-trail-800 hover:bg-terrain-200 rounded-lg transition-colors"
          >
            <Compass className="w-4 h-4 mr-2" />
            Explore
          </button>
          
          {isAuthenticated ? (
            <>
              <button
                onClick={handleOpenTrips}
                className="flex items-center px-3 py-2 text-trail-700 hover:text-trail-800 hover:bg-terrain-200 rounded-lg transition-colors"
              >
                <MapPin className="w-4 h-4 mr-2" />
                My Trips
              </button>
              
              <button
                onClick={handleCreateTrip}
                className="flex items-center px-4 py-2 text-trail-800 rounded-lg hover:bg-terrain-200 transition-colors"
              >
                <Plus className="w-4 h-4 mr-2" />
                New Trip
              </button>
              
              {/* User Profile Dropdown */}
              <div className="relative" ref={profileMenuRef}>
                <button 
                  onClick={() => setShowProfileMenu(!showProfileMenu)}
                  className="p-2 hover:bg-terrain-200 rounded-lg transition-colors"
                >
                  {user?.avatarUrl ? (
                    <img 
                      src={user.avatarUrl} 
                      alt={user.displayName}
                      className="w-8 h-8 rounded-full"
                    />
                  ) : (
                    <User className="w-6 h-6 text-trail-700" />
                  )}
                </button>
                
                {showProfileMenu && (
                  <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-terrain-300 py-1">
                    <div className="px-4 py-2 border-b border-terrain-300">
                      <p className="text-sm font-medium text-trail-800">{user?.displayName}</p>
                      <p className="text-xs text-trail-600">{user?.email}</p>
                    </div>
                    <button
                      onClick={() => {
                        navigate('/profile');
                        setShowProfileMenu(false);
                      }}
                      className="w-full text-left px-4 py-2 text-sm text-trail-700 hover:bg-terrain-200 transition-colors"
                    >
                      <User className="w-4 h-4 inline mr-2" />
                      Profile
                    </button>
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-2 text-sm text-trail-700 hover:bg-terrain-200 transition-colors"
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
              className="flex items-center px-4 py-2 text-trail-800 rounded-lg hover:bg-terrain-200 transition-colors"
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
              <User className="w-5 h-5" />
            </button>
          )}
        </div>
      </div>

      {/* Content Type Toggles - Part of header for explore page */}
      {showContentTypeButtons && (
        <div className="bg-terrain-100">
          <div className="flex justify-center pt-0.5 pb-0.5">
            <div className="flex items-center gap-6">
              {(['all', 'trips', 'places'] as ContentType[]).map((type) => (
                <button
                  key={type}
                  onClick={() => onContentTypeChange?.(type)}
                  className={`px-2 py-1 font-medium rounded-md transition-colors capitalize ${
                    contentType === type
                      ? 'text-trail-800 bg-terrain-200'
                      : 'text-trail-700 hover:text-trail-800 hover:bg-terrain-200'
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