# Binary-upload E2E: synthetic signed APK -> quarantine S3 -> scan -> release.
# Prerequisite: fixture APK. Generate with:
#   go run scripts/smoke/genapk/main.go C:\tmp\fixture.apk   (or -ApkPath <path>)
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
if (-not (Test-Path $ApkPath)) { throw "fixture APK missing: $ApkPath (generate: go run scripts/smoke/genapk/main.go $ApkPath)" }

$login = Invoke-RestMethod -Method Post -Uri "$base/api/v1/auth/login" `
  -ContentType 'application/json' -Body '{"email":"dev@acmecorp.com","password":"DevSecurePassword2026!"}'
$tok = $login.access_token
$appId = '11111111-1111-1111-1111-111111111111'
$orgId = '00000000-0000-0000-0000-000000000020'
$sha = (Get-FileHash $ApkPath -Algorithm SHA256).Hash.ToLower()
$size = (Get-Item $ApkPath).Length

$init = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/uploads" `
  -Headers @{Authorization = "Bearer $tok"} -ContentType 'application/json' `
  -Body "{`"orgId`":`"${orgId}`",`"fileName`":`"quokka-1.0.7.apk`",`"fileSize`":${size},`"expectedSha256`":`"${sha}`"}"
$sid = if ($init.uploadId) { $init.uploadId } else { $init.id }
Check 'binary-init' ([bool]$sid) "session=$sid"

$doneRaw = curl.exe -s -X POST "${base}/api/v1/developer/apps/${appId}/uploads/${sid}/complete" `
  -H "Authorization: Bearer $tok" -F "apk=@${ApkPath};type=application/vnd.android.package-archive" -F "sha256=$sha"
$done = $doneRaw | ConvertFrom-Json
Check 'binary-complete' ($done.status -eq 'COMPLETED') "status=$($done.status)"
Check 'binary-details' ($done.details.packageName -eq 'com.quokka.livee2e' -and $done.details.versionCode -eq 7) `
  "pkg=$($done.details.packageName) vcode=$($done.details.versionCode) sdk=$($done.details.minSdk)"
Check 'binary-release' ([bool]$done.release.id) "release=$($done.release.id)"

$qkey = "quarantine/${orgId}/${appId}/${sid}"
$head = docker @s3 s3api head-object --bucket $QBucket --key $qkey 2>&1 | Out-String
Check 'quarantine-object' ($head -match 'ContentLength') "bucket=$QBucket"
$localMd5 = (Get-FileHash $ApkPath -Algorithm MD5).Hash.ToLower()
$headJson = $head | ConvertFrom-Json
$etag = ($headJson.ETag -replace '"', '').ToLower()
Check 'quarantine-bytes-identical' (($etag -eq $localMd5) -and ($headJson.Metadata.sha256 -eq $sha)) "etag=$etag meta_sha_match=$($headJson.Metadata.sha256 -eq $sha)"
$tagging = docker @s3 s3api get-object-tagging --bucket $QBucket --key $qkey 2>&1 | Out-String
Check 'quarantine-tags' (($tagging -match 'quarantined') -and ($tagging -match $sid)) 'status+session tags present'

$init2 = Invoke-RestMethod -Method Post -Uri "${base}/api/v1/developer/apps/${appId}/uploads" `
  -Headers @{Authorization = "Bearer $tok"} -ContentType 'application/json' `
  -Body "{`"orgId`":`"${orgId}`",`"fileName`":`"evil.apk`",`"fileSize`":1024,`"expectedSha256`":`"${sha}`"}"
$sid2 = if ($init2.uploadId) { $init2.uploadId } else { $init2.id }
$tampered = Join-Path ([IO.Path]::GetTempPath()) 'tampered.apk'
Copy-Item $ApkPath $tampered -Force
$fs = [IO.File]::OpenWrite($tampered); $fs.Seek(0, [IO.SeekOrigin]::End) | Out-Null; $fs.Write([byte[]](9, 8, 7, 6, 5), 0, 5); $fs.Close()
$code = curl.exe -s -o NUL -w '%{http_code}' -X POST "${base}/api/v1/developer/apps/${appId}/uploads/${sid2}/complete" `
  -H "Authorization: Bearer $tok" -F "apk=@${tampered};type=application/vnd.android.package-archive" -F "sha256=$sha"
Check 'binary-tamper-rejected' ($code -eq '409') "http=$code"

Write-Output "BINARY_E2E_DONE failures=$fail"
exit $fail
