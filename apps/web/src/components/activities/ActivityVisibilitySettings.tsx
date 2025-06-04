import React, { useState } from 'react';
import { Globe, Users, Lock, MessageCircle, Download, Share2, Copy, Check } from 'lucide-react';

interface VisibilitySettings {
  privacy: 'public' | 'friends' | 'private';
  allowComments: boolean;
  allowDownloads: boolean;
  shareWithGroups: string[];
}

interface ActivityVisibilitySettingsProps {
  settings: VisibilitySettings;
  onUpdate: (settings: VisibilitySettings) => void;
}

const PRIVACY_OPTIONS = [
  {
    value: 'public',
    label: 'Public',
    description: 'Anyone can find and view this activity',
    icon: Globe,
    color: 'green'
  },
  {
    value: 'friends',
    label: 'Friends Only',
    description: 'Only your friends can view this activity',
    icon: Users,
    color: 'blue'
  },
  {
    value: 'private',
    label: 'Private',
    description: 'Only you can view this activity',
    icon: Lock,
    color: 'gray'
  }
];

export const ActivityVisibilitySettings: React.FC<ActivityVisibilitySettingsProps> = ({
  settings,
  onUpdate
}) => {
  const [shareLink, setShareLink] = useState('');
  const [linkCopied, setLinkCopied] = useState(false);

  const updateSettings = (field: keyof VisibilitySettings, value: any) => {
    onUpdate({
      ...settings,
      [field]: value
    });
  };

  const generateShareLink = () => {
    // In real implementation, this would generate an actual share link
    const mockLink = `https://newmap.app/activities/shared/${Math.random().toString(36).substr(2, 9)}`;
    setShareLink(mockLink);
  };

  const copyShareLink = async () => {
    if (shareLink) {
      try {
        await navigator.clipboard.writeText(shareLink);
        setLinkCopied(true);
        setTimeout(() => setLinkCopied(false), 2000);
      } catch (error) {
        console.error('Failed to copy link:', error);
      }
    }
  };

  return (
    <div className="space-y-6">
      {/* Privacy Level */}
      <div>
        <h3 className="text-lg font-semibold text-trail-800 mb-4">Privacy Settings</h3>
        <div className="space-y-3">
          {PRIVACY_OPTIONS.map((option) => {
            const Icon = option.icon;
            const isSelected = settings.privacy === option.value;
            
            return (
              <button
                key={option.value}
                onClick={() => updateSettings('privacy', option.value)}
                className={`w-full p-4 border-2 rounded-lg text-left transition-colors ${
                  isSelected
                    ? option.color === 'green'
                      ? 'border-green-300 bg-green-50'
                      : option.color === 'blue'
                      ? 'border-blue-300 bg-blue-50'
                      : 'border-gray-300 bg-gray-50'
                    : 'border-terrain-200 bg-white hover:bg-terrain-50'
                }`}
              >
                <div className="flex items-start space-x-3">
                  <Icon 
                    className={`w-5 h-5 mt-0.5 ${
                      isSelected
                        ? option.color === 'green'
                          ? 'text-green-600'
                          : option.color === 'blue'
                          ? 'text-blue-600'
                          : 'text-gray-600'
                        : 'text-trail-500'
                    }`}
                  />
                  <div className="flex-1">
                    <div className={`font-medium ${
                      isSelected ? 'text-trail-800' : 'text-trail-700'
                    }`}>
                      {option.label}
                    </div>
                    <div className={`text-sm mt-1 ${
                      isSelected ? 'text-trail-600' : 'text-trail-500'
                    }`}>
                      {option.description}
                    </div>
                  </div>
                  {isSelected && (
                    <Check className="w-5 h-5 text-green-600" />
                  )}
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Interaction Settings */}
      <div>
        <h3 className="text-lg font-semibold text-trail-800 mb-4">Interaction Settings</h3>
        <div className="space-y-4">
          {/* Allow Comments */}
          <div className="flex items-center justify-between p-4 bg-terrain-50 rounded-lg">
            <div className="flex items-center space-x-3">
              <MessageCircle className="w-5 h-5 text-trail-600" />
              <div>
                <div className="font-medium text-trail-800">Allow Comments</div>
                <div className="text-sm text-trail-600">
                  Let others comment on your activity
                </div>
              </div>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.allowComments}
                onChange={(e) => updateSettings('allowComments', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-terrain-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-forest-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-forest-600"></div>
            </label>
          </div>

          {/* Allow Downloads */}
          <div className="flex items-center justify-between p-4 bg-terrain-50 rounded-lg">
            <div className="flex items-center space-x-3">
              <Download className="w-5 h-5 text-trail-600" />
              <div>
                <div className="font-medium text-trail-800">Allow Downloads</div>
                <div className="text-sm text-trail-600">
                  Let others download GPS tracks and data
                </div>
              </div>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.allowDownloads}
                onChange={(e) => updateSettings('allowDownloads', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-terrain-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-forest-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-forest-600"></div>
            </label>
          </div>
        </div>
      </div>

      {/* Share Link */}
      {settings.privacy !== 'private' && (
        <div>
          <h3 className="text-lg font-semibold text-trail-800 mb-4">Share Activity</h3>
          <div className="space-y-3">
            <div className="text-sm text-trail-600">
              Generate a shareable link for this activity that others can access.
            </div>
            
            {!shareLink ? (
              <button
                onClick={generateShareLink}
                className="flex items-center space-x-2 px-4 py-2 bg-forest-600 text-white rounded-lg hover:bg-forest-700 transition-colors"
              >
                <Share2 className="w-4 h-4" />
                <span>Generate Share Link</span>
              </button>
            ) : (
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  value={shareLink}
                  readOnly
                  className="flex-1 px-3 py-2 bg-terrain-50 border border-terrain-300 rounded-lg text-sm text-trail-700"
                />
                <button
                  onClick={copyShareLink}
                  className={`flex items-center space-x-1 px-3 py-2 rounded-lg transition-colors ${
                    linkCopied
                      ? 'bg-green-100 text-green-700'
                      : 'bg-terrain-100 text-trail-700 hover:bg-terrain-200'
                  }`}
                >
                  {linkCopied ? (
                    <>
                      <Check className="w-4 h-4" />
                      <span className="text-sm">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      <span className="text-sm">Copy</span>
                    </>
                  )}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Privacy Notice */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="text-sm text-blue-800">
          <div className="font-medium mb-1">Privacy Notice</div>
          <div>
            {settings.privacy === 'public' && (
              "This activity will be visible to everyone and may appear in search results and recommendations."
            )}
            {settings.privacy === 'friends' && (
              "This activity will only be visible to your friends and people you share the link with."
            )}
            {settings.privacy === 'private' && (
              "This activity will only be visible to you. You can change this setting later."
            )}
          </div>
        </div>
      </div>

      {/* Additional Options */}
      <div className="border-t border-terrain-200 pt-6">
        <h4 className="text-md font-medium text-trail-800 mb-3">Additional Options</h4>
        <div className="space-y-3">
          <label className="flex items-center space-x-3">
            <input
              type="checkbox"
              className="w-4 h-4 text-forest-600 border-terrain-300 rounded focus:ring-forest-500"
            />
            <span className="text-sm text-trail-700">
              Allow this activity to be featured in community highlights
            </span>
          </label>
          
          <label className="flex items-center space-x-3">
            <input
              type="checkbox"
              className="w-4 h-4 text-forest-600 border-terrain-300 rounded focus:ring-forest-500"
            />
            <span className="text-sm text-trail-700">
              Send me notifications when others interact with this activity
            </span>
          </label>
          
          <label className="flex items-center space-x-3">
            <input
              type="checkbox"
              defaultChecked
              className="w-4 h-4 text-forest-600 border-terrain-300 rounded focus:ring-forest-500"
            />
            <span className="text-sm text-trail-700">
              Include this activity in my public profile
            </span>
          </label>
        </div>
      </div>
    </div>
  );
};