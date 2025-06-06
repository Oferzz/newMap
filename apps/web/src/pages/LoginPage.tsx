import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAppDispatch } from '../hooks/redux';
import { loginThunk, registerThunk } from '../store/thunks/auth.thunks';
import { Mail, Lock, User, Loader2 } from 'lucide-react';

interface LoginPageProps {
  isRegister?: boolean;
}

export const LoginPage: React.FC<LoginPageProps> = ({ isRegister = false }) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    username: '',
    password: '',
    confirmPassword: '',
    displayName: '',
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const validateForm = () => {
    const errors: Record<string, string> = {};
    setFieldErrors({});

    // Email validation
    if (!formData.email) {
      errors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = 'Invalid email format';
    }

    if (isRegister) {
      // Username validation
      if (!formData.username) {
        errors.username = 'Username is required';
      } else if (formData.username.length < 3) {
        errors.username = 'Username must be at least 3 characters';
      } else if (formData.username.length > 30) {
        errors.username = 'Username cannot exceed 30 characters';
      } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.username)) {
        errors.username = 'Username can only contain letters, numbers, underscores, and hyphens';
      }

      // Display name validation
      if (!formData.displayName) {
        errors.displayName = 'Display name is required';
      } else if (formData.displayName.length < 2) {
        errors.displayName = 'Display name must be at least 2 characters';
      } else if (formData.displayName.length > 100) {
        errors.displayName = 'Display name cannot exceed 100 characters';
      }

      // Password confirmation
      if (formData.password !== formData.confirmPassword) {
        errors.confirmPassword = 'Passwords do not match';
      }
    }

    // Password validation
    if (!formData.password) {
      errors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      errors.password = 'Password must be at least 8 characters';
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    setFieldErrors({});

    try {
      if (isRegister) {
        await dispatch(registerThunk({
          email: formData.email,
          username: formData.username,
          password: formData.password,
          display_name: formData.displayName,
        })).unwrap();
        
        // After successful registration, log them in
        await dispatch(loginThunk({
          email: formData.email,
          password: formData.password,
        })).unwrap();
      } else {
        await dispatch(loginThunk({
          email: formData.email,
          password: formData.password,
        })).unwrap();
      }
      
      navigate('/');
    } catch (error: any) {
      // Handle specific field errors from the server
      if (error?.message) {
        const message = error.message.toLowerCase();
        if (message.includes('email already exists')) {
          setFieldErrors({ email: 'This email is already registered' });
        } else if (message.includes('username already exists')) {
          setFieldErrors({ username: 'This username is already taken' });
        } else if (message.includes('invalid credentials')) {
          setFieldErrors({ email: 'Invalid email or password', password: 'Invalid email or password' });
        }
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative">
      {/* Blurred background */}
      <div className="absolute inset-0 bg-gradient-to-br from-terrain-100 via-terrain-200 to-forest-100 backdrop-blur-sm" />
      
      <div className="max-w-md w-full relative z-10">
        <div className="bg-terrain-100 rounded-2xl shadow-xl border border-terrain-300 p-8">
          <div className="text-center mb-8">
            <img 
              src="/logo.svg" 
              alt="newMap" 
              className="h-16 w-auto mx-auto mb-6 drop-shadow-md"
            />
            <h1 className="text-3xl font-bold text-trail-800">
              {isRegister ? 'Create Account' : 'Welcome Back'}
            </h1>
            <p className="text-trail-600 mt-2">
              {isRegister 
                ? 'Start planning your perfect trips' 
                : 'Sign in to continue your journey'}
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-trail-700 mb-1">
                Email
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className={`w-full pl-10 pr-4 py-2 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400 ${fieldErrors.email ? 'border-red-500' : 'border-terrain-300'}`}
                  placeholder="you@example.com"
                />
              </div>
              {fieldErrors.email && <p className="text-red-500 text-sm mt-1">{fieldErrors.email}</p>}
            </div>

            {isRegister && (
              <>
                <div>
                  <label className="block text-sm font-medium text-trail-700 mb-1">
                    Username
                  </label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                    <input
                      type="text"
                      required
                      pattern="^[a-zA-Z0-9_-]+$"
                      title="Username can only contain letters, numbers, underscores, and hyphens"
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value.replace(/\s/g, '') })}
                      className={`w-full pl-10 pr-4 py-2 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400 ${fieldErrors.username ? 'border-red-500' : 'border-terrain-300'}`}
                      placeholder="johndoe"
                    />
                  </div>
                  {fieldErrors.username && <p className="text-red-500 text-sm mt-1">{fieldErrors.username}</p>}
                </div>

                <div>
                  <label className="block text-sm font-medium text-trail-700 mb-1">
                    Display Name
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.displayName}
                    onChange={(e) => setFormData({ ...formData, displayName: e.target.value })}
                    className={`w-full px-4 py-2 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400 ${fieldErrors.displayName ? 'border-red-500' : 'border-terrain-300'}`}
                    placeholder="John Doe"
                  />
                  {fieldErrors.displayName && <p className="text-red-500 text-sm mt-1">{fieldErrors.displayName}</p>}
                </div>
              </>
            )}

            <div>
              <label className="block text-sm font-medium text-trail-700 mb-1">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                <input
                  type="password"
                  required
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className={`w-full pl-10 pr-4 py-2 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400 ${fieldErrors.password ? 'border-red-500' : 'border-terrain-300'}`}
                  placeholder="••••••••"
                />
              </div>
              {fieldErrors.password && <p className="text-red-500 text-sm mt-1">{fieldErrors.password}</p>}
            </div>

            {isRegister && (
              <div>
                <label className="block text-sm font-medium text-trail-700 mb-1">
                  Confirm Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                  <input
                    type="password"
                    required
                    value={formData.confirmPassword}
                    onChange={(e) => setFormData({ ...formData, confirmPassword: e.target.value })}
                    className={`w-full pl-10 pr-4 py-2 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400 ${fieldErrors.confirmPassword ? 'border-red-500' : 'border-terrain-300'}`}
                    placeholder="••••••••"
                  />
                </div>
                {fieldErrors.confirmPassword && <p className="text-red-500 text-sm mt-1">{fieldErrors.confirmPassword}</p>}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3 bg-forest-600 text-trail-800 rounded-lg font-medium hover:bg-forest-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 shadow-soft"
            >
              {isLoading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  {isRegister ? 'Creating Account...' : 'Signing In...'}
                </>
              ) : (
                isRegister ? 'Create Account' : 'Sign In'
              )}
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-trail-600">
              {isRegister ? 'Already have an account?' : "Don't have an account?"}{' '}
              <Link
                to={isRegister ? '/login' : '/register'}
                className="text-forest-600 hover:text-forest-700 font-medium"
              >
                {isRegister ? 'Sign In' : 'Create Account'}
              </Link>
            </p>
          </div>
        </div>

        <p className="text-center text-trail-600 text-sm mt-8">
          By continuing, you agree to our Terms of Service and Privacy Policy
        </p>
      </div>
    </div>
  );
};