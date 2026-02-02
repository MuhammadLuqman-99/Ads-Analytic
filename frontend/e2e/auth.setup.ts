import { test as setup, expect } from '@playwright/test';
import path from 'path';

const authFile = path.join(__dirname, '../.playwright/.auth/user.json');

setup('authenticate', async ({ page }) => {
  // Navigate to login page
  await page.goto('/login');

  // Fill in login form
  await page.getByLabel('Email').fill('test@example.com');
  await page.getByLabel('Password').fill('TestPassword123!');

  // Submit login form
  await page.getByRole('button', { name: /sign in|log in/i }).click();

  // Wait for redirect to dashboard (indicates successful login)
  await page.waitForURL('/dashboard', { timeout: 30000 });

  // Verify we're logged in by checking for dashboard elements
  await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();

  // Save authentication state
  await page.context().storageState({ path: authFile });
});

setup.describe.configure({ mode: 'serial' });
