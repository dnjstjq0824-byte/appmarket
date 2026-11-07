# Shared harness for scripts/smoke/*.ps1 (Windows PowerShell + docker).
# Dot-source AFTER the caller's param block: . "$PSScriptRoot/common.ps1"
# Callers define: $AppBase, $DbExec, $PgPassword, $DbUrl, $S3Endpoint,
# $S3Network, $QBucket, $PBucket, $MBucket. Each script keeps its own Check().
$base = $AppBase.TrimEnd('/')

function Get-TOTP($b32) {
  $b32 = $b32.ToUpper().Trim()
  $alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'
  $bits = ''
  foreach ($c in $b32.ToCharArray()) { $bits += [Convert]::ToString($alphabet.IndexOf($c), 2).PadLeft(5, '0') }
  $bytes = @()
  for ($i = 0; $i + 8 -le $bits.Length; $i += 8) { $bytes += [Convert]::ToByte($bits.Substring($i, 8), 2) }
  $counter = [BitConverter]::GetBytes([int64]([DateTimeOffset]::UtcNow.ToUnixTimeSeconds() / 30))
  if ([BitConverter]::IsLittleEndian) { [Array]::Reverse($counter) }
  $hmac = [System.Security.Cryptography.HMACSHA1]::new([byte[]]$bytes)
  $h = $hmac.ComputeHash($counter)
  $o = $h[$h.Length - 1] -band 0x0F
  $code = ((($h[$o] -band 0x7F) * 16777216) + ($h[$o + 1] * 65536) + ($h[$o + 2] * 256) + $h[$o + 3]) % 1000000
  return $code.ToString('D6')
}

$_DbPasswordCache = $null
function Invoke-Db([string]$Sql) {
  if ($DbExec -eq 'kubectl') {
    if (-not $DbUrl) { throw 'DBURL (or -DbUrl) is required for -DbExec kubectl' }
    $pod = "psql-smoke-$(Get-Random -Maximum 999999)"
    return kubectl -n bn-appmarket run $pod --rm -i --restart=Never --image=postgres:16-alpine `
      --env="DATABASE_URL=$DbUrl" -- psql "$DbUrl" -tA -c $Sql
  }
  if (-not $script:_DbPasswordCache) {
    if ($PgPassword) { $script:_DbPasswordCache = $PgPassword }
    else {
      $envLine = Get-Content 'deploy/.env' | Where-Object { $_ -match '^POSTGRES_PASSWORD=' } | Select-Object -First 1
      $script:_DbPasswordCache = ($envLine -split '=', 2)[1]
    }
  }
  return docker exec -e "PGPASSWORD=$script:_DbPasswordCache" bn-appmarket-postgres-1 psql -U bn_app -d bn_appmarket -tA -c $Sql
}

function Get-S3Base {
  $a = @('run', '--rm')
  if ($S3Network) { $a += @('--network', $S3Network) }
  $ak = if ($env:AWS_ACCESS_KEY_ID) { $env:AWS_ACCESS_KEY_ID } else { 'test' }
  $sk = if ($env:AWS_SECRET_ACCESS_KEY) { $env:AWS_SECRET_ACCESS_KEY } else { 'test' }
  $rg = if ($env:AWS_DEFAULT_REGION) { $env:AWS_DEFAULT_REGION } else { 'ap-northeast-2' }
  $a += @('-e', "AWS_ACCESS_KEY_ID=$ak", '-e', "AWS_SECRET_ACCESS_KEY=$sk",
    '-e', "AWS_DEFAULT_REGION=$rg", '-e', 'AWS_EC2_METADATA_DISABLED=true',
    'amazon/aws-cli:2.0.9', '--endpoint-url', $S3Endpoint)
  return $a
}
