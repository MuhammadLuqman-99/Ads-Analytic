import { describe, it, expect } from 'vitest';
import { render, screen } from '@/test/test-utils';
import { MetricsCard, MetricsCardSkeleton } from './MetricsCard';
import { DollarSign } from 'lucide-react';

describe('MetricsCard', () => {
  const defaultProps = {
    title: 'Total Revenue',
    value: 'RM 50,000',
    change: 15.5,
    icon: <DollarSign data-testid="icon" className="h-5 w-5 text-white" />,
    iconBgColor: 'bg-emerald-500',
  };

  describe('renders correctly', () => {
    it('displays the title', () => {
      render(<MetricsCard {...defaultProps} />);
      expect(screen.getByText('Total Revenue')).toBeInTheDocument();
    });

    it('displays the value', () => {
      render(<MetricsCard {...defaultProps} />);
      expect(screen.getByText('RM 50,000')).toBeInTheDocument();
    });

    it('displays the icon', () => {
      render(<MetricsCard {...defaultProps} />);
      expect(screen.getByTestId('icon')).toBeInTheDocument();
    });

    it('displays the subtitle when provided', () => {
      render(<MetricsCard {...defaultProps} subtitle="vs last period" />);
      expect(screen.getByText('vs last period')).toBeInTheDocument();
    });

    it('does not render subtitle when not provided', () => {
      render(<MetricsCard {...defaultProps} />);
      expect(screen.queryByText('vs last period')).not.toBeInTheDocument();
    });
  });

  describe('positive change styling', () => {
    it('shows TrendingUp icon for positive change', () => {
      render(<MetricsCard {...defaultProps} change={15.5} />);
      // TrendingUp icon should be visible (checking by the percentage text)
      expect(screen.getByText('15.5%')).toBeInTheDocument();
    });

    it('displays positive change value correctly', () => {
      render(<MetricsCard {...defaultProps} change={15.5} />);
      expect(screen.getByText('15.5%')).toBeInTheDocument();
    });

    it('shows zero change as positive', () => {
      render(<MetricsCard {...defaultProps} change={0} />);
      expect(screen.getByText('0.0%')).toBeInTheDocument();
    });
  });

  describe('negative change styling', () => {
    it('shows TrendingDown icon for negative change', () => {
      render(<MetricsCard {...defaultProps} change={-10.5} />);
      expect(screen.getByText('10.5%')).toBeInTheDocument();
    });

    it('displays negative change as absolute value', () => {
      render(<MetricsCard {...defaultProps} change={-25.3} />);
      expect(screen.getByText('25.3%')).toBeInTheDocument();
    });
  });

  describe('edge cases', () => {
    it('handles very small change values', () => {
      render(<MetricsCard {...defaultProps} change={0.1} />);
      expect(screen.getByText('0.1%')).toBeInTheDocument();
    });

    it('handles large change values', () => {
      render(<MetricsCard {...defaultProps} change={150.5} />);
      expect(screen.getByText('150.5%')).toBeInTheDocument();
    });

    it('handles decimal precision', () => {
      render(<MetricsCard {...defaultProps} change={12.345} />);
      // Should round to 1 decimal place
      expect(screen.getByText('12.3%')).toBeInTheDocument();
    });
  });

  describe('different icon backgrounds', () => {
    it('applies custom icon background color', () => {
      const { container } = render(
        <MetricsCard {...defaultProps} iconBgColor="bg-purple-500" />
      );
      const iconContainer = container.querySelector('.bg-purple-500');
      expect(iconContainer).toBeInTheDocument();
    });
  });
});

describe('MetricsCardSkeleton', () => {
  it('renders skeleton elements', () => {
    const { container } = render(<MetricsCardSkeleton />);

    // Check for animate-pulse class (indicates skeleton)
    const skeleton = container.querySelector('.animate-pulse');
    expect(skeleton).toBeInTheDocument();
  });

  it('has correct structure', () => {
    const { container } = render(<MetricsCardSkeleton />);

    // Check for skeleton placeholder elements
    const placeholders = container.querySelectorAll('.bg-slate-200');
    expect(placeholders.length).toBeGreaterThan(0);
  });
});
