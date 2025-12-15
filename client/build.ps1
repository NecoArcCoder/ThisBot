param (
    [ValidateSet("release", "debug")]
    [string]$Mode = "release"
)

$env:GOOS = "windows"
$env:GOARCH = "amd64"

$OutDir = "..\bin\bin"
$OutExe = "$OutDir\Stub.exe"

if (!(Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

if ($Mode -eq "release") {
    Write-Host "[+] Building RELEASE version"

    go build `
        -trimpath `
        -ldflags "-H=windowsgui -s -w" `
        -o $OutExe

} elseif ($Mode -eq "debug") {
    Write-Host "[+] Building DEBUG version"

    go build `
        -gcflags "all=-N -l" `
        -o $OutExe
}
