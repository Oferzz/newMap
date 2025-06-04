import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { 
  User, 
  Mail, 
  MapPin, 
  Edit3, 
  Camera, 
  Save, 
  X, 
  Lock,
  Eye,
  EyeOff,
  ArrowLeft,
  Calendar,
  Settings,
  Shield
} from 'lucide-react';
import { updateProfileThunk, changePasswordThunk } from '../store/thunks/auth.thunks';
import { UpdateProfileInput } from '../services/auth.service';
import toast from 'react-hot-toast';

export const ProfilePage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector((state) => state.auth);
  
  const [activeTab, setActiveTab] = useState<'profile' | 'password' | 'preferences'>('profile');
  const [isEditing, setIsEditing] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  
  // Profile form state
  const [profileForm, setProfileForm] = useState({
    display_name: '',
    bio: '',
    location: '',
    avatar_url: '',
  });
  
  // Password form state
  const [passwordForm, setPasswordForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });
  
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [isSaving, setIsSaving] = useState(false);

  // Initialize form with user data
  useEffect(() => {
    if (user) {
      setProfileForm({
        display_name: user.displayName || '',
        bio: user.bio || '',
        location: user.location || '',
        avatar_url: user.avatarUrl || '',
      });
    }
  }, [user]);

  const validateProfileForm = () => {
    const errors: Record<string, string> = {};
    
    if (!profileForm.display_name.trim()) {
      errors.display_name = 'Display name is required';
    } else if (profileForm.display_name.length < 2) {
      errors.display_name = 'Display name must be at least 2 characters';
    } else if (profileForm.display_name.length > 100) {
      errors.display_name = 'Display name cannot exceed 100 characters';
    }
    
    if (profileForm.bio && profileForm.bio.length > 500) {
      errors.bio = 'Bio cannot exceed 500 characters';
    }
    
    if (profileForm.avatar_url && profileForm.avatar_url.trim()) {
      try {
        new URL(profileForm.avatar_url);
      } catch {
        errors.avatar_url = 'Please enter a valid URL';
      }
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const validatePasswordForm = () => {
    const errors: Record<string, string> = {};
    
    if (!passwordForm.current_password) {
      errors.current_password = 'Current password is required';
    }
    
    if (!passwordForm.new_password) {
      errors.new_password = 'New password is required';
    } else if (passwordForm.new_password.length < 8) {
      errors.new_password = 'Password must be at least 8 characters';
    }
    
    if (passwordForm.new_password !== passwordForm.confirm_password) {
      errors.confirm_password = 'Passwords do not match';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSaveProfile = async () => {
    if (!validateProfileForm()) return;
    
    setIsSaving(true);
    setFormErrors({});
    
    try {
      const updateData: UpdateProfileInput = {};
      
      if (profileForm.display_name !== user?.displayName) {
        updateData.display_name = profileForm.display_name;
      }
      if (profileForm.bio !== (user?.bio || '')) {
        updateData.bio = profileForm.bio || undefined;
      }
      if (profileForm.location !== (user?.location || '')) {
        updateData.location = profileForm.location || undefined;
      }
      if (profileForm.avatar_url !== (user?.avatarUrl || '')) {
        updateData.avatar_url = profileForm.avatar_url || undefined;
      }
      
      if (Object.keys(updateData).length > 0) {
        await dispatch(updateProfileThunk(updateData)).unwrap();
      }
      
      setIsEditing(false);
    } catch (error: any) {
      if (error?.message) {
        toast.error(error.message);
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleChangePassword = async () => {
    if (!validatePasswordForm()) return;
    
    setIsSaving(true);
    setFormErrors({});
    
    try {
      await dispatch(changePasswordThunk({
        current_password: passwordForm.current_password,
        new_password: passwordForm.new_password,
      })).unwrap();
      
      setPasswordForm({
        current_password: '',
        new_password: '',
        confirm_password: '',
      });
    } catch (error: any) {
      if (error?.message) {
        const message = error.message.toLowerCase();
        if (message.includes('current password')) {
          setFormErrors({ current_password: 'Current password is incorrect' });
        } else {
          toast.error(error.message);
        }
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancelEdit = () => {
    if (user) {
      setProfileForm({
        display_name: user.displayName || '',
        bio: user.bio || '',
        location: user.location || '',
        avatar_url: user.avatarUrl || '',
      });
    }
    setIsEditing(false);
    setFormErrors({});
  };

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 relative">
        {/* Blurred background */}
        <div className="absolute inset-0 bg-gradient-to-br from-terrain-100 via-terrain-200 to-forest-100 backdrop-blur-sm" />
        
        <div className="relative z-10 text-center">
          <p className="text-trail-700">Please log in to view your profile.</p>
          <button 
            onClick={() => navigate('/login')} 
            className="mt-4 px-4 py-2 bg-forest-600 text-white rounded-lg hover:bg-forest-700 transition-colors"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative">
      {/* Blurred background */}
      <div className="absolute inset-0 bg-gradient-to-br from-terrain-100 via-terrain-200 to-forest-100 backdrop-blur-sm" />
      
      <div className="relative z-10 max-w-4xl w-full">
        <div className="bg-terrain-100 rounded-2xl shadow-xl border border-terrain-300 overflow-hidden">
          {/* Header */}
          <div className="p-6 border-b border-terrain-300 bg-gradient-to-r from-forest-500 to-forest-600 text-white">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <button
                  onClick={() => navigate('/')}
                  className="p-2 hover:bg-white/20 rounded-lg transition-colors"
                >
                  <ArrowLeft className="w-5 h-5" />
                </button>
                <div>
                  <h1 className="text-2xl font-bold">Profile Settings</h1>
                  <p className="text-forest-100 mt-1">Manage your account and preferences</p>
                </div>
              </div>
              {!isEditing && activeTab === 'profile' && (
                <button
                  onClick={() => setIsEditing(true)}
                  className="flex items-center gap-2 px-4 py-2 bg-white/20 hover:bg-white/30 rounded-lg transition-colors"
                >
                  <Edit3 className="w-4 h-4" />
                  Edit Profile
                </button>
              )}
            </div>
          </div>

          {/* Tabs */}
          <div className="flex border-b border-terrain-300 bg-terrain-50">
            <button
              onClick={() => setActiveTab('profile')}
              className={`flex items-center gap-2 px-6 py-4 font-medium transition-colors ${
                activeTab === 'profile'
                  ? 'text-forest-700 border-b-2 border-forest-600 bg-white'
                  : 'text-trail-600 hover:text-trail-800 hover:bg-terrain-100'
              }`}
            >
              <User className="w-4 h-4" />
              Profile
            </button>
            <button
              onClick={() => setActiveTab('password')}
              className={`flex items-center gap-2 px-6 py-4 font-medium transition-colors ${
                activeTab === 'password'
                  ? 'text-forest-700 border-b-2 border-forest-600 bg-white'
                  : 'text-trail-600 hover:text-trail-800 hover:bg-terrain-100'
              }`}
            >
              <Shield className="w-4 h-4" />
              Security
            </button>
            <button
              onClick={() => setActiveTab('preferences')}
              className={`flex items-center gap-2 px-6 py-4 font-medium transition-colors ${
                activeTab === 'preferences'
                  ? 'text-forest-700 border-b-2 border-forest-600 bg-white'
                  : 'text-trail-600 hover:text-trail-800 hover:bg-terrain-100'
              }`}
            >
              <Settings className="w-4 h-4" />
              Preferences
            </button>
          </div>

          {/* Content */}
          <div className="p-6">
            {activeTab === 'profile' && (
              <div className="max-w-2xl">
                {/* Avatar Section */}
                <div className="flex items-center gap-6 mb-8">
                  <div className="relative">
                    {user.avatarUrl || profileForm.avatar_url ? (
                      <img 
                        src={profileForm.avatar_url || user.avatarUrl} 
                        alt={user.displayName}
                        className="w-24 h-24 rounded-full object-cover border-4 border-terrain-300"
                        onError={(e) => {
                          e.currentTarget.src = '/default-avatar.png';
                        }}
                      />
                    ) : (
                      <div className="w-24 h-24 rounded-full bg-terrain-300 flex items-center justify-center">
                        <User className="w-12 h-12 text-trail-500" />
                      </div>
                    )}
                    {isEditing && (
                      <button className="absolute bottom-0 right-0 p-2 bg-forest-600 text-white rounded-full hover:bg-forest-700 transition-colors">
                        <Camera className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold text-trail-800">{user.displayName}</h2>
                    <p className="text-trail-600">@{user.username}</p>
                    <div className="flex items-center gap-2 mt-2 text-sm text-trail-500">
                      <Calendar className="w-4 h-4" />
                      Member since {user.created_at ? new Date(user.created_at).toLocaleDateString() : 'Unknown'}
                    </div>
                  </div>
                </div>

                {/* Profile Form */}
                <div className="space-y-6">
                  {isEditing && (
                    <div>
                      <label className="block text-sm font-medium text-trail-700 mb-2">
                        Avatar URL
                      </label>
                      <input
                        type="url"
                        value={profileForm.avatar_url}
                        onChange={(e) => setProfileForm(prev => ({ ...prev, avatar_url: e.target.value }))}
                        className={`w-full px-4 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.avatar_url ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        placeholder="https://example.com/avatar.jpg"
                      />
                      {formErrors.avatar_url && <p className="text-red-500 text-sm mt-1">{formErrors.avatar_url}</p>}
                    </div>
                  )}

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Display Name
                    </label>
                    {isEditing ? (
                      <input
                        type="text"
                        value={profileForm.display_name}
                        onChange={(e) => setProfileForm(prev => ({ ...prev, display_name: e.target.value }))}
                        className={`w-full px-4 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.display_name ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        placeholder="Your display name"
                      />
                    ) : (
                      <p className="px-4 py-3 bg-terrain-50 border border-terrain-300 rounded-lg">{user.displayName}</p>
                    )}
                    {formErrors.display_name && <p className="text-red-500 text-sm mt-1">{formErrors.display_name}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Email
                    </label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                      <p className="pl-10 pr-4 py-3 bg-gray-100 border border-terrain-300 rounded-lg text-trail-600">
                        {user.email}
                      </p>
                    </div>
                    <p className="text-xs text-trail-500 mt-1">Email cannot be changed</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Bio
                    </label>
                    {isEditing ? (
                      <textarea
                        value={profileForm.bio}
                        onChange={(e) => setProfileForm(prev => ({ ...prev, bio: e.target.value }))}
                        className={`w-full px-4 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.bio ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        rows={4}
                        placeholder="Tell us about yourself..."
                      />
                    ) : (
                      <p className="px-4 py-3 bg-terrain-50 border border-terrain-300 rounded-lg min-h-[100px]">
                        {user.bio || 'No bio added yet.'}
                      </p>
                    )}
                    {formErrors.bio && <p className="text-red-500 text-sm mt-1">{formErrors.bio}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Location
                    </label>
                    {isEditing ? (
                      <div className="relative">
                        <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                        <input
                          type="text"
                          value={profileForm.location}
                          onChange={(e) => setProfileForm(prev => ({ ...prev, location: e.target.value }))}
                          className="w-full pl-10 pr-4 py-3 bg-terrain-50 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
                          placeholder="Your location"
                        />
                      </div>
                    ) : (
                      <div className="relative">
                        <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                        <p className="pl-10 pr-4 py-3 bg-terrain-50 border border-terrain-300 rounded-lg">
                          {user.location || 'No location added yet.'}
                        </p>
                      </div>
                    )}
                  </div>

                  {isEditing && (
                    <div className="flex gap-3 pt-4">
                      <button
                        onClick={handleSaveProfile}
                        disabled={isSaving}
                        className="flex items-center gap-2 px-6 py-3 bg-forest-600 text-white rounded-lg hover:bg-forest-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        <Save className="w-4 h-4" />
                        {isSaving ? 'Saving...' : 'Save Changes'}
                      </button>
                      <button
                        onClick={handleCancelEdit}
                        disabled={isSaving}
                        className="flex items-center gap-2 px-6 py-3 border border-terrain-300 rounded-lg hover:bg-terrain-100 transition-colors"
                      >
                        <X className="w-4 h-4" />
                        Cancel
                      </button>
                    </div>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'password' && (
              <div className="max-w-md">
                <div className="mb-6">
                  <h3 className="text-lg font-semibold text-trail-800 mb-2">Change Password</h3>
                  <p className="text-trail-600">Ensure your account is protected with a strong password.</p>
                </div>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Current Password
                    </label>
                    <div className="relative">
                      <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                      <input
                        type={showPassword ? 'text' : 'password'}
                        value={passwordForm.current_password}
                        onChange={(e) => setPasswordForm(prev => ({ ...prev, current_password: e.target.value }))}
                        className={`w-full pl-10 pr-12 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.current_password ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        placeholder="Enter current password"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-trail-500 hover:text-trail-700"
                      >
                        {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                    {formErrors.current_password && <p className="text-red-500 text-sm mt-1">{formErrors.current_password}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      New Password
                    </label>
                    <div className="relative">
                      <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                      <input
                        type={showNewPassword ? 'text' : 'password'}
                        value={passwordForm.new_password}
                        onChange={(e) => setPasswordForm(prev => ({ ...prev, new_password: e.target.value }))}
                        className={`w-full pl-10 pr-12 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.new_password ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        placeholder="Enter new password"
                      />
                      <button
                        type="button"
                        onClick={() => setShowNewPassword(!showNewPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-trail-500 hover:text-trail-700"
                      >
                        {showNewPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                    {formErrors.new_password && <p className="text-red-500 text-sm mt-1">{formErrors.new_password}</p>}
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-trail-700 mb-2">
                      Confirm New Password
                    </label>
                    <div className="relative">
                      <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-trail-500 w-5 h-5" />
                      <input
                        type={showConfirmPassword ? 'text' : 'password'}
                        value={passwordForm.confirm_password}
                        onChange={(e) => setPasswordForm(prev => ({ ...prev, confirm_password: e.target.value }))}
                        className={`w-full pl-10 pr-12 py-3 bg-terrain-50 border rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent ${
                          formErrors.confirm_password ? 'border-red-500' : 'border-terrain-300'
                        }`}
                        placeholder="Confirm new password"
                      />
                      <button
                        type="button"
                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-trail-500 hover:text-trail-700"
                      >
                        {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                    {formErrors.confirm_password && <p className="text-red-500 text-sm mt-1">{formErrors.confirm_password}</p>}
                  </div>

                  <button
                    onClick={handleChangePassword}
                    disabled={isSaving || !passwordForm.current_password || !passwordForm.new_password || !passwordForm.confirm_password}
                    className="w-full py-3 bg-forest-600 text-white rounded-lg hover:bg-forest-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                  >
                    <Shield className="w-4 h-4" />
                    {isSaving ? 'Changing Password...' : 'Change Password'}
                  </button>
                </div>
              </div>
            )}

            {activeTab === 'preferences' && (
              <div className="max-w-2xl">
                <div className="mb-6">
                  <h3 className="text-lg font-semibold text-trail-800 mb-2">Preferences</h3>
                  <p className="text-trail-600">Customize your experience on the platform.</p>
                </div>

                <div className="space-y-6">
                  <div className="p-4 bg-terrain-50 rounded-lg border border-terrain-300">
                    <h4 className="font-medium text-trail-800 mb-2">Notifications</h4>
                    <div className="space-y-3">
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" defaultChecked />
                        <span className="ml-2 text-sm text-trail-700">Email notifications</span>
                      </label>
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" defaultChecked />
                        <span className="ml-2 text-sm text-trail-700">Trip invitations</span>
                      </label>
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" />
                        <span className="ml-2 text-sm text-trail-700">Marketing emails</span>
                      </label>
                    </div>
                  </div>

                  <div className="p-4 bg-terrain-50 rounded-lg border border-terrain-300">
                    <h4 className="font-medium text-trail-800 mb-2">Privacy</h4>
                    <div className="space-y-3">
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" defaultChecked />
                        <span className="ml-2 text-sm text-trail-700">Make profile public</span>
                      </label>
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" defaultChecked />
                        <span className="ml-2 text-sm text-trail-700">Allow location sharing</span>
                      </label>
                      <label className="flex items-center">
                        <input type="checkbox" className="rounded border-terrain-300 text-forest-600 focus:ring-forest-500" />
                        <span className="ml-2 text-sm text-trail-700">Show in search results</span>
                      </label>
                    </div>
                  </div>

                  <button className="px-6 py-3 bg-forest-600 text-white rounded-lg hover:bg-forest-700 transition-colors">
                    Save Preferences
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};