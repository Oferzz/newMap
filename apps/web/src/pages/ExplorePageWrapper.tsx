import React, { useState } from 'react';
import { Header } from '../components/layout/Header';
import { ExplorePage } from './ExplorePage';

type ContentType = 'all' | 'trips' | 'places';

export const ExplorePageWrapper: React.FC = () => {
  const [contentType, setContentType] = useState<ContentType>('all');

  return (
    <>
      <Header 
        contentType={contentType} 
        onContentTypeChange={setContentType}
      />
      <ExplorePage 
        contentType={contentType}
        onContentTypeChange={setContentType}
      />
    </>
  );
};