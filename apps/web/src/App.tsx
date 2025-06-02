import { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation, useNavigate } from 'react-router-dom';
import { Header } from './components/layout/Header';
import { MapView } from './components/map/MapView';
import { LoginModal } from './components/auth/LoginModal';
import { TripCreationPage } from './pages/TripCreationPage';
import { PrivateRoute } from './components/auth/PrivateRoute';
import { WebSocketProvider } from './providers/WebSocketProvider';
import { Notifications } from './components/layout/Notifications';
import { Toaster } from 'react-hot-toast';
import { useAppDispatch } from './hooks/redux';
import { initializeAuthThunk } from './store/thunks/auth.thunks';
import './App.css';

function AppContent() {
  const location = useLocation();
  const navigate = useNavigate();
  const [showLoginModal, setShowLoginModal] = useState(false);
  const [isRegisterMode, setIsRegisterMode] = useState(false);

  useEffect(() => {
    // Check if we're on login or register route
    if (location.pathname === '/login') {
      setShowLoginModal(true);
      setIsRegisterMode(false);
    } else if (location.pathname === '/register') {
      setShowLoginModal(true);
      setIsRegisterMode(true);
    } else {
      setShowLoginModal(false);
    }
  }, [location]);

  const handleCloseModal = () => {
    setShowLoginModal(false);
    navigate('/');
  };

  return (
    <div className="relative h-screen w-full overflow-auto">
      {/* Global Toast Notifications */}
      <Toaster 
        position="top-center"
        toastOptions={{
          duration: 4000,
          style: {
            marginTop: '4rem', // Account for header
          },
        }}
      />
      
      {/* Custom Notifications */}
      <Notifications />

      {/* Login Modal */}
      <LoginModal 
        isOpen={showLoginModal}
        onClose={handleCloseModal}
        isRegister={isRegisterMode}
      />

      <Routes>
        {/* Main map view - shows on all routes as background */}
        <Route path="/" element={
          <>
            <Header />
            <MapView />
          </>
        } />
        
        {/* Login/Register routes - still show map in background */}
        <Route path="/login" element={
          <>
            <Header />
            <MapView />
          </>
        } />
        
        <Route path="/register" element={
          <>
            <Header />
            <MapView />
          </>
        } />
        
        {/* Trip Creation - Works in guest mode */}
        <Route
          path="/trips/new"
          element={
            <>
              <Header />
              <TripCreationPage />
            </>
          }
        />
        
        <Route
          path="/trips/:id"
          element={
            <>
              <Header />
              <PrivateRoute>
                <MapView />
              </PrivateRoute>
            </>
          }
        />
        
        <Route
          path="/places/:id"
          element={
            <>
              <Header />
              <MapView />
            </>
          }
        />
        
        <Route
          path="/profile"
          element={
            <>
              <Header />
              <PrivateRoute>
                <MapView />
              </PrivateRoute>
            </>
          }
        />
      </Routes>
    </div>
  );
}

function App() {
  const dispatch = useAppDispatch();

  useEffect(() => {
    // Initialize authentication on app load
    dispatch(initializeAuthThunk());
  }, [dispatch]);

  return (
    <WebSocketProvider>
      <Router>
        <AppContent />
      </Router>
    </WebSocketProvider>
  );
}

export default App;