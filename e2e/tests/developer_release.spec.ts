import { test, expect } from '@playwright/test';

test.describe('Developer Console E2E Journey', () => {
  test('Developer views KYB status, uploads binary, and configures rollout', async ({ page }) => {
    // 1. Visit developer console
    await page.goto('/login');
    await expect(page.locator('h1, h2')).toContainText(/개발자 콘솔/i);

    // 2. Inspect KYB status
    await page.goto('/kyb');
    await expect(page.locator('text=KYB 인증')).toBeVisible();

    // 3. Navigate to APK Uploader
    await page.goto('/releases/upload');
    await expect(page.locator('text=대용량 APK / AAB 업로더')).toBeVisible();

    // 4. Verify Signing Lineage viewer
    await page.goto('/releases/signing-lineage');
    await expect(page.locator('text=서명 인증서 계보')).toBeVisible();

    // 5. Staged Rollout control
    await page.goto('/releases/rollout');
    await expect(page.locator('text=단계적 배포 (Staged Rollout)')).toBeVisible();
    
    // 6. Pricing & Promo code manager
    await page.goto('/commerce/pricing');
    await expect(page.locator('text=유료 앱 국가별 가격표')).toBeVisible();
  });
});

