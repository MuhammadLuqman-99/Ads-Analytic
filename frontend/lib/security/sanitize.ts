/**
 * Frontend Security Utilities
 * Input sanitization and XSS prevention
 */

// HTML entities map for escaping
const HTML_ENTITIES: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#x27;',
  '/': '&#x2F;',
  '`': '&#x60;',
  '=': '&#x3D;',
};

/**
 * Escape HTML special characters to prevent XSS
 */
export function escapeHtml(str: string): string {
  if (typeof str !== 'string') {
    return '';
  }
  return str.replace(/[&<>"'`=/]/g, (char) => HTML_ENTITIES[char] || char);
}

/**
 * Sanitize user input by removing potentially dangerous characters
 */
export function sanitizeInput(input: string): string {
  if (typeof input !== 'string') {
    return '';
  }

  // Remove null bytes
  let sanitized = input.replace(/\0/g, '');

  // Remove control characters (except newline, tab, carriage return)
  sanitized = sanitized.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '');

  // Trim whitespace
  sanitized = sanitized.trim();

  return sanitized;
}

/**
 * Validate and sanitize email input
 */
export function sanitizeEmail(email: string): string {
  if (typeof email !== 'string') {
    return '';
  }

  // Basic sanitization
  const sanitized = sanitizeInput(email).toLowerCase();

  // Email validation regex
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  return emailRegex.test(sanitized) ? sanitized : '';
}

/**
 * Sanitize URL input
 */
export function sanitizeUrl(url: string): string {
  if (typeof url !== 'string') {
    return '';
  }

  const sanitized = sanitizeInput(url);

  // Only allow http, https, and relative URLs
  if (sanitized.startsWith('http://') ||
      sanitized.startsWith('https://') ||
      sanitized.startsWith('/')) {
    return sanitized;
  }

  // Block dangerous protocols
  const dangerousProtocols = ['javascript:', 'data:', 'vbscript:', 'file:'];
  const lowerUrl = sanitized.toLowerCase();

  for (const protocol of dangerousProtocols) {
    if (lowerUrl.startsWith(protocol)) {
      return '';
    }
  }

  // If no protocol, assume relative URL
  return sanitized;
}

/**
 * Sanitize filename to prevent path traversal
 */
export function sanitizeFilename(filename: string): string {
  if (typeof filename !== 'string') {
    return '';
  }

  // Remove path traversal attempts
  let sanitized = filename.replace(/\.\./g, '');
  sanitized = sanitized.replace(/[\/\\]/g, '');

  // Remove null bytes and control characters
  sanitized = sanitizeInput(sanitized);

  // Only allow alphanumeric, dash, underscore, and dot
  sanitized = sanitized.replace(/[^a-zA-Z0-9._-]/g, '_');

  return sanitized;
}

/**
 * Sanitize JSON input to prevent prototype pollution
 */
export function sanitizeJson<T>(json: unknown): T | null {
  if (json === null || json === undefined) {
    return null;
  }

  // Re-parse to remove any prototype pollution attempts
  try {
    const str = JSON.stringify(json);
    const parsed = JSON.parse(str);

    // Remove __proto__ and constructor properties
    removePrototypePollution(parsed);

    return parsed as T;
  } catch {
    return null;
  }
}

/**
 * Recursively remove prototype pollution attempts
 */
function removePrototypePollution(obj: unknown): void {
  if (typeof obj !== 'object' || obj === null) {
    return;
  }

  const dangerousKeys = ['__proto__', 'constructor', 'prototype'];

  for (const key of Object.keys(obj as Record<string, unknown>)) {
    if (dangerousKeys.includes(key)) {
      delete (obj as Record<string, unknown>)[key];
    } else {
      removePrototypePollution((obj as Record<string, unknown>)[key]);
    }
  }
}

/**
 * Validate numeric input
 */
export function sanitizeNumber(value: unknown, options?: {
  min?: number;
  max?: number;
  integer?: boolean;
  defaultValue?: number;
}): number {
  const { min, max, integer, defaultValue = 0 } = options || {};

  let num = typeof value === 'number' ? value : parseFloat(String(value));

  if (isNaN(num) || !isFinite(num)) {
    return defaultValue;
  }

  if (integer) {
    num = Math.floor(num);
  }

  if (min !== undefined && num < min) {
    num = min;
  }

  if (max !== undefined && num > max) {
    num = max;
  }

  return num;
}

/**
 * Sanitize search query
 */
export function sanitizeSearchQuery(query: string): string {
  if (typeof query !== 'string') {
    return '';
  }

  // Remove potentially dangerous SQL/NoSQL characters
  let sanitized = query.replace(/['";\\]/g, '');

  // Remove MongoDB operators
  sanitized = sanitized.replace(/\$[a-zA-Z]+/g, '');

  // Basic sanitization
  sanitized = sanitizeInput(sanitized);

  // Limit length
  return sanitized.slice(0, 200);
}

/**
 * Create a safe innerHTML setter using DOMPurify-like logic
 * Note: For production, use DOMPurify library
 */
export function createSafeHtml(html: string): string {
  if (typeof html !== 'string') {
    return '';
  }

  // Remove script tags
  let safe = html.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');

  // Remove event handlers
  safe = safe.replace(/\s*on\w+\s*=\s*(['"])[^'"]*\1/gi, '');
  safe = safe.replace(/\s*on\w+\s*=\s*[^\s>]+/gi, '');

  // Remove javascript: URLs
  safe = safe.replace(/javascript:/gi, '');

  // Remove data: URLs in src/href
  safe = safe.replace(/(src|href)\s*=\s*(['"])data:[^'"]*\2/gi, '$1=""');

  // Remove base64 encoded scripts
  safe = safe.replace(/data:text\/html[^"']*/gi, '');

  return safe;
}

/**
 * Validate UUID format
 */
export function isValidUuid(uuid: string): boolean {
  if (typeof uuid !== 'string') {
    return false;
  }

  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  return uuidRegex.test(uuid);
}

/**
 * Validate date string format
 */
export function isValidDateString(date: string): boolean {
  if (typeof date !== 'string') {
    return false;
  }

  const parsed = new Date(date);
  return !isNaN(parsed.getTime());
}

/**
 * Rate limit function calls (client-side throttle)
 */
export function createRateLimiter(
  maxCalls: number,
  windowMs: number
): () => boolean {
  const calls: number[] = [];

  return () => {
    const now = Date.now();
    const windowStart = now - windowMs;

    // Remove old calls
    while (calls.length > 0 && calls[0] < windowStart) {
      calls.shift();
    }

    if (calls.length >= maxCalls) {
      return false; // Rate limited
    }

    calls.push(now);
    return true; // Allowed
  };
}
