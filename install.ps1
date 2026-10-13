# Install platform-cli on Windows (PowerShell)
param(
  [string]$Version = "latest",
  [string]$Dest = "$env:LOCALAPPDATA\platform-cli"
)

$repo = "cypher682/platform-cli"
$os = "windows"
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

if ($Version -eq "latest") {
  $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
  $tag = $resp.tag_name
} else {
  $tag = "v$Version"
}
$ver = $tag.TrimStart('v')
$url = "https://github.com/$repo/releases/download/$tag/platform-cli_${ver}_${os}_${arch}.zip"

$temp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "platform-cli-install") -Force
$zip = Join-Path $temp "platform-cli.zip"
Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $temp -Force
New-Item -ItemType Directory -Path $Dest -Force
Copy-Item -Path (Join-Path $temp "platform-cli.exe") -Destination (Join-Path $Dest "platform-cli.exe") -Force

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$Dest*") {
  [Environment]::SetEnvironmentVariable("Path", "$path;$Dest", "User")
  Write-Host "Added $Dest to user PATH. Restart shell or run: \$env:Path = \$env:Path + ';$Dest'"
}
Write-Host "Installed platform-cli $tag to $Dest"
