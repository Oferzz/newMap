import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Provider } from 'react-redux';
import { store } from './store';
import { Header } from './components/layout/Header';
import { MapView } from './components/map/MapView';
import { LoginPage } from './pages/LoginPage';
import { TripCreationPage } from './pages/TripCreationPage';
import { PrivateRoute } from './components/auth/PrivateRoute';
import { Toaster } from 'react-hot-toast';
import './App.css';

function App() {
  return (
    <Provider store={store}>
      <Router>
        <div className="relative h-screen w-screen overflow-hidden">
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

          <Routes>
            {/* Auth Routes - No Header */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<LoginPage isRegister />} />

            {/* Main App Routes - With Header */}
            <Route
              path="/*"
              element={
                <>
                  <Header />
                  <Routes>
                    {/* Public Map View */}
                    <Route path="/" element={<MapView />} />
                    
                    {/* Protected Routes */}
                    <Route
                      path="/trips/new"
                      element={
                        <PrivateRoute>
                          <TripCreationPage />
                        </PrivateRoute>
                      }
                    />
                    
                    <Route
                      path="/trips/:id"
                      element={
                        <PrivateRoute>
                          <MapView />
                        </PrivateRoute>
                      }
                    />
                    
                    <Route
                      path="/places/:id"
                      element={<MapView />}
                    />
                    
                    <Route
                      path="/profile"
                      element={
                        <PrivateRoute>
                          <MapView />
                        </PrivateRoute>
                      }
                    />
                  </Routes>
                </>
              }
            />
          </Routes>
        </div>
      </Router>
    </Provider>
  );
}

export default App;