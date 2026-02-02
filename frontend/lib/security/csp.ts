/**
 * Content Security Policy Configuration
 * CSP headers for Next.js frontend
 */

export interface CSPDirectives {
  'default-src': string[];
  'script-src': string[];
  'style-src': string[];
  'img-src': string[];
  'font-src': string[];
  'connect-src': string[];
  'frame-src': string[];
  'object-src': string[];
  'base-uri': string[];
  'form-action': string[];
  'frame-ancestors': string[];
  'upgrade-insecure-requests'?: boolean;
}

/**
 * Development CSP directives (more permissive)
 */
export const developmentCSP: CSPDirectives = {
  'default-src': ["'self'"],
  'script-src': [
    "'self'",
    "'unsafe-inline'", // Required for Next.js in development
    "'unsafe-eval'",   // Required for React hot reload
  ],
  'style-src': [
    "'self'",
    "'unsafe-inline'", // Required for styled-components/emotion
  ],
  'img-src': [
    "'self'",
    'data:',
    'blob:',
    'https://*.facebook.com',
    'https://*.tiktok.com',
    'https://*.shopee.com',
    'https://*.googleusercontent.com',
  ],
  'font-src': [
    "'self'",
    'data:',
    'https://fonts.gstatic.com',
  ],
  'connect-src': [
    "'self'",
    'ws://localhost:*',   // WebSocket for HMR
    'wss://localhost:*',
    'http://localhost:*', // Local API
    'https://*.facebook.com',
    'https://*.tiktok.com',
    'https://*.shopee.com',
  ],
  'frame-src': [
    "'self'",
    'https://*.facebook.com',
    'https://*.tiktok.com',
  ],
  'object-src': ["'none'"],
  'base-uri': ["'self'"],
  'form-action': ["'self'"],
  'frame-ancestors': ["'none'"],
};

/**
 * Production CSP directives (strict)
 */
export const productionCSP: CSPDirectives = {
  'default-src': ["'self'"],
  'script-src': [
    "'self'",
    // Add nonce for inline scripts in production
    // "'nonce-{NONCE}'" - to be replaced at runtime
  ],
  'style-src': [
    "'self'",
    "'unsafe-inline'", // Often needed for CSS-in-JS
  ],
  'img-src': [
    "'self'",
    'data:',
    'blob:',
    'https://*.facebook.com',
    'https://*.fbcdn.net',
    'https://*.tiktok.com',
    'https://*.shopee.com',
    'https://*.googleusercontent.com',
  ],
  'font-src': [
    "'self'",
    'https://fonts.gstatic.com',
  ],
  'connect-src': [
    "'self'",
    'https://api.adsanalytic.com', // Production API
    'https://*.facebook.com',
    'https://*.tiktok.com',
    'https://*.shopee.com',
  ],
  'frame-src': [
    "'self'",
    'https://*.facebook.com',
    'https://*.tiktok.com',
  ],
  'object-src': ["'none'"],
  'base-uri': ["'self'"],
  'form-action': ["'self'"],
  'frame-ancestors': ["'none'"],
  'upgrade-insecure-requests': true,
};

/**
 * Convert CSP directives to header string
 */
export function cspToString(directives: CSPDirectives): string {
  const parts: string[] = [];

  for (const [directive, values] of Object.entries(directives)) {
    if (directive === 'upgrade-insecure-requests') {
      if (values === true) {
        parts.push('upgrade-insecure-requests');
      }
    } else if (Array.isArray(values)) {
      parts.push(`${directive} ${values.join(' ')}`);
    }
  }

  return parts.join('; ');
}

/**
 * Generate CSP header with optional nonce
 */
export function generateCSPHeader(nonce?: string): string {
  const isProd = process.env.NODE_ENV === 'production';
  const directives = isProd ? { ...productionCSP } : { ...developmentCSP };

  // Add nonce to script-src if provided
  if (nonce && isProd) {
    directives['script-src'] = [
      "'self'",
      `'nonce-${nonce}'`,
    ];
  }

  return cspToString(directives);
}

/**
 * Generate a cryptographically secure nonce
 */
export function generateNonce(): string {
  if (typeof window === 'undefined') {
    // Server-side
    const crypto = require('crypto');
    return crypto.randomBytes(16).toString('base64');
  } else {
    // Client-side (shouldn't normally be called)
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array));
  }
}

/**
 * Security headers for Next.js
 */
export const securityHeaders = [
  {
    key: 'X-DNS-Prefetch-Control',
    value: 'on',
  },
  {
    key: 'Strict-Transport-Security',
    value: 'max-age=63072000; includeSubDomains; preload',
  },
  {
    key: 'X-XSS-Protection',
    value: '1; mode=block',
  },
  {
    key: 'X-Frame-Options',
    value: 'DENY',
  },
  {
    key: 'X-Content-Type-Options',
    value: 'nosniff',
  },
  {
    key: 'Referrer-Policy',
    value: 'strict-origin-when-cross-origin',
  },
  {
    key: 'Permissions-Policy',
    value: 'camera=(), microphone=(), geolocation=(), interest-cohort=()',
  },
];

/**
 * Get all security headers including CSP
 */
export function getSecurityHeaders(nonce?: string) {
  return [
    ...securityHeaders,
    {
      key: 'Content-Security-Policy',
      value: generateCSPHeader(nonce),
    },
  ];
}
