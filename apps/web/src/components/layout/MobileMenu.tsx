import React, { useEffect } from 'react';
import { X, Plus, MapPin, User, Settings, LogOut, LogIn, Compass } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { logout } from '../../store/slices/authSlice';
import { setActivePanel } from '../../store/slices/uiSlice';

interface MobileMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

export const MobileMenu: React.FC<MobileMenuProps> = ({ isOpen, onClose }) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);

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
      <div className="fixed left-0 top-0 bottom-0 w-80 max-w-[85vw] bg-terrain-100 z-50 shadow-xl">
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-terrain-300">
            <h2 className="text-lg font-semibold text-trail-800">Menu</h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-terrain-200 rounded-lg"
            >
              <X className="w-5 h-5 text-trail-700" />
            </button>
          </div>

          {/* User Info */}
          {user && (
            <div className="p-4 border-b border-terrain-300">
              <div className="flex items-center">
                {user.avatarUrl ? (
                  <img
                    src={user.avatarUrl}
                    alt={user.displayName}
                    className="w-12 h-12 rounded-full mr-3"
                  />
                ) : (
                  <div className="w-12 h-12 rounded-full bg-terrain-200 flex items-center justify-center mr-3">
                    <User className="w-6 h-6 text-trail-700" />
                  </div>
                )}
                <div>
                  <div className="font-medium text-trail-800">{user.displayName}</div>
                  <div className="text-sm text-trail-600">@{user.username}</div>
                </div>
              </div>
            </div>
          )}

          {/* Navigation Items */}
          <nav className="flex-1 overflow-y-auto">
            <div className="p-2">
              {/* Create Trip - Only for authenticated users */}
              {isAuthenticated && (
                <>
                  <button
                    onClick={() => handleNavigation('/trips/new')}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg"
                  >
                    <Plus className="w-5 h-5 mr-3 text-forest-600" />
                    <span className="font-medium text-trail-700">Create New Trip</span>
                  </button>

                  {/* My Trips */}
                  <button
                    onClick={() => {
                      dispatch(setActivePanel('trips'));
                      onClose();
                    }}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg"
                  >
                    <MapPin className="w-5 h-5 mr-3 text-trail-600" />
                    <span className="text-trail-700">My Trips</span>
                  </button>
                </>
              )}

              {/* Explore */}
              <button
                onClick={() => handleNavigation('/explore')}
                className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg"
              >
                <Compass className="w-5 h-5 mr-3 text-trail-600" />
                <span className="text-trail-700">Explore</span>
              </button>

              {/* Profile */}
              {isAuthenticated && (
                <>
                  <button
                    onClick={() => handleNavigation('/profile')}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg"
                  >
                    <User className="w-5 h-5 mr-3 text-trail-600" />
                    <span className="text-trail-700">Profile</span>
                  </button>

                  <button
                    onClick={() => handleNavigation('/settings')}
                    className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg"
                  >
                    <Settings className="w-5 h-5 mr-3 text-trail-600" />
                    <span className="text-trail-700">Settings</span>
                  </button>
                </>
              )}
            </div>
          </nav>

          {/* Footer Actions */}
          <div className="p-4 border-t border-terrain-300">
            {isAuthenticated ? (
              <button
                onClick={handleLogout}
                className="w-full flex items-center px-4 py-3 text-left hover:bg-terrain-200 rounded-lg text-red-600"
              >
                <LogOut className="w-5 h-5 mr-3" />
                <span>Log Out</span>
              </button>
            ) : (
              <button
                onClick={() => handleNavigation('/login')}
                className="w-full flex items-center justify-center px-4 py-3 bg-forest-600 text-white rounded-lg hover:bg-forest-700"
              >
                <LogIn className="w-4 h-4 mr-2" />
                Login
              </button>
            )}
          </div>
        </div>
      </div>
    </>
  );
};