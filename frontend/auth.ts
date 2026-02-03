import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import Google from "next-auth/providers/google";
import { z } from "zod";

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
});

export const { handlers, signIn, signOut, auth } = NextAuth({
  pages: {
    signIn: "/login",
    newUser: "/onboarding",
    error: "/login",
  },
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        token.id = user.id;
        token.email = user.email;
        token.name = user.name;
        token.organizationId = (user as any).organizationId;
        token.role = (user as any).role;
        token.onboardingCompleted = (user as any).onboardingCompleted;
      }
      if (account) {
        token.accessToken = account.access_token;
      }
      return token;
    },
    async session({ session, token }) {
      if (token) {
        session.user.id = token.id as string;
        session.user.organizationId = token.organizationId as string;
        session.user.role = token.role as string;
        session.user.onboardingCompleted = token.onboardingCompleted as boolean;
      }
      return session;
    },
    async redirect({ url, baseUrl }) {
      // Handle post-login redirect
      if (url.startsWith("/")) return `${baseUrl}${url}`;
      if (new URL(url).origin === baseUrl) return url;
      return baseUrl;
    },
  },
  providers: [
    Google({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
      authorization: {
        params: {
          prompt: "consent",
          access_type: "offline",
          response_type: "code",
        },
      },
    }),
    Credentials({
      id: "credentials",
      name: "Email & Password",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        console.log("[Auth] authorize called");

        // Validate credentials
        if (!credentials?.email || !credentials?.password) {
          console.log("[Auth] Missing credentials");
          return null;
        }

        const email = credentials.email as string;
        const password = credentials.password as string;

        // Validate format
        if (password.length < 6) {
          console.log("[Auth] Password too short");
          return null;
        }

        try {
          // Call backend API to authenticate
          const apiUrl = process.env.BACKEND_INTERNAL_URL || "http://api:8080";
          console.log("[Auth] Calling API at:", apiUrl);

          const response = await fetch(
            `${apiUrl}/api/v1/auth/login`,
            {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ email, password }),
            }
          );

          console.log("[Auth] API response status:", response.status);

          if (!response.ok) {
            console.log("[Auth] API returned error");
            return null;
          }

          const json = await response.json();
          console.log("[Auth] API success:", json.success);

          if (!json.success || !json.data?.user) {
            console.log("[Auth] Invalid response structure");
            return null;
          }

          const userData = json.data.user;
          const orgData = json.data.organization;

          const user = {
            id: userData.id,
            email: userData.email,
            name: [userData.firstName, userData.lastName].filter(Boolean).join(" ") || userData.email,
            organizationId: orgData?.id || null,
            role: "admin",
            onboardingCompleted: true,
          };

          console.log("[Auth] Returning user:", user.id);
          return user;
        } catch (error) {
          console.error("[Auth] Error:", error);
          return null;
        }
      },
    }),
  ],
  session: {
    strategy: "jwt",
    maxAge: 7 * 24 * 60 * 60, // 7 days
  },
  secret: process.env.NEXTAUTH_SECRET,
});
