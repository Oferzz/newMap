export interface HeroPhoto {
  id: string;
  cloudinaryId: string; // Cloudinary public ID
  location: string;
  country: string;
  photographer: string;
  photographerUrl?: string;
  altText: string;
  focalPoint?: { x: number; y: number }; // For responsive cropping (0-1 range)
}

export interface CloudinaryTransformation {
  width?: number;
  height?: number;
  crop?: 'fill' | 'fit' | 'scale' | 'crop';
  quality?: 'auto' | number;
  format?: 'auto' | 'webp' | 'jpg' | 'png';
  gravity?: 'center' | 'face' | 'faces' | 'custom';
}