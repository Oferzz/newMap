import React, { useEffect } from 'react';
import { X, Plus, MapPin, Route, User, Settings, LogOut } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { logout } from '../../store/slices/authSlice';

interface MobileMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

export const MobileMenu: React.FC<MobileMenuProps> = ({ isOpen, onClose }) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);

  // Prevent body scroll when menu is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  const handleNavigation = (path: string) => {
    navigate(path);
    onClose();
  };

  const handleLogout = () => {
    dispatch(logout());
    onClose();
    navigate('/');
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 z-50"
        onClick={onClose}
      />
      
      {/* Menu Panel */}
      <div className="fixed left-0 top-0 bottom-0 w-80 max-w-[85vw] bg-white z-50 shadow-xl">
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b">
            <h2 className="text-lg font-semibold">Menu</h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* User Info */}
          {user && (
            <div className="p-4 border-b">
              <div className="flex items-center">
                <img
                  src={user.avatar_url || '/default-avatar.png'}
                  alt={user.display_name}
                  className="w-12 h-12 rounded-full mr-3"
                />
                <div>
                  <div className="font-medium">{user.display_name}</div>
                  <div className="text-sm text-gray-600">@{user.username}</div>
                </div>
              </div>
            </div>
          )}

          {/* Navigation Items */}
          <nav className="flex-1 overflow-y-auto">
            <div className="p-2">
              {/* Create Trip */}
              <button
                onClick={() => handleNavigation('/trips/new')}
                className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg"
              >
                <Plus className="w-5 h-5 mr-3 text-blue-600" />
                <span className="font-medium">Create New Trip</span>
              </button>

              {/* My Trips */}
              {user && (
                <button
                  onClick={() => handleNavigation('/trips')}
                  className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg"
                >
                  <Route className="w-5 h-5 mr-3 text-gray-600" />
                  <span>My Trips</span>
                </button>
              )}

              {/* Explore Places */}
              <button
                onClick={() => handleNavigation('/places')}
                className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg"
              >
                <MapPin className="w-5 h-5 mr-3 text-gray-600" />
                <span>Explore Places</span>
              </button>

              {/* Profile */}
              {user && (
                <>
                  <button
                    onClick={() => handleNavigation('/profile')}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg"
                  >
                    <User className="w-5 h-5 mr-3 text-gray-600" />
                    <span>Profile</span>
                  </button>

                  <button
                    onClick={() => handleNavigation('/settings')}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg"
                  >
                    <Settings className="w-5 h-5 mr-3 text-gray-600" />
                    <span>Settings</span>
                  </button>
                </>
              )}
            </div>
          </nav>

          {/* Footer Actions */}
          <div className="p-4 border-t">
            {user ? (
              <button
                onClick={handleLogout}
                className="w-full flex items-center px-4 py-3 text-left hover:bg-gray-50 rounded-lg text-red-600"
              >
                <LogOut className="w-5 h-5 mr-3" />
                <span>Log Out</span>
              </button>
            ) : (
              <button
                onClick={() => handleNavigation('/login')}
                className="w-full flex items-center justify-center px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Sign In
              </button>
            )}
          </div>
        </div>
      </div>
    </>
  );
};