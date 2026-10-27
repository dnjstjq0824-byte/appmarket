$ErrorActionPreference = 'Stop'
$root = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$evidence = Join-Path $root 'artifacts\evidence\AM-ARCH-001'
New-Item -ItemType Directory -Force -Path $evidence | Out-Null

Push-Location (Join-Path $root 'backend')
try {
    $checks = @(
        @{ Name = 'build'; Arguments = @('build', './...'); Log = 'go_build_current.log' },
        @{ Name = 'test'; Arguments = @('test', '-v', '-coverprofile=../artifacts/evidence/AM-ARCH-001/coverage-current.out', './...'); Log = 'go_test_current.log' },
        @{ Name = 'vet'; Arguments = @('vet', './...'); Log = 'go_vet_current.log' }
    )
    foreach ($check in $checks) {
        $logPath = Join-Path $evidence $check.Log
        & go @($check.Arguments) 2>&1 | Tee-Object -FilePath $logPath | Write-Output
        if ($LASTEXITCODE -ne 0) {
            throw "$($check.Name) failed with exit code $LASTEXITCODE; see $logPath"
        }
    }
} finally {
    Pop-Location
}
