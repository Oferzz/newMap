import { format, parseISO, formatDistanceToNow as fDistanceToNow } from 'date-fns';

export function formatDate(date: string | Date): string {
  const dateObj = typeof date === 'string' ? parseISO(date) : date;
  return format(dateObj, 'MMM d, yyyy');
}

export function formatDateTime(date: string | Date): string {
  const dateObj = typeof date === 'string' ? parseISO(date) : date;
  return format(dateObj, 'MMM d, yyyy h:mm a');
}

export function formatDistanceToNow(date: string | Date): string {
  const dateObj = typeof date === 'string' ? parseISO(date) : date;
  return fDistanceToNow(dateObj, { addSuffix: true });
}