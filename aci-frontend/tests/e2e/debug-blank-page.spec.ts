import { test } from '@playwright/test';

test.describe('Debug Blank Pages', () => {
  test('investigate threats page', async ({ page }) => {
    const consoleErrors: string[] = [];
    const networkErrors: string[] = [];

    // Capture console errors
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(`[Console Error] ${msg.text()}`);
      }
    });

    // Capture network failures
    page.on('requestfailed', request => {
      networkErrors.push(`[Network Error] ${request.url()} - ${request.failure()?.errorText}`);
    });

    // Capture all responses
    page.on('response', response => {
      if (response.status() >= 400) {
        console.log(`[HTTP ${response.status()}] ${response.url()}`);
      }
    });

    console.log('=== Navigating to /threats ===');
    await page.goto('http://localhost:5590/threats', { waitUntil: 'networkidle' });

    // Wait a moment for any JS to execute
    await page.waitForTimeout(3000);

    // Get page content
    const bodyContent = await page.locator('body').innerHTML();
    console.log('=== Body HTML ===');
    console.log(bodyContent.substring(0, 2000));

    // Check for root element
    const rootContent = await page.locator('#root').innerHTML().catch(() => 'NO #root ELEMENT');
    console.log('=== Root Content ===');
    console.log(rootContent.substring(0, 2000));

    // Take screenshot
    await page.screenshot({ path: 'threats-debug.png', fullPage: true });

    console.log('=== Console Errors ===');
    consoleErrors.forEach(e => console.log(e));

    console.log('=== Network Errors ===');
    networkErrors.forEach(e => console.log(e));

    // Check current URL (in case of redirect)
    console.log('=== Current URL ===');
    console.log(page.url());
  });

  test('investigate threat detail page', async ({ page }) => {
    const consoleErrors: string[] = [];
    const networkErrors: string[] = [];

    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(`[Console Error] ${msg.text()}`);
      }
    });

    page.on('requestfailed', request => {
      networkErrors.push(`[Network Error] ${request.url()} - ${request.failure()?.errorText}`);
    });

    console.log('=== Navigating to /threats/:id ===');
    await page.goto('http://localhost:5590/threats/e095cfc8-a5f5-45d3-b7fc-334a2bc79f24', { waitUntil: 'networkidle' });

    await page.waitForTimeout(3000);

    const bodyContent = await page.locator('body').innerHTML();
    console.log('=== Body HTML ===');
    console.log(bodyContent.substring(0, 2000));

    await page.screenshot({ path: 'threat-detail-debug.png', fullPage: true });

    console.log('=== Console Errors ===');
    consoleErrors.forEach(e => console.log(e));

    console.log('=== Network Errors ===');
    networkErrors.forEach(e => console.log(e));

    console.log('=== Current URL ===');
    console.log(page.url());
  });

  test('test login flow then threats', async ({ page }) => {
    const consoleErrors: string[] = [];

    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(`[Console Error] ${msg.text()}`);
      }
    });

    console.log('=== Testing Login First ===');
    await page.goto('http://localhost:5590/login', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    console.log('=== Login Page URL ===');
    console.log(page.url());

    // Check if login form exists
    const emailInput = page.locator('input[type="email"], input[name="email"]');
    const passwordInput = page.locator('input[type="password"]');

    const hasEmail = await emailInput.count() > 0;
    const hasPassword = await passwordInput.count() > 0;

    console.log(`Login form found: email=${hasEmail}, password=${hasPassword}`);

    if (hasEmail && hasPassword) {
      // Try to login with test credentials from DB
      await emailInput.fill('test@example.com');
      await passwordInput.fill('TestPass123!');

      // Find and click submit button
      const submitBtn = page.locator('button[type="submit"]');
      if (await submitBtn.count() > 0) {
        await submitBtn.click();
        await page.waitForTimeout(3000);

        console.log('=== After Login URL ===');
        console.log(page.url());

        // Now try threats page
        await page.goto('http://localhost:5590/threats', { waitUntil: 'networkidle' });
        await page.waitForTimeout(3000);

        const threatsContent = await page.locator('body').innerHTML();
        console.log('=== Threats Page After Login ===');
        console.log(threatsContent.substring(0, 2000));

        await page.screenshot({ path: 'threats-after-login.png', fullPage: true });
      }
    }

    console.log('=== Console Errors ===');
    consoleErrors.forEach(e => console.log(e));
  });
});
