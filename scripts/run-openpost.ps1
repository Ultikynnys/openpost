<#
.SYNOPSIS
    Runs the OpenPost app locally via Docker (Windows).

.DESCRIPTION
    Ensures the Docker engine is running, builds the local openpost-patched
    image from docker/Dockerfile when it is missing (or when -Rebuild is
    passed), force-recreates the docker-compose.yml service so every run
    starts a fresh container from the current image (a stale instance can
    never survive), waits for the health endpoint, and prints the local URL.

    Configuration (secrets, provider credentials) comes from the repository
    root .env file through docker-compose.yml's env_file. This script never
    reads or writes .env and contains no credentials.

.PARAMETER Rebuild
    Force a fresh image build instead of reusing an existing image.

.PARAMETER Open
    Open http://localhost:8080 in the default browser once the app is healthy.

.EXAMPLE
    .\scripts\run-openpost.ps1

.EXAMPLE
    .\scripts\run-openpost.ps1 -Rebuild -Open
#>
[CmdletBinding()]
param(
    [switch]$Rebuild,
    [switch]$Open
)

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$ImageTag = 'openpost-patched:latest'
$HealthUrl = 'http://localhost:8080/api/v1/health'

function Fail([string]$Message) {
    Write-Error $Message
    exit 1
}

function Test-DockerEngine {
    docker info *> $null
    return $LASTEXITCODE -eq 0
}

# --- 1. Docker engine ---------------------------------------------------------
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Fail 'Docker is not installed or not on PATH. Install Docker Desktop first.'
}

if (-not (Test-DockerEngine)) {
    Write-Host 'Docker engine is not running. Starting Docker Desktop...'
    $Desktop = 'C:\Program Files\Docker\Docker\Docker Desktop.exe'
    if (-not (Test-Path $Desktop)) {
        Fail 'Docker engine is down and Docker Desktop was not found. Start Docker Desktop manually and re-run this script.'
    }
    Start-Process $Desktop | Out-Null
    $Deadline = (Get-Date).AddSeconds(120)
    do {
        Start-Sleep -Seconds 3
    } while (-not (Test-DockerEngine) -and (Get-Date) -lt $Deadline)
    if (-not (Test-DockerEngine)) {
        Fail 'Docker engine did not start within 120 seconds. Check Docker Desktop and re-run this script.'
    }
    Write-Host 'Docker engine is up.'
}

# --- 2. Image -----------------------------------------------------------------
$HasImage = ((docker images -q $ImageTag 2> $null | Select-Object -First 1) -ne '')
if (-not $HasImage -or $Rebuild) {
    if ($Rebuild) {
        Write-Host "Rebuilding image $ImageTag ..."
    } else {
        Write-Host "Image $ImageTag not found. Building it (first build can take several minutes)..."
    }
    Push-Location $RepoRoot
    try {
        docker build -f docker/Dockerfile -t $ImageTag .
        if ($LASTEXITCODE -ne 0) {
            Fail 'Image build failed. See the build output above.'
        }
    } finally {
        Pop-Location
    }
}

# --- 3. Compose up (force-recreate so stale instances are never reused) ------
Push-Location $RepoRoot
try {
    docker compose up -d --force-recreate
    if ($LASTEXITCODE -ne 0) {
        Fail 'docker compose up failed. See the output above.'
    }
} finally {
    Pop-Location
}

# --- 4. Wait for health -------------------------------------------------------
Write-Host "Waiting for OpenPost on $HealthUrl ..."
$Deadline = (Get-Date).AddSeconds(180)
$Healthy = $false
while ((Get-Date) -lt $Deadline) {
    try {
        $Response = Invoke-WebRequest -Uri $HealthUrl -UseBasicParsing -TimeoutSec 5
        if ($Response.StatusCode -eq 200) {
            $Healthy = $true
            break
        }
    } catch {
        # App not accepting requests yet; keep polling.
    }
    Start-Sleep -Seconds 3
}

if (-not $Healthy) {
    Fail 'OpenPost did not become healthy within 180 seconds. Check the logs with: docker compose logs openpost'
}

Write-Host 'OpenPost is running at http://localhost:8080' -ForegroundColor Green
if ($Open) {
    Start-Process 'http://localhost:8080'
}
