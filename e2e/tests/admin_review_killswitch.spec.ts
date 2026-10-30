import { test, expect } from '@playwright/test';

test.describe('Admin Console Security Review & Kill Switch E2E Journey', () => {
  test('Admin performs APK security review and tests emergency kill switch', async ({ page }) => {
    // 1. Visit admin console
    await page.goto('/');
    await expect(page.locator('text=관리자 콘솔')).toBeVisible();

    // 2. Open Security Review View
    await page.goto('/security/reviews');
    await expect(page.locator('text=APK 바이너리 보안 심사')).toBeVisible();
    await expect(page.locator('text=ClamAV')).toBeVisible();
    await expect(page.locator('text=YARA')).toBeVisible();

    // 3. Open Emergency Takedown (Kill Switch) View
    await page.goto('/security/kill-switch');
    await expect(page.locator('text=긴급 배포 차단')).toBeVisible();

    // 4. Test Dual-Control Payout Screen
    await page.goto('/finance/settlements');
    await expect(page.locator('text=월간 개발자 정산')).toBeVisible();
    await expect(page.locator('text=이중승인(Dual-Control)')).toBeVisible();
  });
});

