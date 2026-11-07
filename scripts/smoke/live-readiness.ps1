# Live-verification readiness probe for the 4 blocked items (docs/LIVE_VERIFICATION_RUNBOOK.md).
# Safe to run without any secret: reports READY/BLOCKED per item, never prints secret values.
# Exit code is always 0 (this is a report, not a gate).
param(
  [string]$WorkerUrl = $(if ($env:WORKER_URL) { $env:WORKER_URL } else { 'https://appmarket.dnjstjqppk.workers.dev' })
)
$ErrorActionPreference = 'Stop'
$ready = 0
function Report($name, $isReady, $detail) {
  if ($isReady) { Write-Output "READY   $name :: $detail"; $script:ready++ }
  else { Write-Output "BLOCKED $name :: $detail" }
}
function HasEnv($name) { return -not [string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($name)) }
function HasCmd($name) { return $null -ne (Get-Command $name -ErrorAction SilentlyContinue) }

# Block 1: live PG keys (presence only).
$pgLive = $env:PG_IS_LIVE -eq 'true'
$toss = HasEnv 'TOSS_SECRET_KEY'
$portone = (HasEnv 'PORTONE_API_KEY') -and (HasEnv 'PORTONE_STORE_ID')
Report 'pg-live' (($pgLive -and ($toss -or $portone))) "PG_IS_LIVE=$pgLive toss=$toss portone=$portone"

# Block 2: real S3/GCS wiring (env shape only; bytes are verified by media/binary smokes).
$objS3 = ($env:OBJECT_STORAGE -eq 's3')
$realEndpoint = [string]::IsNullOrWhiteSpace($env:S3_ENDPOINT) -or ($env:S3_ENDPOINT -match 'storage.googleapis.com|amazonaws.com')
$hasCreds = (HasEnv 'S3_ACCESS_KEY_ID') -or (HasEnv 'AWS_ACCESS_KEY_ID') -or (HasEnv 'AWS_PROFILE') -or (HasEnv 'GOOGLE_APPLICATION_CREDENTIALS')
Report 's3-live' (($objS3 -and $realEndpoint -and $hasCreds)) "OBJECT_STORAGE=s3:$objS3 real-endpoint:$realEndpoint creds-or-role:$hasCreds"

# Block 3: worker edge proxy state via public /healthz/ready.
try {
  $h = Invoke-RestMethod -Uri "$($WorkerUrl.TrimEnd('/'))/healthz/ready" -TimeoutSec 15
  Report 'worker-proxy' ([bool]$h.backend_proxy_configured) "backend_proxy_configured=$($h.backend_proxy_configured) arch=$($h.architecture)"
} catch { Report 'worker-proxy' $false "worker unreachable: $($_.Exception.Message)" }

# Block 4: cloud apply tooling + credentials + live backend files.
$tf = HasCmd 'terraform'; $kc = HasCmd 'kubectl'
$awsCred = (HasEnv 'AWS_ACCESS_KEY_ID') -or (HasEnv 'AWS_PROFILE')
$gcpCred = HasEnv 'GOOGLE_APPLICATION_CREDENTIALS'
$gcloudCli = HasCmd 'gcloud'
$kubeCred = (HasEnv 'KUBECONFIG') -or (Test-Path "$env:USERPROFILE/.kube/config")
$liveHcl = (Test-Path 'deploy/terraform/aws/backend.s3.live.hcl') -or (Test-Path 'deploy/terraform/gcp/backend.gcs.live.hcl')
Report 'cloud-apply' ((($awsCred -or $gcpCred -or $kubeCred) -and ($tf -or $kc))) "terraform=$tf kubectl=$kc gcloud=$gcloudCli aws-cred=$awsCred gcp-cred=$gcpCred kubeconfig=$kubeCred live-hcl=$liveHcl"

Write-Output "LIVE_READINESS_DONE ready=$ready/4 (see docs/LIVE_VERIFICATION_RUNBOOK.md)"
exit 0
