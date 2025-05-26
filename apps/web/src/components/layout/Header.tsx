import React from 'react';
import { Search, Plus, Bell, User, Menu } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { SearchBar } from '../search/SearchBar';
import { UserMenu } from '../user/UserMenu';
import { NotificationBell } from '../notifications/NotificationBell';
import { MobileMenu } from './MobileMenu';
import { Button } from '../ui/Button';

export const Header: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = React.useState(false);

  const handleCreateTrip = () => {
    if (!user) {
      navigate('/login');
      return;
    }
    navigate('/trips/new');
  };

  const handleSearch = (query: string) => {
    dispatch({ type: 'search/setQuery', payload: query });
  };

  return (
    <header className="fixed top-0 left-0 right-0 h-16 bg-white border-b border-gray-200 z-50">
      <div className="h-full px-4 flex items-center justify-between">
        {/* Logo */}
        <div className="flex items-center">
          <button
            className="md:hidden p-2 mr-2"
            onClick={() => setIsMobileMenuOpen(true)}
          >
            <Menu className="w-5 h-5" />
          </button>
          <a href="/" className="flex items-center">
            <img 
              src="/logo.svg" 
              alt="TripPlanner" 
              className="h-8 w-auto"
            />
            <span className="hidden sm:block ml-2 font-semibold text-lg">
              TripPlanner
            </span>
          </a>
        </div>

        {/* Search Bar - Hidden on mobile */}
        <div className="hidden md:flex flex-1 max-w-2xl mx-4">
          <SearchBar 
            onSearch={handleSearch}
            placeholder="Search places, trips, or users..."
          />
        </div>

        {/* Mobile Search Icon */}
        <button className="md:hidden p-2">
          <Search className="w-5 h-5" />
        </button>

        {/* Desktop Actions */}
        <div className="hidden md:flex items-center space-x-4">
          <Button
            onClick={handleCreateTrip}
            variant="primary"
            className="flex items-center"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Trip
          </Button>
          
          {user && <NotificationBell />}
          
          <UserMenu />
        </div>

        {/* Mobile User Avatar */}
        <div className="md:hidden">
          {user ? (
            <button className="p-1">
              <img 
                src={user.avatar_url || '/default-avatar.png'} 
                alt={user.display_name}
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

      {/* Mobile Menu */}
      <MobileMenu 
        isOpen={isMobileMenuOpen} 
        onClose={() => setIsMobileMenuOpen(false)} 
      />
    </header>
  );
};