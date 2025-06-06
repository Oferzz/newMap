import { CloudinaryTransformation } from '../types/hero.types';

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Cloudinary configuration - will be fetched from backend
let CLOUDINARY_CLOUD_NAME = 'demo'; // fallback
let CLOUDINARY_BASE_URL = `https://res.cloudinary.com/${CLOUDINARY_CLOUD_NAME}`;

// Fetch Cloudinary configuration from backend
const fetchCloudinaryConfig = async (): Promise<string> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/media/cloudinary/config`);
    if (!response.ok) {
      throw new Error('Failed to fetch Cloudinary config');
    }
    const data = await response.json();
    const cloudName = data.data.cloudName;
    
    // Update global configuration
    CLOUDINARY_CLOUD_NAME = cloudName;
    CLOUDINARY_BASE_URL = `https://res.cloudinary.com/${cloudName}`;
    
    return cloudName;
  } catch (error) {
    console.warn('Failed to fetch Cloudinary config, using fallback:', error);
    return CLOUDINARY_CLOUD_NAME;
  }
};

/**
 * Generate Cloudinary URL with transformations
 */
export const generateCloudinaryUrl = (
  publicId: string, 
  transformations: CloudinaryTransformation = {}
): string => {
  const {
    width,
    height,
    crop = 'fill',
    quality = 'auto',
    format = 'auto',
    gravity = 'center'
  } = transformations;

  const params: string[] = [];
  
  if (width) params.push(`w_${width}`);
  if (height) params.push(`h_${height}`);
  if (crop) params.push(`c_${crop}`);
  if (quality) params.push(`q_${quality}`);
  if (format) params.push(`f_${format}`);
  if (gravity) params.push(`g_${gravity}`);

  const transformationString = params.length > 0 ? `/${params.join(',')}` : '';
  
  return `${CLOUDINARY_BASE_URL}/image/upload${transformationString}/${publicId}`;
};

/**
 * Generate responsive srcSet for different screen sizes
 */
export const generateResponsiveSrcSet = (publicId: string): string => {
  const sizes = [
    { width: 480, suffix: '480w' },
    { width: 768, suffix: '768w' },
    { width: 1024, suffix: '1024w' },
    { width: 1440, suffix: '1440w' },
    { width: 1920, suffix: '1920w' },
    { width: 2560, suffix: '2560w' }
  ];

  return sizes
    .map(({ width, suffix }) => 
      `${generateCloudinaryUrl(publicId, { 
        width, 
        height: Math.round(width * 0.6), // 16:10 aspect ratio
        crop: 'fill',
        quality: 'auto',
        format: 'auto'
      })} ${suffix}`
    )
    .join(', ');
};

/**
 * Get optimized image URL for hero display
 */
export const getHeroImageUrl = (
  publicId: string,
  screenWidth: number = 1920,
  screenHeight: number = 1080
): string => {
  return generateCloudinaryUrl(publicId, {
    width: screenWidth,
    height: screenHeight,
    crop: 'fill',
    quality: 'auto',
    format: 'auto',
    gravity: 'center'
  });
};

/**
 * Get thumbnail URL for preloading
 */
export const getThumbnailUrl = (publicId: string): string => {
  return generateCloudinaryUrl(publicId, {
    width: 50,
    height: 30,
    crop: 'fill',
    quality: 30,
    format: 'auto'
  });
};

/**
 * Generate signed URL for private images via backend
 */
export const getSignedImageUrl = async (
  publicId: string,
  transformations: CloudinaryTransformation = {}
): Promise<string> => {
  try {
    // Ensure we have the latest Cloudinary configuration
    await fetchCloudinaryConfig();
    
    const response = await fetch(`${API_BASE_URL}/api/v1/media/cloudinary/sign`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        publicId,
        transformations
      })
    });

    if (!response.ok) {
      throw new Error('Failed to get signed URL');
    }

    const data = await response.json();
    return data.data.signedUrl;
  } catch (error) {
    console.error('Error getting signed URL:', error);
    // Fallback to unsigned URL for development
    return generateCloudinaryUrl(publicId, transformations);
  }
};

/**
 * Get signed hero image URL for private images
 */
export const getSignedHeroImageUrl = async (
  publicId: string,
  screenWidth: number = 1920,
  screenHeight: number = 1080
): Promise<string> => {
  return getSignedImageUrl(publicId, {
    width: screenWidth,
    height: screenHeight,
    crop: 'fill',
    quality: 'auto',
    format: 'auto',
    gravity: 'center'
  });
};

/**
 * Get signed thumbnail URL for private images
 */
export const getSignedThumbnailUrl = async (publicId: string): Promise<string> => {
  return getSignedImageUrl(publicId, {
    width: 50,
    height: 30,
    crop: 'fill',
    quality: 30,
    format: 'auto'
  });
};

/**
 * Fetch images from a Cloudinary folder
 */
export const fetchFolderImages = async (folderName: string, maxImages: number = 100): Promise<CloudinaryImage[]> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/media/cloudinary/list`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        folder: folderName,
        maxImages: maxImages
      })
    });

    if (!response.ok) {
      throw new Error('Failed to fetch folder images');
    }

    const data = await response.json();
    return data.data.images || [];
  } catch (error) {
    console.error('Error fetching folder images:', error);
    return [];
  }
};

// Type for Cloudinary image from backend
interface CloudinaryImage {
  publicId: string;
  format: string;
  width: number;
  height: number;
  createdAt: string;
  tags?: string[];
}