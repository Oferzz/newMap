import { HeroPhoto } from '../types/hero.types';

// Hero photos collection for the landing page
// Note: These are placeholder Cloudinary IDs. Replace with actual uploaded images.
export const heroPhotos: HeroPhoto[] = [
  {
    id: '1',
    cloudinaryId: 'samples/landscapes/nature-mountains',
    location: 'Swiss Alps',
    country: 'Switzerland',
    photographer: 'Cloudinary Sample',
    altText: 'Majestic snow-capped mountain peaks in the Swiss Alps',
    focalPoint: { x: 0.5, y: 0.3 }
  },
  {
    id: '2',
    cloudinaryId: 'samples/landscapes/beach-boat',
    location: 'Tropical Paradise',
    country: 'Maldives',
    photographer: 'Cloudinary Sample',
    altText: 'Crystal clear turquoise waters with traditional boat',
    focalPoint: { x: 0.3, y: 0.4 }
  },
  {
    id: '3',
    cloudinaryId: 'samples/landscapes/girl-urban-view',
    location: 'City Overlook',
    country: 'Norway',
    photographer: 'Cloudinary Sample',
    altText: 'Person overlooking dramatic Norwegian fjord landscape',
    focalPoint: { x: 0.7, y: 0.6 }
  },
  {
    id: '4',
    cloudinaryId: 'samples/landscapes/landscape-panorama',
    location: 'Mountain Range',
    country: 'Canada',
    photographer: 'Cloudinary Sample',
    altText: 'Panoramic view of vast mountain ranges and valleys',
    focalPoint: { x: 0.5, y: 0.4 }
  },
  {
    id: '5',
    cloudinaryId: 'samples/landscapes/architecture-signs',
    location: 'Desert Highway',
    country: 'United States',
    photographer: 'Cloudinary Sample',
    altText: 'Open road through desert landscape with dramatic sky',
    focalPoint: { x: 0.5, y: 0.5 }
  },
  {
    id: '6',
    cloudinaryId: 'samples/animals/reindeer',
    location: 'Arctic Tundra',
    country: 'Finland',
    photographer: 'Cloudinary Sample',
    altText: 'Reindeer in pristine Arctic landscape',
    focalPoint: { x: 0.4, y: 0.6 }
  }
  // TODO: Add more diverse adventure locations:
  // - Tropical rainforests
  // - Desert dunes
  // - Ocean cliffs
  // - Northern lights
  // - Canyon views
  // - Waterfalls
  // - Volcanic landscapes
];

// Helper function to get photo by current hour
export const getCurrentHeroPhoto = (): HeroPhoto => {
  const currentHour = new Date().getHours();
  const photoIndex = currentHour % heroPhotos.length;
  return heroPhotos[photoIndex];
};

// Helper function to get the next photo (for preloading)
export const getNextHeroPhoto = (): HeroPhoto => {
  const currentHour = new Date().getHours();
  const nextPhotoIndex = (currentHour + 1) % heroPhotos.length;
  return heroPhotos[nextPhotoIndex];
};