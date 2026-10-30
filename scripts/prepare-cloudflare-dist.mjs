import { existsSync, cpSync, rmSync, mkdirSync, writeFileSync } from 'node:fs';
import { resolve, join } from 'node:path';

const rootDir = process.cwd();
const distDir = resolve(rootDir, 'dist');
const webUserPublic = resolve(rootDir, 'apps/web-user/.output/public');
const webAdminDist = resolve(rootDir, 'apps/web-admin/dist');
const webDevDist = resolve(rootDir, 'apps/web-dev/dist');

console.log('[cloudflare-dist] Preparing unified deployment bundle for Cloudflare Pages...');

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

// 3. Copy web-admin to dist/admin
if (existsSync(webAdminDist)) {
  const adminTarget = join(distDir, 'admin');
  mkdirSync(adminTarget, { recursive: true });
  console.log(`[cloudflare-dist] Copying web-admin (dist) -> dist/admin/`);
  cpSync(webAdminDist, adminTarget, { recursive: true });
} else {
  console.warn(`[cloudflare-dist] Warning: ${webAdminDist} not found`);
}

// 4. Copy web-dev to dist/developer
if (existsSync(webDevDist)) {
  const devTarget = join(distDir, 'developer');
  mkdirSync(devTarget, { recursive: true });
  console.log(`[cloudflare-dist] Copying web-dev (dist) -> dist/developer/`);
  cpSync(webDevDist, devTarget, { recursive: true });
} else {
  console.warn(`[cloudflare-dist] Warning: ${webDevDist} not found`);
}

// 5. Generate Cloudflare Pages configuration files
console.log('[cloudflare-dist] Generating Cloudflare Pages routing and headers...');

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

const redirectsContent = `# SPA routing fallbacks for Admin and Developer portals
/admin/*       /admin/index.html       200
/developer/*   /developer/index.html   200

# Nuxt 3 web-user fallback
/apps/*        /200.html               200
/search/*      /200.html               200
/account/*     /200.html               200
/support/*     /200.html               200
`;
writeFileSync(join(distDir, '_redirects'), redirectsContent, 'utf-8');

console.log('[cloudflare-dist] Unified Cloudflare Pages distribution bundle ready in ./dist');

