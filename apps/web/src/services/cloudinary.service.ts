import { CloudinaryTransformation } from '../types/hero.types';

// Cloudinary configuration
const CLOUDINARY_CLOUD_NAME = import.meta.env.VITE_CLOUDINARY_CLOUD_NAME || 'demo';
const CLOUDINARY_BASE_URL = `https://res.cloudinary.com/${CLOUDINARY_CLOUD_NAME}`;

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