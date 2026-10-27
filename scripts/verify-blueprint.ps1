param(
    [string]$ArchivePath,
    [string]$BlueprintPath = (Join-Path $PSScriptRoot '..\BN_APP_MARKET_MASTER_BLUEPRINT'),
    [string]$LogPath = (Join-Path $PSScriptRoot '..\artifacts\evidence\intake\checksum-verification-current.log')
)

$ErrorActionPreference = 'Stop'
$expectedArchiveHash = 'fd8111b8d631c533c276857a21fcca1a48972f890dbbb5b97fa21759206abe08'
$lines = [System.Collections.Generic.List[string]]::new()

if ($ArchivePath) {
    $actualArchiveHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $ArchivePath).Hash.ToLowerInvariant()
    if ($actualArchiveHash -ne $expectedArchiveHash) {
        throw "Archive SHA-256 mismatch: $actualArchiveHash"
    }
    $lines.Add("Archive SHA-256: OK ($actualArchiveHash)")
} else {
    $lines.Add('Archive SHA-256: NOT RECHECKED (archive path not supplied)')
}

$blueprint = (Resolve-Path -LiteralPath $BlueprintPath).Path
$manifest = Join-Path $blueprint 'FILE_CHECKSUMS.sha256'
$entries = Get-Content -LiteralPath $manifest
$checked = 0
$failed = 0
foreach ($entry in $entries) {
    if ($entry -notmatch '^([0-9a-fA-F]{64})\s+\*?(.+)$') {
        throw "Malformed checksum manifest entry: $entry"
    }
    $expected = $Matches[1].ToLowerInvariant()
    $relative = $Matches[2]
    $target = Join-Path $blueprint $relative
    if (-not (Test-Path -LiteralPath $target -PathType Leaf)) {
        $lines.Add($relative + ': MISSING')
        $failed++
        continue
    }
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $target).Hash.ToLowerInvariant()
    if ($actual -eq $expected) {
        $lines.Add($relative + ': OK')
    } else {
        $lines.Add($relative + ': MISMATCH')
        $failed++
    }
    $checked++
}
$lines.Add("Summary: $checked checked, $failed failures, $($entries.Count) manifest entries")
$directory = Split-Path -Parent $LogPath
New-Item -ItemType Directory -Force -Path $directory | Out-Null
$lines | Set-Content -LiteralPath $LogPath -Encoding UTF8
$lines | Write-Output
if ($failed -ne 0 -or $entries.Count -ne 140) {
    exit 1
}
