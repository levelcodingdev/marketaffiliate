const COOKIE_DAYS = 30;

export function captureReferral(productId, referralCode) {
  if (!referralCode) return;

  const expires = new Date();
  expires.setDate(expires.getDate() + COOKIE_DAYS);
  document.cookie = `referral_code=${encodeURIComponent(referralCode)}; expires=${expires.toUTCString()}; path=/; SameSite=Lax`;
  document.cookie = `referral_product_id=${encodeURIComponent(productId)}; expires=${expires.toUTCString()}; path=/; SameSite=Lax`;
}

export function getReferralCode() {
  const match = document.cookie.match(/(?:^|; )referral_code=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}
