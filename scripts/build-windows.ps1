$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectDir = Split-Path -Parent $scriptDir
$wailsVersion = "v2.15.0"
$wailsCommand = Get-Command wails -ErrorAction SilentlyContinue

if (-not $wailsCommand) {
    Write-Host "Wails CLI bulunamadi. $wailsVersion kuruluyor..."
    go install "github.com/wailsapp/wails/v2/cmd/wails@$wailsVersion"

    $goBin = go env GOPATH
    if ($goBin -is [array]) {
        $goBin = $goBin[0]
    }
    $goBin = Join-Path $goBin "bin"
    $env:Path = "$goBin;$env:Path"
    $wailsCommand = Get-Command wails -ErrorAction SilentlyContinue
}

if (-not $wailsCommand) {
    throw "Wails CLI bulunamadi. Go bin dizininin PATH icinde oldugunu kontrol edin."
}

Push-Location $projectDir
try {
    & $wailsCommand.Source build
    if ($LASTEXITCODE -ne 0) {
        throw "Wails build basarisiz oldu. Cikis kodu: $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

$outputFile = Join-Path $projectDir "build\bin\BrowserProfileViewer.exe"
if (-not (Test-Path $outputFile)) {
    throw "Build tamamlandi ancak beklenen cikti bulunamadi: $outputFile"
}

Write-Host "Build tamamlandi: $outputFile"