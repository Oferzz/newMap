import React, { useState, useEffect } from 'react';
import { MapPin, Camera } from 'lucide-react';
import { getCurrentHeroPhoto, getNextHeroPhoto } from '../../data/dynamicHeroPhotos';
import { getSignedHeroImageUrl, getSignedThumbnailUrl } from '../../services/cloudinary.service';
import { HeroPhoto } from '../../types/hero.types';

interface HeroLandingProps {
  className?: string;
}

export const HeroLanding: React.FC<HeroLandingProps> = ({ className = '' }) => {
  const [currentPhoto, setCurrentPhoto] = useState<HeroPhoto | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [imageLoaded, setImageLoaded] = useState(false);
  const [nextPhotoPreloaded, setNextPhotoPreloaded] = useState(false);
  const [imageUrl, setImageUrl] = useState<string>('');
  const [thumbnailUrl, setThumbnailUrl] = useState<string>('');

  // Initial photo loading
  useEffect(() => {
    const loadInitialPhoto = async () => {
      try {
        const photo = await getCurrentHeroPhoto();
        setCurrentPhoto(photo);
      } catch (error) {
        console.error('Failed to load initial photo:', error);
      }
    };

    loadInitialPhoto();
  }, []);

  // Load signed URLs when photo changes
  useEffect(() => {
    if (!currentPhoto) return;

    const loadSignedUrls = async () => {
      setIsLoading(true);
      try {
        const [mainUrl, thumbUrl] = await Promise.all([
          getSignedHeroImageUrl(currentPhoto.cloudinaryId, 1920, 1080),
          getSignedThumbnailUrl(currentPhoto.cloudinaryId)
        ]);
        setImageUrl(mainUrl);
        setThumbnailUrl(thumbUrl);
      } catch (error) {
        console.error('Failed to load signed URLs:', error);
      }
    };

    loadSignedUrls();
  }, [currentPhoto?.cloudinaryId]);

  // Update photo based on current hour
  useEffect(() => {
    const updatePhoto = async () => {
      try {
        const newPhoto = await getCurrentHeroPhoto();
        if (!currentPhoto || newPhoto.id !== currentPhoto.id) {
          setImageLoaded(false);
          setCurrentPhoto(newPhoto);
        }
      } catch (error) {
        console.error('Failed to update photo:', error);
      }
    };

    // Check for photo update every minute
    const interval = setInterval(updatePhoto, 60 * 1000);
    
    return () => clearInterval(interval);
  }, [currentPhoto?.id]);

  // Preload next photo
  useEffect(() => {
    if (imageLoaded && !nextPhotoPreloaded) {
      const preloadNext = async () => {
        try {
          const nextPhoto = await getNextHeroPhoto();
          const nextUrl = await getSignedHeroImageUrl(nextPhoto.cloudinaryId, 1920, 1080);
          const img = new Image();
          img.onload = () => setNextPhotoPreloaded(true);
          img.src = nextUrl;
        } catch (error) {
          console.error('Failed to preload next photo:', error);
        }
      };
      preloadNext();
    }
  }, [imageLoaded, nextPhotoPreloaded]);

  const handleImageLoad = () => {
    setImageLoaded(true);
    setIsLoading(false);
  };

  const handleImageError = () => {
    setIsLoading(false);
    console.error('Failed to load hero image:', currentPhoto?.cloudinaryId);
  };

  return (
    <div className={`relative w-full h-screen overflow-hidden ${className}`}>
      {/* Loading placeholder with blurred thumbnail */}
      {isLoading && thumbnailUrl && (
        <div className="absolute inset-0 z-10">
          <img
            src={thumbnailUrl}
            alt=""
            className="w-full h-full object-cover blur-md scale-110"
          />
          <div className="absolute inset-0 bg-gray-900/20" />
        </div>
      )}

      {/* Main hero image */}
      {imageUrl && (
        <img
          src={imageUrl}
          alt={currentPhoto?.altText || 'Adventure landscape'}
          className={`hero-image w-full h-full object-cover transition-opacity duration-1000 ${
            imageLoaded ? 'opacity-100' : 'opacity-0'
          }`}
          onLoad={handleImageLoad}
          onError={handleImageError}
          loading="eager"
        />
      )}

      {/* Dark overlay for better text readability */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-black/40" />

      {/* Photo credit */}
      {currentPhoto && (
        <div className="absolute bottom-4 right-4 z-20">
          <div className="bg-black/50 backdrop-blur-sm rounded-lg px-3 py-2 text-white text-sm">
            <div className="flex items-center gap-2 mb-1">
              <MapPin className="w-3 h-3" />
              <span className="font-medium">{currentPhoto.location}</span>
              <span className="text-white/70">•</span>
              <span className="text-white/70">{currentPhoto.country}</span>
            </div>
            <div className="flex items-center gap-1 text-xs text-white/80">
              <Camera className="w-3 h-3" />
              {currentPhoto.photographerUrl ? (
                <a 
                  href={currentPhoto.photographerUrl} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="hover:text-white transition-colors"
                >
                  {currentPhoto.photographer}
                </a>
              ) : (
                <span>{currentPhoto.photographer}</span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Subtle Ken Burns effect */}
      <style>{`
        .hero-image {
          animation: kenBurns 20s ease-in-out infinite alternate;
        }
        
        @keyframes kenBurns {
          0% {
            transform: scale(1) translate(0, 0);
          }
          100% {
            transform: scale(1.05) translate(-1%, -1%);
          }
        }
      `}</style>
    </div>
  );
};