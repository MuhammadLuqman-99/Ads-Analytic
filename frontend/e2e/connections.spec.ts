import { test, expect } from '@playwright/test';

test.describe('Platform Connections', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to connections/settings page
    await page.goto('/settings/connections');
  });

  test('displays available platforms', async ({ page }) => {
    // Check for platform connection options
    await expect(page.getByText(/meta|facebook/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/tiktok/i)).toBeVisible();
    await expect(page.getByText(/shopee/i)).toBeVisible();
  });

  test('shows connected accounts', async ({ page }) => {
    // Look for any connected accounts section
    const connectedSection = page.locator('text=/connected|active/i').first();
    await expect(connectedSection).toBeVisible();
  });

  test('connect Meta button initiates OAuth flow', async ({ page }) => {
    // Find the Meta connect button
    const metaButton = page.getByRole('button', { name: /connect.*meta|connect.*facebook/i }).first();

    if (await metaButton.isVisible()) {
      // Click connect button
      await metaButton.click();

      // Should redirect to OAuth or show modal
      await page.waitForTimeout(2000);

      // Check if we got redirected to Facebook OAuth or a modal appeared
      const currentUrl = page.url();
      const modalVisible = await page.locator('[role="dialog"]').isVisible();

      expect(currentUrl.includes('facebook.com') || modalVisible).toBeTruthy();
    }
  });

  test('connect TikTok button initiates OAuth flow', async ({ page }) => {
    const tiktokButton = page.getByRole('button', { name: /connect.*tiktok/i }).first();

    if (await tiktokButton.isVisible()) {
      await tiktokButton.click();
      await page.waitForTimeout(2000);

      const currentUrl = page.url();
      const modalVisible = await page.locator('[role="dialog"]').isVisible();

      expect(currentUrl.includes('tiktok.com') || modalVisible).toBeTruthy();
    }
  });

  test('can disconnect a connected account', async ({ page }) => {
    // Find a disconnect button
    const disconnectButton = page.getByRole('button', { name: /disconnect|remove/i }).first();

    if (await disconnectButton.isVisible()) {
      await disconnectButton.click();

      // Confirm disconnection (if confirmation dialog appears)
      const confirmButton = page.getByRole('button', { name: /confirm|yes/i });
      if (await confirmButton.isVisible()) {
        await confirmButton.click();
      }

      // Wait for update
      await page.waitForTimeout(2000);

      // Verify disconnection (button text might change)
      await expect(page.getByText(/disconnected|removed|success/i)).toBeVisible({ timeout: 5000 });
    }
  });

  test('shows sync status for connected accounts', async ({ page }) => {
    // Look for sync status indicators
    const syncStatus = page.locator('text=/last synced|syncing|sync status/i').first();
    await expect(syncStatus).toBeVisible();
  });

  test('can trigger manual sync for a platform', async ({ page }) => {
    // Find a sync button for a specific platform
    const syncButton = page.getByRole('button', { name: /sync.*now|refresh/i }).first();

    if (await syncButton.isVisible()) {
      await syncButton.click();

      // Check for loading state or success message
      await page.waitForTimeout(3000);

      // Should show syncing or success state
      const syncing = await page.getByText(/syncing/i).isVisible();
      const success = await page.getByText(/synced|success/i).isVisible();

      expect(syncing || success).toBeTruthy();
    }
  });
});

test.describe('OAuth Callback', () => {
  test('handles successful OAuth callback', async ({ page }) => {
    // Simulate a successful OAuth callback
    await page.goto('/oauth/callback?platform=meta&code=test_auth_code&state=valid_state');

    // Should redirect to connections page with success message
    await page.waitForURL(/settings|connections|dashboard/, { timeout: 10000 });

    // Or show success message
    const successMessage = page.getByText(/connected|success/i);
    await expect(successMessage).toBeVisible({ timeout: 5000 });
  });

  test('handles OAuth error', async ({ page }) => {
    // Simulate an OAuth error
    await page.goto('/oauth/callback?platform=meta&error=access_denied');

    // Should show error message
    await expect(page.getByText(/error|denied|failed/i)).toBeVisible({ timeout: 5000 });
  });
});
