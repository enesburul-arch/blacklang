param(
  [string]$Version = "",
  [ValidateSet("dev", "preview", "stable")]
  [string]$Channel = "dev",
  [string]$ReleaseRoot = "artifacts/releases",
  [string]$Cli = "black"
)

$ErrorActionPreference = "Stop"

$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $ScriptRoot ".."))
$ReleaseRootBase = [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $ReleaseRoot))

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

function Normalize-Version {
  param([string]$Value)

  if ($Value.StartsWith("v")) {
    return $Value
  }
  return "v$Value"
}

function Get-ArchiveKind {
  param([string]$Name)

  if ($Name.EndsWith(".zip")) {
    return "zip"
  }
  if ($Name.EndsWith(".tar.gz")) {
    return "targz"
  }
  throw "Unsupported release archive extension: $Name"
}

function Parse-ArtifactTarget {
  param(
    [string]$Name,
    [string]$Version
  )

  $base = $Name
  if ($base.EndsWith(".tar.gz")) {
    $base = $base.Substring(0, $base.Length - ".tar.gz".Length)
  } elseif ($base.EndsWith(".zip")) {
    $base = $base.Substring(0, $base.Length - ".zip".Length)
  }

  $prefix = "blacklang-$Version-"
  if (-not $base.StartsWith($prefix)) {
    throw "Archive name must start with $prefix, got $Name"
  }

  $target = $base.Substring($prefix.Length)
  $lastDash = $target.LastIndexOf("-")
  if ($lastDash -lt 1 -or $lastDash -ge $target.Length - 1) {
    throw "Archive name must end with <os>-<arch>, got $Name"
  }

  return @{
    Os = $target.Substring(0, $lastDash)
    Arch = $target.Substring($lastDash + 1)
  }
}

if ($Version -eq "") {
  $BinaryPath = Join-Path $RepoRoot "dist/black.exe"
  if (-not (Test-Path -LiteralPath $BinaryPath)) {
    throw "Version was not provided and dist/black.exe does not exist. Run scripts/build-release-windows.ps1 first or pass -Version."
  }
  $Version = Normalize-Version ((& $BinaryPath version).Trim())
} else {
  $Version = Normalize-Version $Version
}

Assert-ChildPath -Parent $RepoRoot -Child $ReleaseRootBase -Label "Release root"

$ReleaseDir = [System.IO.Path]::GetFullPath((Join-Path $ReleaseRootBase $Version))
Assert-ChildPath -Parent $ReleaseRootBase -Child $ReleaseDir -Label "Release directory"

if (-not (Test-Path -LiteralPath $ReleaseDir)) {
  throw "Release directory does not exist: $ReleaseDir"
}

$archives = Get-ChildItem -LiteralPath $ReleaseDir -File |
  Where-Object { $_.Name -like "blacklang-$Version-*.zip" -or $_.Name -like "blacklang-$Version-*.tar.gz" } |
  Sort-Object Name

if ($archives.Count -eq 0) {
  throw "No release archives found in $ReleaseDir"
}

$checksumLines = New-Object System.Collections.Generic.List[string]
$manifestLines = New-Object System.Collections.Generic.List[string]
$manifestLines.Add("release $Version")
$manifestLines.Add("channel $Channel")
$manifestLines.Add("cli $Cli")
$manifestLines.Add("")

foreach ($archive in $archives) {
  Assert-ChildPath -Parent $ReleaseDir -Child $archive.FullName -Label "Release archive"
  $hash = (Get-FileHash -LiteralPath $archive.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
  $kind = Get-ArchiveKind $archive.Name
  $target = Parse-ArtifactTarget -Name $archive.Name -Version $Version
  $checksumLines.Add("$hash  $($archive.Name)")
  $manifestLines.Add("artifact $($target.Os) $($target.Arch) $kind $($archive.Name) sha256 $hash")
}

$checksumsPath = Join-Path $ReleaseDir "checksums.sha256"
$releaseManifestPath = Join-Path $ReleaseDir "release.blackdir"

Set-Content -LiteralPath $checksumsPath -Value $checksumLines -Encoding utf8
Set-Content -LiteralPath $releaseManifestPath -Value $manifestLines -Encoding utf8

Write-Output "wrote $checksumsPath"
Write-Output "wrote $releaseManifestPath"
