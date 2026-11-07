# Deep-link SPA fallback regression test (edge assets layer, not the Go API).
# Every app shell must serve its OWN entry for deep links; serving the store
# home HTML (with the home payload) makes the client boot the wrong route.
# Requires: combined dist/ (node scripts/prepare-cloudflare-dist.mjs) served by
# `wrangler dev --port <port>` (assets + _redirects honored).
param(
  [string]$WorkerBase = $(if ($env:WORKER_BASE) { $env:WORKER_BASE } else { 'http://127.0.0.1:8789' })
)
$ErrorActionPreference = 'Stop'
$failures = 0
function Check($name, $cond, $detail) {
  if ($cond) { Write-Output "PASS $name :: $detail" }
  else { Write-Output "FAIL $name :: $detail"; $script:failures++ }
}
# Redirects are failures here: a 307 to a canonical shell URL would boot the
# client at the wrong path (deep link lost). Do not follow them.
function Fetch($path) {
  try {
    $r = Invoke-WebRequest -Uri "$($WorkerBase.TrimEnd('/'))$path" -UseBasicParsing -TimeoutSec 20 -MaximumRedirection 0
    return @{ status = [int]$r.StatusCode; content = [string]$r.Content }
  } catch {
    $s = 0
    try { $s = [int]$_.Exception.Response.StatusCode.value__ } catch { $s = -1 }
    return @{ status = $s; content = '' }
  }
}
function BodyOf($path) {
  return (Fetch $path).content
}
function StatusOf($path) {
  return (Fetch $path).status
}
$userShell = BodyOf '/apps/smoke-app'
Check 'apps-shell' (($userShell -match 'id="__nuxt"') -and ($userShell -notmatch 'ledger-home') -and ($userShell -notmatch '__NUXT_DATA__')) 'payload-free user shell'
Check 'apps-no-redirect' ((StatusOf '/apps/smoke-app') -eq 200) 'served directly, deep path kept'
$coShell = BodyOf '/checkout/smoke-app'
Check 'checkout-shell' (($coShell -match 'id="__nuxt"') -and ($coShell -notmatch 'ledger-home') -and ($coShell -notmatch '__NUXT_DATA__')) 'payload-free user shell'
Check 'checkout-no-redirect' ((StatusOf '/checkout/smoke-app') -eq 200) 'served directly, deep path kept'
$devShell = BodyOf '/developer/login'
Check 'dev-shell' ($devShell -match '/developer/assets/') 'dev console entry served'
Check 'dev-no-redirect' ((StatusOf '/developer/login') -eq 200) 'served directly, deep path kept'
$adminShell = BodyOf '/admin/login'
Check 'admin-shell' ($adminShell -match '/admin/assets/') 'admin console entry served'
Check 'admin-no-redirect' ((StatusOf '/admin/login') -eq 200) 'served directly, deep path kept'
$homeBody = BodyOf '/'
Check 'home-intact' ($homeBody -match 'ledger-home') 'store home still prerendered'
Write-Output "ROUTES_E2E_DONE failures=$failures"
exit $failures
