import { test } from '@playwright/test';

test.describe('Debug React Mount Issue', () => {
  test('capture all messages after login and threats navigation', async ({ page }) => {
    const allLogs: string[] = [];
    const networkRequests: string[] = [];
    const networkResponses: string[] = [];

    // Capture ALL console messages
    page.on('console', msg => {
      allLogs.push(`[${msg.type().toUpperCase()}] ${msg.text()}`);
    });

    // Capture page errors (uncaught exceptions)
    page.on('pageerror', err => {
      allLogs.push(`[PAGE ERROR] ${err.message}`);
    });

    // Capture network requests
    page.on('request', request => {
      networkRequests.push(`[REQUEST] ${request.method()} ${request.url()}`);
    });

    // Capture network responses
    page.on('response', response => {
      networkResponses.push(`[RESPONSE ${response.status()}] ${response.url()}`);
    });

    // Capture request failures
    page.on('requestfailed', request => {
      allLogs.push(`[NETWORK FAILED] ${request.url()} - ${request.failure()?.errorText}`);
    });

    // Step 1: Go to login
    console.log('=== STEP 1: Navigate to login ===');
    await page.goto('http://localhost:5590/login', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(2000);

    console.log('Logs after login page load:');
    allLogs.forEach(log => console.log('  ', log));
    allLogs.length = 0;

    // Step 2: Fill login form
    console.log('\n=== STEP 2: Login ===');
    const emailInput = page.locator('input[type="email"], input[name="email"]');
    const passwordInput = page.locator('input[type="password"]');

    await emailInput.fill('test@example.com');
    await passwordInput.fill('TestPass123!');

    const submitBtn = page.locator('button[type="submit"]');
    await submitBtn.click();
    await page.waitForTimeout(3000);

    console.log('After login - Current URL:', page.url());
    console.log('Logs after login:');
    allLogs.forEach(log => console.log('  ', log));
    allLogs.length = 0;

    // Step 3: Check localStorage for auth tokens
    const tokens = await page.evaluate(() => {
      return {
        accessToken: localStorage.getItem('access_token'),
        refreshToken: localStorage.getItem('refresh_token'),
        allKeys: Object.keys(localStorage)
      };
    });
    console.log('\n=== AUTH TOKENS IN LOCALSTORAGE ===');
    console.log('access_token exists:', !!tokens.accessToken);
    console.log('refresh_token exists:', !!tokens.refreshToken);
    console.log('All localStorage keys:', tokens.allKeys);

    // Step 4: Navigate to /threats using click (SPA navigation)
    console.log('\n=== STEP 3: Navigate to /threats via SPA ===');

    // First check if there's a nav link to threats
    const threatsLink = page.locator('a[href="/threats"]');
    const hasThreatsLink = await threatsLink.count() > 0;
    console.log('Found threats nav link:', hasThreatsLink);

    // Use direct click on nav link using JS (bypassing visibility)
    await page.evaluate(() => {
      const link = document.querySelector('a[href="/threats"]') as HTMLAnchorElement;
      if (link) {
        link.click();
      } else {
        // Fallback to history API
        window.history.pushState({}, '', '/threats');
        window.dispatchEvent(new PopStateEvent('popstate'));
      }
    });

    await page.waitForTimeout(5000);

    console.log('After SPA nav to /threats - Current URL:', page.url());
    console.log('Logs after threats navigation:');
    allLogs.forEach(log => console.log('  ', log));
    allLogs.length = 0;

    // Get HTML
    const html1 = await page.locator('#root').innerHTML().catch(() => 'FAILED TO GET ROOT');
    console.log('\n=== ROOT CONTENT AFTER SPA NAV ===');
    console.log(html1.substring(0, 1000));

    // Step 5: Try direct navigation to /threats
    console.log('\n=== STEP 4: Direct navigation to /threats ===');
    await page.goto('http://localhost:5590/threats', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(5000);

    console.log('After direct nav to /threats - Current URL:', page.url());
    console.log('Logs after direct navigation:');
    allLogs.forEach(log => console.log('  ', log));

    // Get HTML
    const html2 = await page.locator('#root').innerHTML().catch(() => 'FAILED TO GET ROOT');
    console.log('\n=== ROOT CONTENT AFTER DIRECT NAV ===');
    console.log(html2.substring(0, 1000));

    // Check if redirected to login
    if (page.url().includes('/login')) {
      console.log('\n=== REDIRECTED TO LOGIN - AUTH LOST ===');
      const newTokens = await page.evaluate(() => ({
        accessToken: localStorage.getItem('access_token'),
        allKeys: Object.keys(localStorage)
      }));
      console.log('Tokens after redirect:', newTokens);
    }

    // Take screenshot
    await page.screenshot({ path: 'mount-debug.png', fullPage: true });

    // Summary
    console.log('\n=== NETWORK SUMMARY ===');
    console.log('Requests:', networkRequests.length);
    const apiRequests = networkRequests.filter(r => r.includes('localhost:5580'));
    console.log('API requests to backend:');
    apiRequests.forEach(r => console.log('  ', r));
  });
});
