import { existsSync, cpSync, rmSync, mkdirSync, writeFileSync, copyFileSync, readFileSync } from 'node:fs';
import { resolve, join } from 'node:path';

const rootDir = process.cwd();
const distDir = resolve(rootDir, 'dist');
const webUserPublic = resolve(rootDir, 'apps/web-user/.output/public');
const webAdminDist = resolve(rootDir, 'apps/web-admin/dist');
const webDevDist = resolve(rootDir, 'apps/web-dev/dist');

console.log('[cloudflare-dist] Preparing unified deployment bundle for Cloudflare Pages / Workers...');

// 1. Clean and recreate dist
if (existsSync(distDir)) {
  rmSync(distDir, { recursive: true, force: true });
}
mkdirSync(distDir, { recursive: true });

// 2. Copy web-user public static assets if present
if (existsSync(webUserPublic)) {
  console.log(`[cloudflare-dist] Copying web-user (.output/public) -> dist/`);
  cpSync(webUserPublic, distDir, { recursive: true });
} else {
  console.warn(`[cloudflare-dist] Warning: ${webUserPublic} not found, generating minimal portal index`);
  const minimalHtml = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Booking Network App Market</title>
</head>
<body>
  <h1>Booking Network App Market</h1>
  <ul>
    <li><a href="/admin/">Admin Console</a></li>
    <li><a href="/developer/">Developer Console</a></li>
  </ul>
</body>
</html>`;
  writeFileSync(join(distDir, 'index.html'), minimalHtml, 'utf-8');
}

// Ensure root 200.html exists for Cloudflare native SPA fallback
if (existsSync(join(distDir, 'index.html')) && !existsSync(join(distDir, '200.html'))) {
  copyFileSync(join(distDir, 'index.html'), join(distDir, '200.html'));
}

// 3. Copy web-admin to dist/admin
if (existsSync(webAdminDist)) {
  const adminTarget = join(distDir, 'admin');
  mkdirSync(adminTarget, { recursive: true });
  console.log(`[cloudflare-dist] Copying web-admin (dist) -> dist/admin/`);
  cpSync(webAdminDist, adminTarget, { recursive: true });
  // Add 200.html for admin SPA fallback
  if (existsSync(join(adminTarget, 'index.html'))) {
    copyFileSync(join(adminTarget, 'index.html'), join(adminTarget, '200.html'));
  }
} else {
  console.warn(`[cloudflare-dist] Warning: ${webAdminDist} not found`);
}

// 4. Copy web-dev to dist/developer
if (existsSync(webDevDist)) {
  const devTarget = join(distDir, 'developer');
  mkdirSync(devTarget, { recursive: true });
  console.log(`[cloudflare-dist] Copying web-dev (dist) -> dist/developer/`);
  cpSync(webDevDist, devTarget, { recursive: true });
  // Add 200.html for developer SPA fallback
  if (existsSync(join(devTarget, 'index.html'))) {
    copyFileSync(join(devTarget, 'index.html'), join(devTarget, '200.html'));
  }
} else {
  console.warn(`[cloudflare-dist] Warning: ${webDevDist} not found`);
}

// 5. Generate Cloudflare Pages / Workers configuration files
console.log('[cloudflare-dist] Generating Cloudflare edge routing and headers...');

const routesJson = {
  version: 1,
  include: ['/*'],
  exclude: [
    '/_nuxt/*',
    '/assets/*',
    '/admin/assets/*',
    '/developer/assets/*',
    '/*.ico',
    '/*.png',
    '/*.svg',
    '/*.json',
    '/*.txt',
    '/*.xml'
  ]
};
writeFileSync(join(distDir, '_routes.json'), JSON.stringify(routesJson, null, 2), 'utf-8');

const headersContent = `/*
  X-Content-Type-Options: nosniff
  X-Frame-Options: SAMEORIGIN
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: camera=(), microphone=(), geolocation=()

/_nuxt/*
  Cache-Control: public, max-age=31536000, immutable

/assets/*
  Cache-Control: public, max-age=31536000, immutable

/admin/assets/*
  Cache-Control: public, max-age=31536000, immutable

/developer/assets/*
  Cache-Control: public, max-age=31536000, immutable
`;
writeFileSync(join(distDir, '_headers'), headersContent, 'utf-8');

// Note: Do NOT use _redirects for SPA fallbacks here. Catch-all rules are
// rejected by the edge validator (100324), and scoped splats are not applied
// under existing directories (/developer/*, /admin/* still fall through).
// Fallbacks live in worker/index.ts instead: exact asset first, then the
// owning app's shell on miss (200.html for the store, per-console index).
// Requires not_found_handling = "none". Verified by
// scripts/smoke/routes-e2e.ps1 against `wrangler dev`.
//
// The generated 200.html embeds the prerender payload of the internal /200
// route, which pins client boot to a nonexistent route ("Page not found:
// /200"). The store fetches all data client-side, so strip the payload tag;
// the shell then boots fresh and routes by URL. Config (window.__NUXT__)
// and head payload stay untouched.
const shellPath = join(distDir, '200.html');
const shellHtml = readFileSync(shellPath, 'utf-8');
const strippedShell = shellHtml.replace(
  /<script[^>]*id="__NUXT_DATA__"[^>]*>[\s\S]*?<\/script>/,
  ''
);
if (strippedShell === shellHtml) {
  throw new Error('[cloudflare-dist] __NUXT_DATA__ tag not found in 200.html');
}
writeFileSync(shellPath, strippedShell, 'utf-8');

console.log('[cloudflare-dist] Unified Cloudflare distribution bundle ready in ./dist');
