/**
 * Cloudflare Worker / Fastly Compute / CloudFront Function Edge Script
 * AM-OPS-003: CDN Edge HMAC-SHA256 Signed Download Token Verifier & Immutable Caching
 * TC-OPS-003 Compliance
 */

const CDN_HMAC_SECRET = "cdn_master_shared_secret_2026_super_secure_key";

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // Only inspect download endpoints
    if (!url.pathname.startsWith("/artifacts/download/") && !url.pathname.startsWith("/api/v1/distribution/download/")) {
      return fetch(request);
    }

    const pkg = url.searchParams.get("pkg");
    const expStr = url.searchParams.get("exp");
    const sig = url.searchParams.get("sig");
    const uid = url.searchParams.get("uid") || "";

    // 1. Missing required parameters
    if (!pkg || !expStr || !sig) {
      return new Response(JSON.stringify({
        error: "ERR_DOWNLOAD_TOKEN_MISSING",
        message: "Download request is missing required cryptographic signature or expiration timestamp."
      }), {
        status: 403,
        headers: { "Content-Type": "application/json", "Cache-Control": "no-store" }
      });
    }

    // 2. Expiration check
    const exp = parseInt(expStr, 10);
    const nowSec = Math.floor(Date.now() / 1000);
    if (isNaN(exp) || nowSec > exp) {
      return new Response(JSON.stringify({
        error: "ERR_DOWNLOAD_TOKEN_EXPIRED",
        message: "Signed download token has expired (15-minute TTL exceeded)."
      }), {
        status: 403,
        headers: { "Content-Type": "application/json", "Cache-Control": "no-store" }
      });
    }

    // 3. Cryptographic HMAC-SHA256 signature verification at edge
    const canonicalMessage = `${url.pathname}|${pkg}|${exp}|${uid}`;
    const encoder = new TextEncoder();
    const keyData = encoder.encode(env?.CDN_SECRET || CDN_HMAC_SECRET);
    
    try {
      const cryptoKey = await crypto.subtle.importKey(
        "raw",
        keyData,
        { name: "HMAC", hash: "SHA-256" },
        false,
        ["sign", "verify"]
      );

      // Convert hex signature to ArrayBuffer
      const sigBytes = new Uint8Array(sig.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
      const isValid = await crypto.subtle.verify(
        "HMAC",
        cryptoKey,
        sigBytes,
        encoder.encode(canonicalMessage)
      );

      if (!isValid) {
        return new Response(JSON.stringify({
          error: "ERR_DOWNLOAD_SIGNATURE_INVALID",
          message: "Cryptographic signature mismatch. Token was tampered or generated with unauthorized key."
        }), {
          status: 403,
          headers: { "Content-Type": "application/json", "Cache-Control": "no-store" }
        });
      }
    } catch (e) {
      return new Response(JSON.stringify({
        error: "ERR_CRYPTO_VERIFY_FAILURE",
        message: "Edge crypto verification engine encountered internal error: " + e.message
      }), {
        status: 403,
        headers: { "Content-Type": "application/json", "Cache-Control": "no-store" }
      });
    }

    // 4. Token is valid. Fetch from primary storage origin with immutable cache injection
    const response = await fetch(request);
    const newHeaders = new Headers(response.headers);

    // Enforce immutable 1-year browser/CDN caching for binary APK blobs
    newHeaders.set("Cache-Control", "public, max-age=31536000, immutable");
    newHeaders.set("X-Edge-Verification", "PASSED_HMAC_SHA256");
    newHeaders.set("X-Content-Type-Options", "nosniff");

    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: newHeaders
    });
  }
};

