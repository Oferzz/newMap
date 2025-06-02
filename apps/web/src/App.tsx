import { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Header } from './components/layout/Header';
import { MapView } from './components/map/MapView';
import { LoginPage } from './pages/LoginPage';
import { TripCreationPage } from './pages/TripCreationPage';
import { PrivateRoute } from './components/auth/PrivateRoute';
import { WebSocketProvider } from './providers/WebSocketProvider';
import { Notifications } from './components/layout/Notifications';
import { Toaster } from 'react-hot-toast';
import { useAppDispatch } from './hooks/redux';
import { initializeAuthThunk } from './store/thunks/auth.thunks';
import './App.css';

function App() {
  const dispatch = useAppDispatch();

  useEffect(() => {
    // Initialize authentication on app load
    dispatch(initializeAuthThunk());
  }, [dispatch]);
  return (
    <WebSocketProvider>
      <Router>
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

          <Routes>
            {/* Auth Routes - No Header */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<LoginPage isRegister />} />

            {/* Public Map View */}
            <Route path="/" element={
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
      </Router>
    </WebSocketProvider>
  );
}

export default App;