# Release-workflow E2E (Play model): upload -> submit -> approve -> publish -> CDN -> halt/resume/full.
# Prerequisite: fixture APK (see binary-e2e.ps1 header).
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
$s3 = Get-S3Base
if (-not (Test-Path $ApkPath)) { throw "fixture APK missing: $ApkPath" }

$dev = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" `
  -ContentType 'application/json' -Body '{"email":"dev@acmecorp.com","password":"DevSecurePassword2026!"}'
$H = @{ Authorization = "Bearer $($dev.access_token)" }
$appId = '11111111-1111-1111-1111-111111111111'
$orgId = '00000000-0000-0000-0000-000000000020'

$sha = (Get-FileHash $ApkPath -Algorithm SHA256).Hash.ToLower()
$size = (Get-Item $ApkPath).Length
$init = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/uploads" `
  -Headers $H -ContentType 'application/json' `
  -Body "{`"orgId`":`"${orgId}`",`"fileName`":`"quokka-rel.apk`",`"fileSize`":${size},`"expectedSha256`":`"${sha}`"}"
$sid = if ($init.uploadId) { $init.uploadId } else { $init.id }
$tok = $dev.access_token
$done = curl.exe -s -X POST "${base}/api/v1/developer/apps/${appId}/uploads/${sid}/complete" `
  -H "Authorization: Bearer $tok" -F "apk=@${ApkPath};type=application/vnd.android.package-archive" -F "sha256=$sha" | ConvertFrom-Json
Check 'upload-complete' ($done.status -eq 'COMPLETED') "status=$($done.status) release=$($done.release.id)"

$cur = Invoke-RestMethod -Uri "${base}/api/v1/developer/apps/${appId}/releases/current?org=${orgId}" -Headers $H
Check 'release-scanning' ($cur.status -eq 'SECURITY_SCANNING') "status=$($cur.status)"

try {
  Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/publish?org=${orgId}" `
    -Headers $H -ContentType 'application/json' -Body '{"mode":"full"}' | Out-Null
  Check 'publish-premature' $false 'unexpected success'
} catch { Check 'publish-premature' ([int]$_.Exception.Response.StatusCode -eq 409) "http=$([int]$_.Exception.Response.StatusCode)" }

$sub = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/submit?org=${orgId}" `
  -Headers $H -ContentType 'application/json' -Body '{}'
Check 'submit-review' ($sub.status -eq 'SUBMITTED_FOR_REVIEW') "status=$($sub.status)"

Invoke-Db "UPDATE catalog_apps SET catalog_status='PENDING_INITIAL_REVIEW' WHERE id='$appId';" | Out-Null
Invoke-Db "DELETE FROM mfa_credentials WHERE user_id='00000000-0000-0000-0000-000000000010';" | Out-Null
$enroll = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/mfa/enroll" `
  -ContentType 'application/json' -Body '{"email":"admin@bookingnetwork.com","password":"AdminSecurePassword2026!"}'
$admin = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/mfa/verify" `
  -ContentType 'application/json' -Body "{`"user_id`":`"$($enroll.user_id)`",`"mfa_code`":`"$(Get-TOTP $enroll.secret)`"}"
$AH = @{ Authorization = "Bearer $($admin.access_token)" }
$dec = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/admin/reviews/${appId}/decision" `
  -Headers $AH -ContentType 'application/json' -Body '{"decision":"APPROVED","notes":"release e2e","checklist":{"icons":true}}'
Check 'admin-approve' ($dec.decision -eq 'APPROVED') "decision=$($dec.decision)"
$cur2 = Invoke-RestMethod -Uri "${base}/api/v1/developer/apps/${appId}/releases/current?org=${orgId}" -Headers $H
Check 'release-approved' ($cur2.status -eq 'APPROVED_PENDING_PUBLISH') "status=$($cur2.status)"

$pub = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/publish?org=${orgId}" `
  -Headers $H -ContentType 'application/json' -Body '{"mode":"staged","percentage":25}'
Check 'publish-staged' (($pub.release.status -eq 'PUBLISHED_STAGED') -and ($pub.rolloutPercentage -eq 25) -and ($pub.cdnUrls.Count -eq 1)) `
  "status=$($pub.release.status) pct=$($pub.rolloutPercentage) url=$($pub.cdnUrls[0])"
$cdnKey = $pub.cdnUrls[0] -replace '^https://[^/]+/', ''
$head = docker @s3 s3api head-object --bucket $PBucket --key $cdnKey 2>&1 | Out-String | ConvertFrom-Json
$localMd5 = (Get-FileHash $ApkPath -Algorithm MD5).Hash.ToLower()
Check 'cdn-bytes-identical' ((($head.ETag -replace '"', '').ToLower() -eq $localMd5) -and ($head.Metadata.sha256 -eq $sha)) "etag=$($head.ETag)"
Check 'cdn-cache' ($head.CacheControl -eq 'public, max-age=31536000, immutable') "cc=$($head.CacheControl)"

$halt = Invoke-RestMethod -Method Patch -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/rollout?org=${orgId}" `
  -Headers $H -ContentType 'application/json' -Body '{"action":"HALT","reason":"e2e drill"}'
Check 'halt' ($halt.status -eq 'HALTED') "status=$($halt.status)"
$res = Invoke-RestMethod -Method Patch -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/rollout?org=${orgId}" `
  -Headers $H -ContentType 'application/json' -Body '{"action":"RESUME"}'
Check 'resume' ($res.status -eq 'PUBLISHED_STAGED') "status=$($res.status)"
$full = Invoke-RestMethod -Method Patch -Uri "${base}/api/v1/developer/apps/${appId}/releases/current/rollout?org=${orgId}" `
  -Headers $H -ContentType 'application/json' -Body '{"rolloutPercentage":100}'
Check 'rollout-100' ($full.status -eq 'PUBLISHED_FULL') "status=$($full.status) pct=$($full.rolloutPercentage)"

Write-Output "RELEASE_E2E_DONE failures=$fail"
exit $fail
