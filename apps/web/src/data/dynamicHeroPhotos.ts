import { HeroPhoto } from '../types/hero.types';
import { fetchCollectionImages } from '../services/cloudinary.service';

// Configuration for the hero photos collection
const HERO_PHOTOS_COLLECTION = '0adfb2492c4ee8e9dbd1f35986d7f80f'; // Your Cloudinary collection ID
const MAX_HERO_IMAGES = 50; // Maximum number of images to fetch

// Cache for hero photos
let cachedHeroPhotos: HeroPhoto[] = [];
let lastFetched: number = 0;
const CACHE_DURATION = 60 * 60 * 1000; // 1 hour cache

/**
 * Fetch hero photos from Cloudinary folder
 */
export const fetchHeroPhotos = async (): Promise<HeroPhoto[]> => {
  const now = Date.now();
  
  // Return cached photos if still valid
  if (cachedHeroPhotos.length > 0 && now - lastFetched < CACHE_DURATION) {
    return cachedHeroPhotos;
  }

  try {
    // Fetch images from the collection
    const images = await fetchCollectionImages(HERO_PHOTOS_COLLECTION, MAX_HERO_IMAGES);
    
    // Convert Cloudinary images to HeroPhoto format
    const heroPhotos: HeroPhoto[] = images.map((image, index) => ({
      id: `hero-${index + 1}`,
      cloudinaryId: image.publicId,
      location: extractLocationFromTags(image.tags) || 'Unknown Location',
      country: extractCountryFromTags(image.tags) || 'Unknown Country',
      photographer: extractPhotographerFromTags(image.tags) || 'Unknown Photographer',
      altText: `Adventure photo ${index + 1}: ${extractLocationFromTags(image.tags) || 'scenic landscape'}`,
      focalPoint: { x: 0.5, y: 0.4 } // Default center focus
    }));

    // Update cache
    cachedHeroPhotos = heroPhotos;
    lastFetched = now;
    
    return heroPhotos;
  } catch (error) {
    console.error('Failed to fetch hero photos from collection:', error);
    
    // Return fallback photos if fetch fails
    return getFallbackHeroPhotos();
  }
};

/**
 * Get current hero photo based on time rotation
 */
export const getCurrentHeroPhoto = async (): Promise<HeroPhoto> => {
  const photos = await fetchHeroPhotos();
  
  if (photos.length === 0) {
    return getFallbackHeroPhotos()[0];
  }
  
  // Rotate based on current hour
  const currentHour = new Date().getHours();
  const photoIndex = currentHour % photos.length;
  return photos[photoIndex];
};

/**
 * Get next hero photo (for preloading)
 */
export const getNextHeroPhoto = async (): Promise<HeroPhoto> => {
  const photos = await fetchHeroPhotos();
  
  if (photos.length === 0) {
    return getFallbackHeroPhotos()[0];
  }
  
  // Get next hour's photo
  const currentHour = new Date().getHours();
  const nextPhotoIndex = (currentHour + 1) % photos.length;
  return photos[nextPhotoIndex];
};

/**
 * Extract location from image tags
 * Looks for tags like "location:Paris" or "loc:Paris"
 */
function extractLocationFromTags(tags?: string[]): string | null {
  if (!tags) return null;
  
  for (const tag of tags) {
    if (tag.startsWith('location:') || tag.startsWith('loc:')) {
      return tag.split(':')[1];
    }
  }
  return null;
}

/**
 * Extract country from image tags
 * Looks for tags like "country:France" or "country:France"
 */
function extractCountryFromTags(tags?: string[]): string | null {
  if (!tags) return null;
  
  for (const tag of tags) {
    if (tag.startsWith('country:')) {
      return tag.split(':')[1];
    }
  }
  return null;
}

/**
 * Extract photographer from image tags
 * Looks for tags like "photographer:John Doe" or "credit:John Doe"
 */
function extractPhotographerFromTags(tags?: string[]): string | null {
  if (!tags) return null;
  
  for (const tag of tags) {
    if (tag.startsWith('photographer:') || tag.startsWith('credit:')) {
      return tag.split(':')[1];
    }
  }
  return null;
}

/**
 * Fallback hero photos when folder fetch fails
 */
function getFallbackHeroPhotos(): HeroPhoto[] {
  return [
    {
      id: 'fallback-1',
      cloudinaryId: 'samples/landscapes/nature-mountains',
      location: 'Swiss Alps',
      country: 'Switzerland',
      photographer: 'Cloudinary Sample',
      altText: 'Majestic snow-capped mountain peaks in the Swiss Alps',
      focalPoint: { x: 0.5, y: 0.3 }
    },
    {
      id: 'fallback-2',
      cloudinaryId: 'samples/landscapes/beach-boat',
      location: 'Tropical Paradise',
      country: 'Maldives',
      photographer: 'Cloudinary Sample',
      altText: 'Crystal clear turquoise waters with traditional boat',
      focalPoint: { x: 0.3, y: 0.4 }
    }
  ];
}

/**
 * Get collection configuration (useful for setup instructions)
 */
export const getHeroPhotosConfig = () => ({
  collectionId: HERO_PHOTOS_COLLECTION,
  maxImages: MAX_HERO_IMAGES,
  tagFormat: {
    location: 'location:Location Name',
    country: 'country:Country Name', 
    photographer: 'photographer:Photographer Name'
  }
});