import { test, expect } from '@playwright/test';

test.describe('User Market Website E2E Journey', () => {
  test('User searches for app, compares stores, and verifies authentic CTA', async ({ page }) => {
    // 1. Visit homepage
    await page.goto('/');
    await expect(page).toHaveTitle(/BookingNetwork App Market/i);

    // 2. Perform search
    const searchInput = page.locator('input[placeholder*="검색"]');
    await searchInput.fill('vector');
    await searchInput.press('Enter');

    // 3. Verify search result and click
    const appCard = page.locator('text=Vector Pro Studio').first();
    await expect(appCard).toBeVisible();
    await appCard.click();

    // 4. Verify detail page & multi-store comparison matrix
    await expect(page.locator('text=배포처 및 가격 정직 비교')).toBeVisible();
    await expect(page.locator('text=자체 호스팅 (BN_HOSTED)')).toBeVisible();

    // 5. Verify 0 fake download button guarantee
    const fakeAds = page.locator('.fake-download-btn, iframe[src*="doubleclick"]');
    await expect(fakeAds).toHaveCount(0);

    // 6. Navigate to checkout for paid app
    const buyButton = page.locator('button:has-text("안전 결제"), a:has-text("구매하기")').first();
    if (await buyButton.isVisible()) {
      await buyButton.click();
      await expect(page.locator('text=결제 금액 산출서')).toBeVisible();
    }
  });

  test('User applies promo code and verifies real-time discount', async ({ page }) => {
    await page.goto('/checkout/com.creativedesign.vectorpro');

    const promoInput = page.locator('input[placeholder*="쿠폰 코드"]');
    await promoInput.fill('FREE2026');
    await page.locator('button:has-text("적용")').click();

    // Verify 100% discount applied
    await expect(page.locator('text=100% 무료 리딤 적용')).toBeVisible();
    await expect(page.locator('text=최종 결제 금액 (Payable)')).toBeVisible();
  });
});

