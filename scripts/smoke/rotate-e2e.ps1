# JWT rotation E2E: MFA admin -> rotate -> grace -> dual JWKS -> negatives.
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
$login = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/mfa/verify" `
  -ContentType 'application/json' -Body "{`"user_id`":`"$($enroll.user_id)`",`"mfa_code`":`"$(Get-TOTP $enroll.secret)`"}"
$tokA = $login.access_token
Check 'admin-mfa-login' ($login.token_type -eq 'Bearer') 'token A issued'

$jwks1 = Invoke-RestMethod -Uri "$base/api/v1/auth/jwks.json"
Check 'jwks-single-before' ($jwks1.keys.Count -eq 1) "keys=$($jwks1.keys.Count)"

$rot = Invoke-RestMethod -Method Post -Uri "$base/api/v1/admin/security/jwt/rotate" `
  -Headers @{Authorization = "Bearer $tokA"} -ContentType 'application/json' -Body '{}'
Check 'rotate' (($rot.kid -ne $rot.previous_kid) -and ($rot.previous_kid -eq $jwks1.keys[0].kid)) `
  "kid=$($rot.kid) prev=$($rot.previous_kid) by=$($rot.rotated_by)"

$me = Invoke-RestMethod -Uri "$base/api/v1/auth/users/me" -Headers @{Authorization = "Bearer $tokA"}
Check 'grace-token-works' ($me.email -eq 'admin@bookingnetwork.com') "email=$($me.email)"

$jwks2 = Invoke-RestMethod -Uri "$base/api/v1/auth/jwks.json"
$kids = @($jwks2.keys | ForEach-Object { $_.kid })
Check 'jwks-dual-after' (($jwks2.keys.Count -eq 2) -and ($kids -contains $rot.kid) -and ($kids -contains $rot.previous_kid)) "kids=$($kids -join ',')"

$dev = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" `
  -ContentType 'application/json' -Body '{"email":"dev@acmecorp.com","password":"DevSecurePassword2026!"}'
try {
  Invoke-RestMethod -Method Post -Uri "$base/api/v1/admin/security/jwt/rotate" `
    -Headers @{Authorization = "Bearer $($dev.access_token)"} -ContentType 'application/json' -Body '{}' | Out-Null
  Check 'user-rotate-forbidden' $false 'unexpected success'
} catch { Check 'user-rotate-forbidden' ([int]$_.Exception.Response.StatusCode -eq 403) "http=$([int]$_.Exception.Response.StatusCode)" }

Write-Output "ROTATE_E2E_DONE failures=$fail"
exit $fail
