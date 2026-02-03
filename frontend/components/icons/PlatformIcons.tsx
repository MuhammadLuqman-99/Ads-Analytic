import { SVGProps } from "react";

interface IconProps extends SVGProps<SVGSVGElement> {
  className?: string;
}

// Shopee icon - stylized "S" shape similar to Shopee logo
export function ShopeeIcon({ className, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <path
        d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm3.5 14.5c-1.5 1.5-4 1.5-5.5 0L7.5 14c-.5-.5-.5-1.3 0-1.8s1.3-.5 1.8 0l2.2 2.2c.8.8 2.2.8 3 0 .4-.4.6-.9.6-1.4s-.2-1-.6-1.4L12 9.1c-1.5-1.5-1.5-4 0-5.5.7-.7 1.7-1.1 2.7-1.1s2 .4 2.8 1.1c.5.5.5 1.3 0 1.8s-1.3.5-1.8 0c-.4-.4-1-.4-1.4 0-.4.4-.4 1 0 1.4l2.5 2.5c1.5 1.5 1.5 4 0 5.5-.3.4-.5.5-.8.7z"
        fill="currentColor"
      />
    </svg>
  );
}

// Alternative Shopee icon - shopping bag with S
export function ShopeeIcon2({ className, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      {/* Shopping bag outline */}
      <path
        d="M6 6h12l1.5 12.5c.1.8-.5 1.5-1.3 1.5H5.8c-.8 0-1.4-.7-1.3-1.5L6 6z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
      {/* Bag handles */}
      <path
        d="M9 6V5a3 3 0 0 1 6 0v1"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
      {/* S letter */}
      <path
        d="M14.5 10.5c0-.8-.7-1.5-1.5-1.5h-2c-.8 0-1.5.7-1.5 1.5s.7 1.5 1.5 1.5h2c.8 0 1.5.7 1.5 1.5s-.7 1.5-1.5 1.5h-2c-.8 0-1.5-.7-1.5-1.5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
    </svg>
  );
}

// TikTok icon - musical note style
export function TikTokIcon({ className, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <path
        d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-5.2 1.74 2.89 2.89 0 0 1 2.31-4.64 2.93 2.93 0 0 1 .88.13V9.4a6.84 6.84 0 0 0-1-.05A6.33 6.33 0 0 0 5 20.1a6.34 6.34 0 0 0 10.86-4.43v-7a8.16 8.16 0 0 0 4.77 1.52v-3.4a4.85 4.85 0 0 1-1-.1z"
        fill="currentColor"
      />
    </svg>
  );
}

// Meta/Facebook icon
export function MetaIcon({ className, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <path
        d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"
        fill="currentColor"
      />
    </svg>
  );
}
