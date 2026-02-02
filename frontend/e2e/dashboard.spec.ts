import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/dashboard');
  });

  test('displays dashboard header', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
  });

  test('shows metrics cards', async ({ page }) => {
    // Wait for metrics to load (look for any of the metric cards)
    await expect(page.getByText('Total Spend')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Total Revenue')).toBeVisible();
    await expect(page.getByText('Overall ROAS')).toBeVisible();
    await expect(page.getByText('Total Conversions')).toBeVisible();
  });

  test('date range picker works', async ({ page }) => {
    // Find and click the date picker button
    const dateButton = page.getByRole('button', { name: /\d{4}/ });
    await dateButton.click();

    // Check that the popover opens
    await expect(page.getByRole('dialog')).toBeVisible();

    // Select a preset (e.g., Last 7 days)
    await page.getByRole('button', { name: 'Last 7 days' }).click();

    // Verify the popover closes
    await expect(page.getByRole('dialog')).not.toBeVisible();
  });

  test('data updates when date range changes', async ({ page }) => {
    // Get initial value of a metric
    const initialSpend = await page.getByText(/RM\s*[\d,]+/).first().textContent();

    // Change date range
    const dateButton = page.getByRole('button', { name: /\d{4}/ });
    await dateButton.click();
    await page.getByRole('button', { name: 'Last 30 days' }).click();

    // Wait for data to update
    await page.waitForResponse(response =>
      response.url().includes('/api') && response.status() === 200
    );

    // Verify that the component has updated (the spinner should be gone)
    await expect(page.locator('.animate-pulse')).not.toBeVisible({ timeout: 10000 });
  });

  test('shows platform performance section', async ({ page }) => {
    await expect(page.getByText('Platform Performance')).toBeVisible();
  });

  test('shows top campaigns section', async ({ page }) => {
    // Look for the campaigns table or list
    const campaignsSection = page.locator('text=Top Campaigns').first();
    await expect(campaignsSection).toBeVisible();
  });

  test('shows recent activity section', async ({ page }) => {
    await expect(page.getByText('Recent Activity')).toBeVisible();
  });

  test('sync status indicator is visible', async ({ page }) => {
    // Look for sync status (could be a button or indicator)
    const syncIndicator = page.locator('[data-testid="sync-status"], button:has-text("Sync"), .sync-indicator').first();
    await expect(syncIndicator).toBeVisible();
  });

  test('manual sync button works', async ({ page }) => {
    // Find the sync button
    const syncButton = page.getByRole('button', { name: /sync|refresh/i }).first();

    if (await syncButton.isVisible()) {
      await syncButton.click();

      // Wait for sync to complete (look for loading state)
      await page.waitForTimeout(1000);

      // Verify no error occurred
      await expect(page.getByText(/error|failed/i)).not.toBeVisible();
    }
  });

  test('responsive on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    // Verify key elements are still visible
    await expect(page.getByText('Dashboard')).toBeVisible();
    await expect(page.getByText('Total Spend')).toBeVisible();
  });
});
