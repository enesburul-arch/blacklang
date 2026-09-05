param(
  [string]$Version = "",
  [ValidateSet("amd64", "arm64")]
  [string]$Arch = "amd64",
  [string]$OutRoot = "artifacts/releases"
)

$ErrorActionPreference = "Stop"

$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $ScriptRoot ".."))
$CliDir = Join-Path $RepoRoot "packages/cli"
$ReleaseRootBase = [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutRoot))

function Assert-ChildPath {
  param(
    [string]$Parent,
    [string]$Child,
    [string]$Label
  )

  $parentFull = [System.IO.Path]::GetFullPath($Parent).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
  $childFull = [System.IO.Path]::GetFullPath($Child).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
  if ($childFull -ne $parentFull -and -not $childFull.StartsWith($parentFull + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "$Label must stay under $parentFull, got $childFull"
  }
}

Assert-ChildPath -Parent $RepoRoot -Child $ReleaseRootBase -Label "Release output root"

$DistDir = Join-Path $RepoRoot "dist"
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

$BinaryPath = Join-Path $DistDir "black.exe"
$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
try {
  $env:GOOS = "windows"
  $env:GOARCH = $Arch
  Push-Location $CliDir
  try {
    go build -trimpath -o $BinaryPath ./cmd/black
  } finally {
    Pop-Location
  }
} finally {
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
}

$cliVersion = (& $BinaryPath version).Trim()
$normalizedCliVersion = if ($cliVersion.StartsWith("v")) { $cliVersion } else { "v$cliVersion" }
if ($Version -eq "") {
  $Version = $normalizedCliVersion
} elseif ($Version -ne $normalizedCliVersion) {
  throw "Requested release version $Version does not match CLI version $normalizedCliVersion"
}

$ArtifactName = "blacklang-$Version-windows-$Arch"
$ReleaseDir = [System.IO.Path]::GetFullPath((Join-Path $ReleaseRootBase $Version))
$StageParent = Join-Path $ReleaseDir ".staging"
$StageDir = Join-Path $StageParent $ArtifactName
$ArchivePath = Join-Path $ReleaseDir "$ArtifactName.zip"

Assert-ChildPath -Parent $ReleaseRootBase -Child $ReleaseDir -Label "Release directory"
Assert-ChildPath -Parent $ReleaseDir -Child $StageParent -Label "Staging directory"
Assert-ChildPath -Parent $StageParent -Child $StageDir -Label "Artifact staging directory"
Assert-ChildPath -Parent $ReleaseDir -Child $ArchivePath -Label "Archive path"

if (Test-Path -LiteralPath $StageParent) {
  Remove-Item -LiteralPath $StageParent -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $StageDir | Out-Null

Copy-Item -LiteralPath $BinaryPath -Destination (Join-Path $StageDir "black.exe")
Copy-Item -LiteralPath (Join-Path $RepoRoot "LICENSE") -Destination (Join-Path $StageDir "LICENSE")

$readme = @"
# BlackLang CLI $Version for Windows $Arch

This archive contains the BlackLang CLI executable.

## Verify

```powershell
.\black.exe version
.\black.exe --help
```

## Notes

This is a compiled CLI release artifact. Protected `.black` source files are not included.
"@

Set-Content -LiteralPath (Join-Path $StageDir "README.md") -Value $readme -Encoding utf8

$manifest = @"
package blacklang-cli
version $Version
os windows
arch $Arch
binary black.exe
command black
license MIT
source protected
"@

Set-Content -LiteralPath (Join-Path $StageDir "manifest.blackdir") -Value $manifest -Encoding utf8

if (Test-Path -LiteralPath $ArchivePath) {
  Remove-Item -LiteralPath $ArchivePath -Force
}
Compress-Archive -LiteralPath $StageDir -DestinationPath $ArchivePath
Remove-Item -LiteralPath $StageParent -Recurse -Force

Write-Output "built $ArchivePath"
