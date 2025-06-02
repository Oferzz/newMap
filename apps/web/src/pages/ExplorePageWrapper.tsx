import React, { useState } from 'react';
import { Header } from '../components/layout/Header';
import { ExplorePage } from './ExplorePage';

type ContentType = 'all' | 'trips' | 'places';

export const ExplorePageWrapper: React.FC = () => {
  const [contentType, setContentType] = useState<ContentType>('all');

  return (
    <>
      <Header />
      
      {/* Content Type Toggles - Below header */}
      <div className="fixed top-16 left-0 right-0 bg-terrain-100 border-b border-terrain-300 z-40">
        <div className="flex justify-center pt-3 pb-2">
          <div className="flex items-center gap-6">
            {(['all', 'trips', 'places'] as ContentType[]).map((type) => (
              <button
                key={type}
                onClick={() => setContentType(type)}
                className={`px-3 py-2 text-sm font-medium rounded-lg transition-colors capitalize ${
                  contentType === type
                    ? 'text-trail-800 bg-terrain-200'
                    : 'text-trail-700 hover:text-trail-800 hover:bg-terrain-200'
                }`}
              >
                {type}
              </button>
            ))}
          </div>
        </div>
      </div>
      
      <ExplorePage 
        contentType={contentType}
        onContentTypeChange={setContentType}
      />
    </>
  );
};