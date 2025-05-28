import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Header } from './components/layout/Header';
import { MapView } from './components/map/MapView';
import { LoginPage } from './pages/LoginPage';
import { TripCreationPage } from './pages/TripCreationPage';
import { PrivateRoute } from './components/auth/PrivateRoute';
import { WebSocketProvider } from './providers/WebSocketProvider';
import { Toaster } from 'react-hot-toast';
import './App.css';

function App() {
  return (
    <WebSocketProvider>
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

            {/* Public Map View */}
            <Route path="/" element={
              <>
                <Header />
                <MapView />
              </>
            } />
            
            {/* Protected Routes */}
            <Route
              path="/trips/new"
              element={
                <>
                  <Header />
                  <PrivateRoute>
                    <TripCreationPage />
                  </PrivateRoute>
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