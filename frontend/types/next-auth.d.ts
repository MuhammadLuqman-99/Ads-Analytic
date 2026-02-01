import "next-auth";

declare module "next-auth" {
  interface Session {
    user: {
      id: string;
      email: string;
      name?: string | null;
      image?: string | null;
      organizationId?: string;
      role?: string;
      onboardingCompleted?: boolean;
    };
  }

  interface User {
    id: string;
    email: string;
    name?: string | null;
    organizationId?: string;
    role?: string;
    onboardingCompleted?: boolean;
    accessToken?: string;
    refreshToken?: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    id?: string;
    organizationId?: string;
    role?: string;
    onboardingCompleted?: boolean;
    accessToken?: string;
  }
}
