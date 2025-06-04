import React, { useState } from 'react';
import { Clock, TrendingUp, Mountain, Droplets, Package, Calendar, AlertTriangle, Tag } from 'lucide-react';

interface ActivityMetadata {
  difficulty: 'easy' | 'moderate' | 'hard' | 'expert';
  duration: number; // in hours
  distance: number; // in kilometers
  elevationGain: number; // in meters
  terrain: string[];
  waterFeatures: string[];
  gear: string[];
  seasons: string[];
  conditions: string[];
  tags: string[];
}

interface ActivityMetadataFormProps {
  metadata: ActivityMetadata;
  activityType: string;
  onUpdate: (metadata: ActivityMetadata) => void;
}

const TERRAIN_OPTIONS = [
  'trail', 'dirt-road', 'paved-road', 'rock', 'sand', 'mud', 'snow', 'ice',
  'grass', 'forest', 'desert', 'alpine', 'coastal', 'urban'
];

const WATER_FEATURES = [
  'river', 'lake', 'waterfall', 'stream', 'pond', 'ocean', 'bay', 'creek',
  'hot-springs', 'swimming-hole', 'rapids', 'beach'
];

const GEAR_OPTIONS = [
  'hiking-boots', 'trail-shoes', 'trekking-poles', 'backpack', 'water-bottles',
  'headlamp', 'first-aid-kit', 'map-compass', 'sunscreen', 'hat', 'layers',
  'rain-gear', 'gloves', 'helmet', 'harness', 'rope', 'crampons', 'microspikes'
];

const SEASON_OPTIONS = [
  'spring', 'summer', 'fall', 'winter', 'year-round'
];

const CONDITION_OPTIONS = [
  'well-maintained', 'overgrown', 'rocky', 'muddy', 'icy', 'steep', 'exposed',
  'shaded', 'buggy', 'crowded', 'remote', 'dog-friendly', 'family-friendly'
];

export const ActivityMetadataForm: React.FC<ActivityMetadataFormProps> = ({
  metadata,
  activityType: _activityType,
  onUpdate
}) => {
  const [showAdvanced, setShowAdvanced] = useState(false);

  const updateMetadata = (field: keyof ActivityMetadata, value: any) => {
    onUpdate({
      ...metadata,
      [field]: value
    });
  };

  const toggleArrayItem = (field: keyof ActivityMetadata, item: string) => {
    const currentArray = metadata[field] as string[];
    const newArray = currentArray.includes(item)
      ? currentArray.filter(i => i !== item)
      : [...currentArray, item];
    updateMetadata(field, newArray);
  };

  const formatTime = (hours: number): string => {
    if (hours < 1) {
      return `${Math.round(hours * 60)} min`;
    } else if (hours === Math.floor(hours)) {
      return `${hours} hr`;
    } else {
      const h = Math.floor(hours);
      const m = Math.round((hours - h) * 60);
      return `${h}h ${m}m`;
    }
  };

  return (
    <div className="space-y-6">
      {/* Basic Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Difficulty */}
        <div>
          <label className="block text-sm font-medium text-trail-700 mb-2">
            <AlertTriangle className="w-4 h-4 inline mr-1" />
            Difficulty Level
          </label>
          <select
            value={metadata.difficulty}
            onChange={(e) => updateMetadata('difficulty', e.target.value)}
            className="w-full px-3 py-2 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
          >
            <option value="easy">Easy - Beginner friendly</option>
            <option value="moderate">Moderate - Some experience required</option>
            <option value="hard">Hard - Experienced hikers</option>
            <option value="expert">Expert - Very challenging</option>
          </select>
        </div>

        {/* Duration */}
        <div>
          <label className="block text-sm font-medium text-trail-700 mb-2">
            <Clock className="w-4 h-4 inline mr-1" />
            Duration ({formatTime(metadata.duration)})
          </label>
          <div className="flex items-center space-x-2">
            <input
              type="range"
              min="0.5"
              max="24"
              step="0.5"
              value={metadata.duration}
              onChange={(e) => updateMetadata('duration', parseFloat(e.target.value))}
              className="flex-1"
            />
            <span className="text-sm text-trail-600 w-16">
              {formatTime(metadata.duration)}
            </span>
          </div>
        </div>

        {/* Distance */}
        <div>
          <label className="block text-sm font-medium text-trail-700 mb-2">
            Distance (km)
          </label>
          <input
            type="number"
            min="0"
            step="0.1"
            value={metadata.distance}
            onChange={(e) => updateMetadata('distance', parseFloat(e.target.value) || 0)}
            className="w-full px-3 py-2 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
            placeholder="10.5"
          />
        </div>

        {/* Elevation Gain */}
        <div>
          <label className="block text-sm font-medium text-trail-700 mb-2">
            <TrendingUp className="w-4 h-4 inline mr-1" />
            Elevation Gain (m)
          </label>
          <input
            type="number"
            min="0"
            step="10"
            value={metadata.elevationGain}
            onChange={(e) => updateMetadata('elevationGain', parseInt(e.target.value) || 0)}
            className="w-full px-3 py-2 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
            placeholder="500"
          />
        </div>
      </div>

      {/* Advanced Options Toggle */}
      <div className="border-t border-terrain-200 pt-4">
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="flex items-center space-x-2 text-forest-600 hover:text-forest-800 font-medium text-sm"
        >
          <span>{showAdvanced ? 'Hide' : 'Show'} Advanced Options</span>
          <svg
            className={`w-4 h-4 transition-transform ${showAdvanced ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      {/* Advanced Options */}
      {showAdvanced && (
        <div className="space-y-6">
          {/* Terrain Types */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-3">
              <Mountain className="w-4 h-4 inline mr-1" />
              Terrain Types
            </label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {TERRAIN_OPTIONS.map((terrain) => (
                <button
                  key={terrain}
                  onClick={() => toggleArrayItem('terrain', terrain)}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    metadata.terrain.includes(terrain)
                      ? 'bg-forest-100 border-forest-300 text-forest-800'
                      : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
                  }`}
                >
                  {terrain.replace('-', ' ')}
                </button>
              ))}
            </div>
          </div>

          {/* Water Features */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-3">
              <Droplets className="w-4 h-4 inline mr-1" />
              Water Features
            </label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {WATER_FEATURES.map((feature) => (
                <button
                  key={feature}
                  onClick={() => toggleArrayItem('waterFeatures', feature)}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    metadata.waterFeatures.includes(feature)
                      ? 'bg-blue-100 border-blue-300 text-blue-800'
                      : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
                  }`}
                >
                  {feature.replace('-', ' ')}
                </button>
              ))}
            </div>
          </div>

          {/* Recommended Gear */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-3">
              <Package className="w-4 h-4 inline mr-1" />
              Recommended Gear
            </label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {GEAR_OPTIONS.map((gear) => (
                <button
                  key={gear}
                  onClick={() => toggleArrayItem('gear', gear)}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    metadata.gear.includes(gear)
                      ? 'bg-orange-100 border-orange-300 text-orange-800'
                      : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
                  }`}
                >
                  {gear.replace('-', ' ')}
                </button>
              ))}
            </div>
          </div>

          {/* Best Seasons */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-3">
              <Calendar className="w-4 h-4 inline mr-1" />
              Best Seasons
            </label>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-2">
              {SEASON_OPTIONS.map((season) => (
                <button
                  key={season}
                  onClick={() => toggleArrayItem('seasons', season)}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    metadata.seasons.includes(season)
                      ? 'bg-green-100 border-green-300 text-green-800'
                      : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
                  }`}
                >
                  {season.replace('-', ' ')}
                </button>
              ))}
            </div>
          </div>

          {/* Trail Conditions */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-3">
              Trail Conditions & Notes
            </label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {CONDITION_OPTIONS.map((condition) => (
                <button
                  key={condition}
                  onClick={() => toggleArrayItem('conditions', condition)}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    metadata.conditions.includes(condition)
                      ? 'bg-purple-100 border-purple-300 text-purple-800'
                      : 'bg-white border-terrain-300 text-trail-700 hover:bg-terrain-50'
                  }`}
                >
                  {condition.replace('-', ' ')}
                </button>
              ))}
            </div>
          </div>

          {/* Custom Tags */}
          <div>
            <label className="block text-sm font-medium text-trail-700 mb-2">
              <Tag className="w-4 h-4 inline mr-1" />
              Custom Tags
            </label>
            <input
              type="text"
              placeholder="Add custom tags (comma separated)"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && e.currentTarget.value.trim()) {
                  e.preventDefault();
                  const newTags = e.currentTarget.value
                    .split(',')
                    .map(tag => tag.trim().toLowerCase())
                    .filter(tag => tag && !metadata.tags.includes(tag));
                  
                  if (newTags.length > 0) {
                    updateMetadata('tags', [...metadata.tags, ...newTags]);
                    e.currentTarget.value = '';
                  }
                }
              }}
              className="w-full px-3 py-2 border border-terrain-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-forest-500 focus:border-transparent"
            />
            
            {metadata.tags.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-2">
                {metadata.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="inline-flex items-center px-2 py-1 bg-terrain-100 text-trail-700 text-xs rounded-full"
                  >
                    {tag}
                    <button
                      onClick={() => updateMetadata('tags', metadata.tags.filter((_, i) => i !== index))}
                      className="ml-1 text-trail-500 hover:text-trail-700"
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};