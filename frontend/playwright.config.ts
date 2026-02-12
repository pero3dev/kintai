import { defineConfig, devices } from '@playwright/test';

const compatibilitySpec = /compatibility\.spec\.ts/;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      testMatch: compatibilitySpec,
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      testMatch: compatibilitySpec,
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'edge',
      testMatch: compatibilitySpec,
      use: { ...devices['Desktop Edge'] },
    },
    {
      name: 'mobile-android',
      testMatch: compatibilitySpec,
      use: { ...devices['Pixel 7'] },
    },
    {
      name: 'mobile-ios',
      testMatch: compatibilitySpec,
      use: { ...devices['iPhone 14'] },
    },
    {
      name: 'tablet-ipad',
      testMatch: compatibilitySpec,
      use: { ...devices['iPad Pro 11'] },
    },
  ],
  webServer: {
    command: 'pnpm dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
