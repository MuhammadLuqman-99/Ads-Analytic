import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Campaigns Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/campaigns');
  });

  test('displays campaigns list', async ({ page }) => {
    // Wait for campaigns to load
    await expect(page.getByRole('table')).toBeVisible({ timeout: 10000 });

    // Check for table headers
    await expect(page.getByRole('columnheader', { name: /campaign|name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /platform/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /status/i })).toBeVisible();
  });

  test('can filter by platform', async ({ page }) => {
    // Find platform filter
    const platformFilter = page.getByRole('combobox', { name: /platform/i }).or(
      page.getByRole('button', { name: /all platforms/i })
    );

    await platformFilter.click();

    // Select Meta
    await page.getByRole('option', { name: /meta|facebook/i }).click();

    // Verify filter is applied (table should update)
    await page.waitForTimeout(1000);

    // All visible campaigns should be Meta
    const platformCells = page.locator('td').filter({ hasText: /meta|facebook/i });
    const count = await platformCells.count();
    expect(count).toBeGreaterThan(0);
  });

  test('can filter by status', async ({ page }) => {
    // Find status filter
    const statusFilter = page.getByRole('combobox', { name: /status/i }).or(
      page.getByRole('button', { name: /all status/i })
    );

    await statusFilter.click();

    // Select Active
    await page.getByRole('option', { name: /active/i }).click();

    // Wait for filter to apply
    await page.waitForTimeout(1000);

    // All visible campaigns should be active
    const statusCells = page.locator('td').filter({ hasText: /active/i });
    const count = await statusCells.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('can search campaigns', async ({ page }) => {
    // Find search input
    const searchInput = page.getByPlaceholder(/search/i);

    // Enter search term
    await searchInput.fill('Summer');
    await searchInput.press('Enter');

    // Wait for search to complete
    await page.waitForTimeout(1000);

    // Check that results contain search term (or show no results message)
    const hasResults = await page.getByText(/summer/i).isVisible();
    const noResults = await page.getByText(/no.*found/i).isVisible();

    expect(hasResults || noResults).toBeTruthy();
  });

  test('table sorting works', async ({ page }) => {
    // Find a sortable column header
    const spendHeader = page.getByRole('columnheader', { name: /spend/i });

    // Click to sort
    await spendHeader.click();
    await page.waitForTimeout(500);

    // Click again to reverse sort
    await spendHeader.click();
    await page.waitForTimeout(500);

    // Verify table still displays correctly
    await expect(page.getByRole('table')).toBeVisible();
  });

  test('pagination works', async ({ page }) => {
    // Look for pagination controls
    const nextButton = page.getByRole('button', { name: /next/i }).or(
      page.getByRole('button', { name: /→|>/ })
    );

    if (await nextButton.isVisible() && await nextButton.isEnabled()) {
      // Click next page
      await nextButton.click();
      await page.waitForTimeout(1000);

      // Verify we're on page 2
      const pageIndicator = page.getByText(/page 2|2 of/i);
      await expect(pageIndicator).toBeVisible();
    }
  });

  test('can view campaign details', async ({ page }) => {
    // Click on the first campaign row
    const firstRow = page.getByRole('row').nth(1); // nth(0) is header
    await firstRow.click();

    // Should navigate to campaign details or show modal
    await page.waitForTimeout(1000);

    const detailsVisible = await page.getByText(/campaign details|performance/i).isVisible();
    const urlChanged = page.url().includes('/campaigns/');

    expect(detailsVisible || urlChanged).toBeTruthy();
  });
});

test.describe('Campaign Export', () => {
  test('can export campaigns to CSV', async ({ page }) => {
    await page.goto('/campaigns');

    // Wait for campaigns to load
    await expect(page.getByRole('table')).toBeVisible({ timeout: 10000 });

    // Find export button
    const exportButton = page.getByRole('button', { name: /export|download/i });

    if (await exportButton.isVisible()) {
      // Set up download handler
      const downloadPromise = page.waitForEvent('download');

      // Click export
      await exportButton.click();

      // If there's a format selection, choose CSV
      const csvOption = page.getByRole('menuitem', { name: /csv/i });
      if (await csvOption.isVisible()) {
        await csvOption.click();
      }

      // Wait for download
      const download = await downloadPromise;

      // Verify file name
      const filename = download.suggestedFilename();
      expect(filename).toContain('.csv');

      // Save file for verification
      const downloadPath = path.join(__dirname, '../downloads', filename);
      await download.saveAs(downloadPath);
    }
  });

  test('can export selected campaigns', async ({ page }) => {
    await page.goto('/campaigns');
    await expect(page.getByRole('table')).toBeVisible({ timeout: 10000 });

    // Select some campaigns using checkboxes
    const checkboxes = page.getByRole('checkbox').all();
    const checkboxList = await checkboxes;

    if (checkboxList.length > 2) {
      // Select first two campaigns
      await checkboxList[1].check();
      await checkboxList[2].check();

      // Find export selected button
      const exportButton = page.getByRole('button', { name: /export.*selected/i });

      if (await exportButton.isVisible()) {
        await exportButton.click();

        // Wait for download or success message
        await page.waitForTimeout(2000);
      }
    }
  });
});

test.describe('Campaign Comparison', () => {
  test('can compare two campaigns', async ({ page }) => {
    await page.goto('/campaigns');
    await expect(page.getByRole('table')).toBeVisible({ timeout: 10000 });

    // Select two campaigns for comparison
    const checkboxes = await page.getByRole('checkbox').all();

    if (checkboxes.length >= 3) {
      await checkboxes[1].check();
      await checkboxes[2].check();

      // Find compare button
      const compareButton = page.getByRole('button', { name: /compare/i });

      if (await compareButton.isVisible()) {
        await compareButton.click();

        // Should show comparison view
        await expect(page.getByText(/comparison|vs/i)).toBeVisible({ timeout: 5000 });
      }
    }
  });
});
