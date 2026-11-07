# Media-on-S3 E2E: PNG upload -> object store -> content roundtrip -> delete.
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

[IO.File]::WriteAllBytes('C:\tmp\probe.png', [Convert]::FromBase64String(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=='))

$login = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" `
  -ContentType 'application/json' -Body '{"email":"dev@acmecorp.com","password":"DevSecurePassword2026!"}'
$tok = $login.access_token
Check 'dev-login' ($login.token_type -eq 'Bearer') "token_type=$($login.token_type)"

$appId = '11111111-1111-1111-1111-111111111111'
$orgId = '00000000-0000-0000-0000-000000000020'
$up = curl.exe -s -X POST "${base}/api/v1/developer/apps/${appId}/media" `
  -H "Authorization: Bearer $tok" `
  -F 'image=@C:\tmp\probe.png;type=image/png' `
  -F "org=$orgId" -F 'locale=ko-KR' -F 'formFactor=PHONE' -F 'assetType=SCREENSHOT' | ConvertFrom-Json
Check 'media-upload' ([bool]$up.id) "asset=$($up.id) url=$($up.url)"
$assetId = $up.id

$ls = docker @s3 s3 ls "s3://${MBucket}/media/11111111-1111-1111-1111-111111111111/" 2>&1 | Out-String
Check 's3-has-object' ($ls -match '\.png') "ls=$($ls.Trim())"

curl.exe -s "${base}/api/v1/developer/apps/${appId}/media/${assetId}/content?org=${orgId}" `
  -H "Authorization: Bearer $tok" -o C:\tmp\probe-back.png
$same = (Get-FileHash C:\tmp\probe.png -Algorithm SHA256).Hash -eq (Get-FileHash C:\tmp\probe-back.png -Algorithm SHA256).Hash
Check 'media-content-roundtrip' $same 'sha256 match'

$del = curl.exe -s -o NUL -w '%{http_code}' -X DELETE "${base}/api/v1/developer/apps/${appId}/media/${assetId}?org=${orgId}" `
  -H "Authorization: Bearer $tok"
$ls2 = docker @s3 s3 ls "s3://${MBucket}/media/11111111-1111-1111-1111-111111111111/" 2>&1 | Out-String
Check 'media-delete' (($del -eq '200' -or $del -eq '204') -and ($ls2 -notmatch '\.png')) "http=$del remaining=[$($ls2.Trim())]"

Write-Output "MEDIA_E2E_DONE failures=$fail"
exit $fail
