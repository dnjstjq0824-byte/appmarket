# Search-index E2E: admin MFA login -> APPROVED -> OpenSearch doc -> primary hit.
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

Invoke-Db "DELETE FROM mfa_credentials WHERE user_id='00000000-0000-0000-0000-000000000010';" | Out-Null
$enroll = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/mfa/enroll" `
  -ContentType 'application/json' -Body '{"email":"admin@bookingnetwork.com","password":"AdminSecurePassword2026!"}'
Check 'mfa-enroll' ([bool]$enroll.secret) "user=$($enroll.user_id)"
$code = Get-TOTP $enroll.secret
$login = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/mfa/verify" `
  -ContentType 'application/json' -Body "{`"user_id`":`"$($enroll.user_id)`",`"mfa_code`":`"$code`"}"
$H = @{ Authorization = "Bearer $($login.access_token)" }
Check 'admin-mfa-login' ($login.token_type -eq 'Bearer') "token_type=$($login.token_type)"

$cat = Invoke-Db "SELECT id FROM categories ORDER BY id LIMIT 1;"
$appId = '22222222-2222-2222-2222-222222222222'
Invoke-Db "INSERT INTO catalog_apps (id, canonical_slug, primary_package_name, owner_org_id, primary_category_id, privacy_policy_url, catalog_status) VALUES ('$appId','quokka-search-probe','com.quokka.searchprobe','00000000-0000-0000-0000-000000000020','$cat','https://example.local/privacy','PENDING_INITIAL_REVIEW') ON CONFLICT (id) DO UPDATE SET catalog_status='PENDING_INITIAL_REVIEW';" | Out-Null
Invoke-Db "INSERT INTO app_localizations (id, app_id, locale, title, short_description, full_description_sanitized, icon_s3_key) VALUES (gen_random_uuid(), '$appId', 'ko-KR', 'Quokka Search Probe', 'short desc', 'probe app verifying search indexing', 'icons/probe.png') ON CONFLICT (app_id, locale) DO NOTHING;" | Out-Null
Check 'draft-fixture' $true "app=$appId"

$dec = Invoke-RestMethod -Method Post -Uri "$base/api/v1/admin/reviews/$appId/decision" `
  -Headers $H -ContentType 'application/json' -Body '{"decision":"APPROVED","notes":"search e2e probe","checklist":{"icons":true}}'
Check 'review-approve' ($dec.decision -eq 'APPROVED') "decision=$($dec.decision)"

Start-Sleep -Seconds 3
$osBase = if ($env:OSBASE) { $env:OSBASE } else { 'http://127.0.0.1:9200' }
$idx = curl.exe -s "$osBase/apps/_count" | ConvertFrom-Json
Check 'opensearch-has-doc' ($idx.count -ge 1) "apps.count=$($idx.count)"
$s = Invoke-RestMethod -Uri "$base/api/v1/search?q=quokka" -Headers $H
$hit = ($s | ConvertTo-Json -Compress -Depth 6) -match 'quokka-search-probe|com.quokka.searchprobe'
Check 'search-primary-hit' (($s.fallback_used -eq $false) -and $hit) "fallback_used=$($s.fallback_used) hit=$hit total=$($s.total_count)"

Write-Output "SEARCH_E2E_DONE failures=$fail"
exit $fail
