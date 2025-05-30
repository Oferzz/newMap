import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAppDispatch } from '../hooks/redux';
import { loginThunk, registerThunk } from '../store/thunks/auth.thunks';
import { Mail, Lock, User, Loader2 } from 'lucide-react';
import toast from 'react-hot-toast';

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (isRegister && formData.password !== formData.confirmPassword) {
      toast.error('Passwords do not match');
      return;
    }

    setIsLoading(true);

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
    } catch (error) {
      // Error handling is done in the thunks
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-terrain-200 p-4">
      <div className="max-w-md w-full">
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
                  className="w-full pl-10 pr-4 py-2 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400"
                  placeholder="you@example.com"
                />
              </div>
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
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                      className="w-full pl-10 pr-4 py-2 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400"
                      placeholder="johndoe"
                    />
                  </div>
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
                    className="w-full px-4 py-2 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400"
                    placeholder="John Doe"
                  />
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
                  className="w-full pl-10 pr-4 py-2 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400"
                  placeholder="••••••••"
                />
              </div>
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
                    className="w-full pl-10 pr-4 py-2 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent placeholder-trail-400"
                    placeholder="••••••••"
                  />
                </div>
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-3 bg-forest-600 text-white rounded-lg font-medium hover:bg-forest-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 shadow-soft"
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