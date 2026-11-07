# Core live-stack smoke: auth, upload sessions, search fallback, webhook guard,
# sandbox order+checkout, catalog. Run from repo root.
#   powershell -File scripts/smoke/core-e2e.ps1
# Remote target: $env:APPBASE='https://api.example.com'; DB via -DbExec kubectl -DbUrl ... (see docs/LIVE_VERIFICATION_RUNBOOK.md)
param(
  [string]$AppBase = $(if ($env:APPBASE) { $env:APPBASE } else { 'http://127.0.0.1:18080' }),
  [ValidateSet('docker', 'kubectl')][string]$DbExec = $(if ($env:DBEXEC) { $env:DBEXEC } else { 'docker' }),
  [string]$PgPassword = $env:PGPASSWORD,
  [string]$DbUrl = $env:DBURL,
  [string]$S3Endpoint = $(if ($env:S3ENDPOINT) { $env:S3ENDPOINT } else { 'http://localstack:4566' }),
  [string]$S3Network = $(if ($env:S3NETWORK) { $env:S3NETWORK } else { 'bn-appmarket_data' }),
  [string]$QBucket = $(if ($env:QBUCKET) { $env:QBUCKET } else { 'bn-quarantine-bucket' }),
  [string]$PBucket = $(if ($env:PBUCKET) { $env:PBUCKET } else { 'bn-public-cdn-bucket' }),
  [string]$MBucket = $(if ($env:MBUCKET) { $env:MBUCKET } else { 'bn-public-media' }),
  [string]$ApkPath = $(if ($env:APKPATH) { $env:APKPATH } else { 'C:\tmp\fixture.apk' })
)
$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/common.ps1"
$fail = 0
function Check($name, $cond, $detail) {
  if ($cond) { Write-Output "PASS $name :: $detail" }
  else { Write-Output "FAIL $name :: $detail"; $script:fail++ }
}
function FailCode($name, $status, $detail) {
  Check $name ($status -ge 400 -and $status -lt 500) "http=$status $detail"
}

try {
  $reg = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/register" -ContentType 'application/json' -Body '{"email":"smoke@test.local","password":"SmokeTest123!","display_name":"Smoke"}'
  Check 'register' ([bool]$reg.id) "id=$($reg.id)"
} catch {
  Check 'register' ([int]$_.Exception.Response.StatusCode -eq 409) "http=$([int]$_.Exception.Response.StatusCode) (already exists, rerunnable)"
}
$login = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" -ContentType 'application/json' -Body '{"email":"smoke@test.local","password":"SmokeTest123!"}'
Check 'login' ($login.token_type -eq 'Bearer') "token_type=$($login.token_type) has_refresh=$([bool]$login.refresh_token)"
$me = Invoke-RestMethod -Uri "$base/api/v1/auth/users/me" -Headers @{Authorization = "Bearer $($login.access_token)"}
Check 'me' ($me.email -eq 'smoke@test.local') "email=$($me.email)"
$jwks = Invoke-RestMethod -Uri "$base/api/v1/auth/jwks.json"
Check 'jwks' ($jwks.keys.Count -ge 1) "keys=$($jwks.keys.Count)"

$H = @{ Authorization = "Bearer $($login.access_token)" }
$devLogin = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" -ContentType 'application/json' -Body '{"email":"dev@acmecorp.com","password":"DevSecurePassword2026!"}'
$DH = @{ Authorization = "Bearer $($devLogin.access_token)" }
$cat = Invoke-Db "SELECT id FROM categories ORDER BY id LIMIT 1;"
$appId = '11111111-1111-1111-1111-111111111111'
Invoke-Db "INSERT INTO catalog_apps (id, canonical_slug, primary_package_name, owner_org_id, primary_category_id, privacy_policy_url) VALUES ('$appId','smoke-app','com.smoke.e2e','00000000-0000-0000-0000-000000000020','$cat','https://example.local/privacy') ON CONFLICT (id) DO NOTHING;" | Out-Null

$sha = ('ab' * 32)
$init = Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads" `
  -Headers $DH -ContentType 'application/json' `
  -Body "{`"orgId`":`"00000000-0000-0000-0000-000000000020`",`"fileName`":`"smoke.apk`",`"fileSize`":1024,`"expectedSha256`":`"$sha`"}"
$sid = if ($init.uploadId) { $init.uploadId } else { $init.id }
Check 'upload-init' ([bool]$sid) "session=$sid status=$($init.status)"
$done = Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads/$sid/complete" `
  -Headers $DH -ContentType 'application/json' -Body "{`"sha256`":`"$sha`"}"
Check 'upload-complete' ($done.status -eq 'COMPLETED') "status=$($done.status)"
try {
  Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads/$sid/complete" `
    -Headers $DH -ContentType 'application/json' -Body "{`"sha256`":`"$sha`"}" | Out-Null
  FailCode 'upload-double-complete' 200 'unexpected success'
} catch { FailCode 'upload-double-complete' ([int]$_.Exception.Response.StatusCode) $_.Exception.Message }
try {
  Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads" `
    -Headers $DH -ContentType 'application/json' `
    -Body "{`"orgId`":`"00000000-0000-0000-0000-000000000020`",`"fileName`":`"evil.exe`",`"fileSize`":1024,`"expectedSha256`":`"$sha`"}" | Out-Null
  FailCode 'upload-exe-rejected' 200 'unexpected success'
} catch { FailCode 'upload-exe-rejected' ([int]$_.Exception.Response.StatusCode) $_.Exception.Message }
$init2 = Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads" `
  -Headers $DH -ContentType 'application/json' `
  -Body "{`"orgId`":`"00000000-0000-0000-0000-000000000020`",`"fileName`":`"smoke2.apk`",`"fileSize`":1024,`"expectedSha256`":`"$sha`"}"
$sid2 = if ($init2.uploadId) { $init2.uploadId } else { $init2.id }
try {
  Invoke-RestMethod -Method Post -Uri "$base/api/v1/developer/apps/$appId/uploads/$sid2/complete" `
    -Headers $DH -ContentType 'application/json' -Body '{"sha256":"cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd"}' | Out-Null
  FailCode 'upload-sha-mismatch' 200 'unexpected success'
} catch { FailCode 'upload-sha-mismatch' ([int]$_.Exception.Response.StatusCode) $_.Exception.Message }

$s = Invoke-RestMethod -Uri "$base/api/v1/search?q=smoke" -Headers $DH
Check 'search' $true "fallback_used=$($s.fallback_used) total=$($s.total_count)"
$sg = Invoke-RestMethod -Uri "$base/api/v1/search/suggest?q=sm" -Headers $DH
Check 'search-suggest' $true 'ok'

try {
  Invoke-RestMethod -Method Post -Uri "$base/api/v1/commerce/webhooks/pg" `
    -ContentType 'application/json' -Headers @{ 'X-PG-Signature' = 'bogus' } `
    -Body '{"orderId":"ord-smoke-1","amount":9900,"status":"DONE"}' | Out-Null
  FailCode 'webhook-bogus-sig' 200 'unexpected success'
} catch { FailCode 'webhook-bogus-sig' ([int]$_.Exception.Response.StatusCode) $_.Exception.Message }

$order = Invoke-RestMethod -Method Post -Uri "$base/api/v1/commerce/orders" `
  -Headers $DH -ContentType 'application/json' `
  -Body "{`"app_id`":`"$appId`",`"org_id`":`"00000000-0000-0000-0000-000000000020`",`"gross_amount`":9900,`"currency`":`"KRW`",`"idempotency_key`":`"smoke-e2e-$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())`"}"
$oid = $order.id; if (-not $oid) { $oid = $order.orderId }
Check 'order-create' ([bool]$oid) "order=$oid"
if ($oid) {
  $co = Invoke-RestMethod -Method Post -Uri "$base/api/v1/commerce/orders/$oid/checkout" `
    -Headers $DH -ContentType 'application/json' -Body '{}'
  Check 'checkout-session' ([bool]$co.checkout_url) "url=$($co.checkout_url)"
}

Invoke-Db "UPDATE catalog_apps SET catalog_status='PUBLISHED', published_at=NOW() WHERE id='$appId';" | Out-Null
$apps = Invoke-RestMethod -Uri "$base/api/v1/catalog/apps?limit=50"
Check 'catalog-has-fixture' (($apps | ConvertTo-Json -Compress) -match 'smoke-app') 'slug=smoke-app'

Write-Output "E2E_DONE failures=$fail"
exit $fail
