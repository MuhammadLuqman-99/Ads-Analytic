import React, { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Create a new QueryClient for each test
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });

interface AllTheProvidersProps {
  children: React.ReactNode;
}

const AllTheProviders: React.FC<AllTheProvidersProps> = ({ children }) => {
  const queryClient = createTestQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options });

// Re-export everything
export * from '@testing-library/react';
export { customRender as render };

// Mock data generators
export const mockCampaign = (overrides = {}) => ({
  id: 'camp-1',
  name: 'Test Campaign',
  platform: 'meta',
  status: 'active',
  spend: 1000,
  revenue: 5000,
  impressions: 100000,
  clicks: 5000,
  conversions: 250,
  roas: 5.0,
  ctr: 5.0,
  cpc: 0.2,
  ...overrides,
});

export const mockDashboardSummary = (overrides = {}) => ({
  totals: {
    spend: 10000,
    revenue: 50000,
    impressions: 1000000,
    clicks: 50000,
    conversions: 2500,
    roas: 5.0,
  },
  changes: {
    spend: 12.5,
    revenue: 15.0,
    impressions: 8.3,
    clicks: 10.2,
    conversions: 18.0,
    roas: 2.2,
  },
  byPlatform: [],
  byDate: [],
  ...overrides,
});

export const mockPlatformMetrics = (platform: string, overrides = {}) => ({
  platform,
  spend: 3000,
  revenue: 15000,
  impressions: 300000,
  clicks: 15000,
  conversions: 750,
  roas: 5.0,
  ...overrides,
});
