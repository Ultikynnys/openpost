#requires -Version 5.1
<#
.SYNOPSIS
    Runs the OpenPost app locally via Docker (Windows) with a self-renewing
    public HTTPS tunnel.

.DESCRIPTION
    Ensures the Docker engine is running, builds the local openpost-patched
    image from docker/Dockerfile when it is missing (or when -Rebuild is
    passed), force-recreates the docker-compose.yml service so every run
    starts a fresh container from the current image (a stale instance can
    never survive), waits for the health endpoint, and prints the local URL.

    Public media URL: Threads, Instagram, Facebook, and TikTok require a
    public HTTPS media URL so their servers can fetch attachments. Unless
    -NoTunnel is passed, this script keeps OPENPOST_MEDIA_URL fresh: it
    reuses the previous localhost.run tunnel while its URL still answers
    /api/v1/health, otherwise it starts a new tunnel, reads the freshly
    assigned *.lhr.life URL, and rewrites only the OPENPOST_MEDIA_URL line in
    the repository .env. Docker Compose recreates the container because the
    env_file content changed. Secrets are never read or written.

    OUTPUT: Every step is logged in real time with timestamps straight to
    stdout (console AND pipes), so progress is always visible. On failure the
    script prints the failing command, the exit code, and diagnostic tails
    (tunnel logs, container logs) before exiting non-zero.

    Configuration (secrets, provider credentials) comes from the repository
    root .env file through docker-compose.yml's env_file. This script never
    reads secrets and contains no credentials.

.PARAMETER Rebuild
    Force a fresh image build instead of reusing an existing image.

.PARAMETER Open
    Open http://localhost:8080 in the default browser once the app is healthy.

.PARAMETER NoTunnel
    Skip tunnel management. Use when OPENPOST_MEDIA_URL already points at a
    stable public HTTPS URL (e.g. a real domain behind a reverse proxy).

.EXAMPLE
    .\scripts\run-openpost.ps1

.EXAMPLE
    .\scripts\run-openpost.ps1 -Rebuild -Open

.EXAMPLE
    .\scripts\run-openpost.ps1 -NoTunnel
#>
[CmdletBinding()]
param(
    [switch]$Rebuild,
    [switch]$Open,
    [switch]$NoTunnel
)

$ErrorActionPreference = 'Stop'

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$ImageTag = 'openpost-patched:latest'
$HealthUrl = 'http://localhost:8080/api/v1/health'
$EnvPath = Join-Path $RepoRoot '.env'
$TunnelLog = Join-Path $env:TEMP 'openpost-tunnel.log'
$TunnelErr = Join-Path $env:TEMP 'openpost-tunnel.err.log'
$TunnelPidFile = Join-Path $env:TEMP 'openpost-tunnel.pid'
$DockerDesktop = 'C:\Program Files\Docker\Docker\Docker Desktop.exe'

# ---------------------------------------------------------------------------
# Console output. PowerShell 5.1's Write-Host renders to the console host and
# vanishes when stdout is redirected (pipes, CI, log capture). Every message
# goes through Write-Log, which writes straight to the stdout pipe and flushes
# immediately, so progress is visible in real time everywhere.
# ---------------------------------------------------------------------------
function Write-Log {
    param([string]$Message)
    $Line = "[{0}] {1}" -f (Get-Date -Format 'HH:mm:ss'), $Message
    try {
        [Console]::Out.WriteLine($Line)
        [Console]::Out.Flush()
    } catch {
        Write-Output $Line
    }
}

function Write-Step {
    param([int]$Number, [string]$Title)
    Write-Log ''
    Write-Log ('===== STEP {0}/5: {1} =====' -f $Number, $Title)
}

function Fail([string]$Message) {
    Write-Log ''
    Write-Log '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    Write-Log ('FAILED: ' + $Message)
    Write-Log '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    exit 1
}

# ---------------------------------------------------------------------------
# Docker engine helpers
# ---------------------------------------------------------------------------
function Test-DockerEngine {
    $Version = docker info --format '{{.ServerVersion}}' 2>$null
    return -not [string]::IsNullOrWhiteSpace($Version)
}

function Get-DockerEngineError {
    $Output = (docker info 2>&1 | Out-String).Trim()
    return ($Output -replace '\s+', ' ')
}

# ---------------------------------------------------------------------------
# HTTP / tunnel helpers
# ---------------------------------------------------------------------------
function Test-UrlAlive([string]$url) {
    try {
        $Response = Invoke-WebRequest -Uri "$url/api/v1/health" -UseBasicParsing -TimeoutSec 10
        return $Response.StatusCode -eq 200
    } catch {
        return $false
    }
}

function Read-LogFile([string]$Path) {
    if (-not (Test-Path $Path)) { return '' }
    try {
        $Stream = [System.IO.File]::Open($Path, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read, [System.IO.FileShare]::ReadWrite)
        try {
            $Reader = New-Object System.IO.StreamReader($Stream)
            try { return $Reader.ReadToEnd() } finally { $Reader.Dispose() }
        } finally { $Stream.Dispose() }
    } catch {
        return ''
    }
}

function Get-TunnelUrlFromLog {
    # localhost.run prints the assigned URL on stdout; with ssh -tt the
    # banner can end up on stderr. Check both.
    foreach ($LogPath in @($TunnelLog, $TunnelErr)) {
        $Content = Read-LogFile $LogPath
        if ($Content -match 'https://[a-z0-9]+\.lhr\.life') { return $Matches[0] }
    }
    return ''
}

function Start-Tunnel {
    # Stop any tunnel a previous run left behind.
    if (Test-Path $TunnelPidFile) {
        $OldPid = Get-Content $TunnelPidFile -ErrorAction SilentlyContinue
        if ($OldPid) {
            $OldProcess = Get-Process -Id $OldPid -ErrorAction SilentlyContinue
            if ($OldProcess -and $OldProcess.ProcessName -like 'ssh*') {
                Write-Log "Stopping leftover tunnel process (pid $OldPid)."
                Stop-Process -Id $OldPid -Force -ErrorAction SilentlyContinue
            }
        }
    }
    Remove-Item $TunnelLog, $TunnelErr -ErrorAction SilentlyContinue

    $Ssh = (Get-Command ssh -ErrorAction SilentlyContinue).Source
    if (-not $Ssh) {
        Fail 'OpenSSH client (ssh) is required for the public tunnel. Install it or use -NoTunnel with a stable OPENPOST_MEDIA_URL.'
    }
    Write-Log "Using ssh: $Ssh"

    $SshArgs = '-tt -o StrictHostKeyChecking=no -o ServerAliveInterval=30 -o ExitOnForwardFailure=yes -R 80:localhost:8080 nokey@localhost.run'
    Write-Log "Starting ssh reverse tunnel (stdout -> $TunnelLog, stderr -> $TunnelErr)"
    $Process = Start-Process -FilePath $Ssh -ArgumentList $SshArgs -WindowStyle Hidden -RedirectStandardOutput $TunnelLog -RedirectStandardError $TunnelErr -PassThru
    Set-Content -Path $TunnelPidFile -Value $Process.Id
    Write-Log "Tunnel process started (pid $($Process.Id)). Waiting for localhost.run to assign a public URL (up to 90s)..."

    $Deadline = (Get-Date).AddSeconds(90)
    $TunnelUrl = ''
    while ((Get-Date) -lt $Deadline) {
        if ($Process.HasExited) {
            $ErrorText = if (Test-Path $TunnelErr) { (Get-Content $TunnelErr -Raw) -replace '\s+', ' ' } else { '' }
            Fail "Tunnel exited early (exit code $($Process.ExitCode)). $ErrorText"
        }
        $TunnelUrl = Get-TunnelUrlFromLog
        if ($TunnelUrl) {
            Write-Log "Public URL assigned by localhost.run: $TunnelUrl"
            break
        }
        Start-Sleep -Seconds 3
    }
    if (-not $TunnelUrl) {
        $LogTail = if (Test-Path $TunnelLog) { (Get-Content $TunnelLog -Tail 10) -join ' | ' } else { '(no stdout)' }
        $ErrTail = if (Test-Path $TunnelErr) { (Get-Content $TunnelErr -Tail 10) -join ' | ' } else { '(no stderr)' }
        Fail "Tunnel did not hand out a public URL within 90 seconds. stdout: $LogTail | stderr: $ErrTail"
    }
    return $TunnelUrl
}

function Set-MediaUrlInEnv([string]$PublicUrl) {
    $MediaBase = "$PublicUrl/media"
    if (-not (Test-Path $EnvPath)) {
        Fail ".env not found at $EnvPath. Create it from .env.example first."
    }
    $Lines = [System.IO.File]::ReadAllLines($EnvPath)
    $Found = $false
    $Changed = $false
    for ($i = 0; $i -lt $Lines.Count; $i++) {
        if ($Lines[$i] -match '^\s*OPENPOST_MEDIA_URL\s*=') {
            $Found = $true
            if ($Lines[$i].Trim() -ne "OPENPOST_MEDIA_URL=$MediaBase") {
                Write-Log "OPENPOST_MEDIA_URL in .env is '$($Lines[$i].Trim())'; updating to '$MediaBase'"
                $Lines[$i] = "OPENPOST_MEDIA_URL=$MediaBase"
                $Changed = $true
            } else {
                Write-Log "OPENPOST_MEDIA_URL in .env already matches ($MediaBase)."
            }
            break
        }
    }
    if (-not $Found) {
        Write-Log 'OPENPOST_MEDIA_URL not found in .env; appending it.'
        $Lines += "OPENPOST_MEDIA_URL=$MediaBase"
        $Changed = $true
    }
    if ($Changed) {
        [System.IO.File]::WriteAllLines($EnvPath, $Lines, (New-Object System.Text.UTF8Encoding($false)))
        Write-Log ".env updated (only the OPENPOST_MEDIA_URL line was touched) -> $MediaBase"
    }
    return $MediaBase
}

# ---------------------------------------------------------------------------
# Image staleness: find the newest file among the docker build inputs, walking
# the tree WITHOUT descending into node_modules/.git/build output dirs (the
# raw -Recurse filter applies after enumeration, which crawls ~100k files).
# ---------------------------------------------------------------------------
function Get-NewestSourceTime {
    param([string[]]$Roots)
    $ExcludedDirs = @('.git', '.hg', 'node_modules', '.svelte-kit', '.vite', '.turbo', '.wrangler', '.pnpm-store', '.devenv', 'build', 'dist', 'media', 'test-results', 'playwright-report', 'coverage')
    $Newest = [datetime]::MinValue
    $NewestPath = ''
    $Stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $Pending = New-Object 'System.Collections.Generic.Queue[string]'
    foreach ($r in $Roots) { $Pending.Enqueue((Join-Path $RepoRoot $r)) }
    while ($Pending.Count -gt 0) {
        $Dir = $Pending.Dequeue()
        foreach ($Item in Get-ChildItem -LiteralPath $Dir -Force -ErrorAction SilentlyContinue) {
            if ($Item.PSIsContainer) {
                if ($ExcludedDirs -notcontains $Item.Name) { $Pending.Enqueue($Item.FullName) }
            } else {
                # The compiled backend binary is a build artifact, not an input.
                if ($Item.FullName -eq (Join-Path $RepoRoot 'backend\openpost')) { continue }
                if ($Item.LastWriteTimeUtc -gt $Newest) {
                    $Newest = $Item.LastWriteTimeUtc
                    $NewestPath = $Item.FullName
                }
            }
        }
    }
    $Stopwatch.Stop()
    Write-Log ("  scanned {0} root(s) in {1:n0} ms" -f $Roots.Count, $Stopwatch.ElapsedMilliseconds)
    return [pscustomobject]@{ Time = $Newest; Path = $NewestPath }
}

# ===========================================================================
# 0. Banner
# ===========================================================================
Write-Log 'OpenPost local runner (Docker)'
Write-Log "  repo root : $RepoRoot"
Write-Log "  image tag : $ImageTag"
Write-Log "  health    : $HealthUrl"
Write-Log "  tunnel    : $(if ($NoTunnel) { 'disabled (-NoTunnel)' } else { 'enabled (localhost.run)' })"
Write-Log "  env file  : $EnvPath"

# ===========================================================================
# 1. Docker engine
# ===========================================================================
Write-Step 1 'Docker engine'

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Fail 'Docker is not installed or not on PATH. Install Docker Desktop first.'
}

if (Test-DockerEngine) {
    $EngineVersion = docker info --format '{{.ServerVersion}}' 2>$null
    Write-Log "Docker engine is up (server version $EngineVersion)."
} else {
    $EngineError = Get-DockerEngineError
    Write-Log 'Docker engine is not running.'
    if ($EngineError) { Write-Log "  engine says: $EngineError" }
    if (-not (Test-Path $DockerDesktop)) {
        Fail "Docker engine is down and Docker Desktop was not found at $DockerDesktop. Start Docker Desktop manually and re-run this script."
    }
    Write-Log "Starting Docker Desktop: $DockerDesktop"
    Start-Process $DockerDesktop | Out-Null
    Write-Log 'Waiting for the engine to come up (up to 180s)...'
    $Deadline = (Get-Date).AddSeconds(180)
    $Attempt = 0
    while ((Get-Date) -lt $Deadline) {
        Start-Sleep -Seconds 5
        $Attempt++
        if (Test-DockerEngine) {
            Write-Log ("Docker engine is up after ~{0}s." -f ($Attempt * 5))
            break
        }
        if ($Attempt % 6 -eq 0) {
            Write-Log ("  still waiting for Docker engine ({0}s elapsed)..." -f ($Attempt * 5))
        }
    }
    if (-not (Test-DockerEngine)) {
        Fail ("Docker engine did not start within 180 seconds. Last engine message: {0}" -f (Get-DockerEngineError))
    }
}

# ===========================================================================
# 2. App image
# ===========================================================================
Write-Step 2 'App image'

$HasImage = ((docker images -q $ImageTag 2>$null | Select-Object -First 1) -ne '')
if ($HasImage) {
    Write-Log "Image $ImageTag exists."
} else {
    Write-Log "Image $ImageTag not found."
}

$NeedBuild = $Rebuild -or (-not $HasImage)

if ($HasImage -and -not $Rebuild) {
    Write-Log 'Checking whether any source file is newer than the image...'
    $ImageCreated = docker inspect -f '{{.Created}}' $ImageTag 2>$null
    if ($ImageCreated -and $ImageCreated.Contains('.')) {
        # Docker timestamps carry 9 fractional digits, which Windows
        # PowerShell 5.1 cannot parse; truncate them.
        $ImageCreated = $ImageCreated.Substring(0, $ImageCreated.IndexOf('.')) + 'Z'
    }
    try {
        # Parse as UTC so the comparison against LastWriteTimeUtc is exact
        # regardless of the machine's time zone.
        $ImageDate = [datetime]::Parse($ImageCreated, [System.Globalization.CultureInfo]::InvariantCulture, [System.Globalization.DateTimeStyles]::AdjustToUniversal)
    } catch {
        Write-Log "Could not parse image timestamp '$ImageCreated'; forcing a rebuild to be safe."
        $NeedBuild = $true
    }
    if ($ImageDate) {
        $ImageInputs = @('package.json', 'pnpm-lock.yaml', 'pnpm-workspace.yaml', 'turbo.json', 'frontend', 'packages', 'assets', 'scripts', 'backend', 'docker')
        $NewestSource = Get-NewestSourceTime $ImageInputs
        if ($NewestSource.Time -gt $ImageDate) {
            Write-Log ('Image is STALE: newer source file found.')
            Write-Log ("  image built : {0} (UTC)" -f $ImageDate.ToString('u'))
            Write-Log ("  newest file : $($NewestSource.Path)")
            Write-Log ("  file mtime  : {0} (UTC)" -f $NewestSource.Time.ToString('u'))
            $NeedBuild = $true
        } else {
            Write-Log ("Image is fresh. Newest source file mtime {0} (UTC) predates the image build." -f $NewestSource.Time.ToString('u'))
        }
    }
}

if ($NeedBuild) {
    if ($Rebuild) {
        Write-Log 'Rebuilding image (forced via -Rebuild)...'
    } elseif (-not $HasImage) {
        Write-Log 'Building image for the first time (can take several minutes)...'
    } else {
        Write-Log 'Rebuilding stale image (can take several minutes)...'
    }
    Write-Log '  docker build -f docker/Dockerfile -t openpost-patched:latest .'
    Push-Location $RepoRoot
    try {
        docker build -f docker/Dockerfile -t $ImageTag .
        if ($LASTEXITCODE -ne 0) {
            Fail "Image build failed (exit code $LASTEXITCODE). See the build output above."
        }
    } finally {
        Pop-Location
    }
    Write-Log 'Image build succeeded.'
} else {
    Write-Log 'No rebuild needed.'
}

# ===========================================================================
# 3. Public tunnel + media URL sync
# ===========================================================================
if (-not $NoTunnel) {
    Write-Step 3 'Public HTTPS tunnel (localhost.run)'

    $TunnelUrl = Get-TunnelUrlFromLog
    if ($TunnelUrl -and (Test-UrlAlive $TunnelUrl)) {
        Write-Log "Reusing live tunnel: $TunnelUrl"
    } else {
        if ($TunnelUrl) {
            Write-Log "Previous tunnel URL $TunnelUrl no longer answers /api/v1/health; starting a fresh tunnel."
        } else {
            Write-Log 'No previous tunnel found.'
        }
        $TunnelUrl = Start-Tunnel
    }
    $MediaBase = Set-MediaUrlInEnv $TunnelUrl
} else {
    Write-Log 'Tunnel skipped (-NoTunnel). OPENPOST_MEDIA_URL must already point at a stable public HTTPS URL.'
}

# ===========================================================================
# 4. Compose up (force-recreate so stale instances are never reused)
# ===========================================================================
Write-Step 4 'Docker Compose (force-recreate)'

Push-Location $RepoRoot
try {
    Write-Log 'Running: docker compose up -d --force-recreate'
    docker compose up -d --force-recreate
    if ($LASTEXITCODE -ne 0) {
        Fail "docker compose up failed (exit code $LASTEXITCODE). See the output above."
    }
    Write-Log 'Container status after start:'
    $PsOutput = (docker compose ps 2>&1 | Out-String).Trim()
    Write-Log ($PsOutput -replace '(?m)^', '  ')
} finally {
    Pop-Location
}

# ===========================================================================
# 5. Wait for health
# ===========================================================================
Write-Step 5 'Health check'

Write-Log "Waiting for OpenPost at $HealthUrl (up to 180s)..."
$Deadline = (Get-Date).AddSeconds(180)
$Healthy = $false
$LastError = ''
$Poll = 0
while ((Get-Date) -lt $Deadline) {
    $Poll++
    try {
        $Response = Invoke-WebRequest -Uri $HealthUrl -UseBasicParsing -TimeoutSec 5
        if ($Response.StatusCode -eq 200) {
            $Healthy = $true
            break
        }
        $LastError = "HTTP $($Response.StatusCode)"
    } catch {
        $LastError = $_.Exception.Message
    }
    if ($Poll % 2 -eq 0) {
        $Elapsed = $Poll * 3
        Write-Log ("  not healthy yet (poll {0}, ~{1}s elapsed): {2}" -f $Poll, $Elapsed, (($LastError -split "`r?`n")[0]))
    }
    Start-Sleep -Seconds 3
}

if (-not $Healthy) {
    Write-Log "OpenPost did not become healthy within 180s. Last error: $LastError"
    Write-Log 'Recent container logs (docker logs --tail=40 openpost):'
    $ContainerLogs = (docker logs --tail=40 openpost 2>&1 | Out-String).Trim()
    Write-Log ($ContainerLogs -replace '(?m)^', '  ')
    Fail 'OpenPost did not become healthy within 180 seconds. See the container logs above; full logs with: docker compose logs openpost'
}

Write-Log ("OpenPost is healthy (after ~{0}s of polling)." -f ($Poll * 3))

# ===========================================================================
# Summary
# ===========================================================================
Write-Log ''
Write-Log '=============================================================='
Write-Log 'OpenPost is running at http://localhost:8080'
if (-not $NoTunnel) {
    Write-Log "Public media URL: $MediaBase (fresh per tunnel session)"
    if (Test-UrlAlive $MediaBase) {
        Write-Log 'Public media URL verified: it answers /api/v1/health through the tunnel.'
    } else {
        Write-Log 'WARNING: tunnel URL not answering yet (tunnel may still be warming up). The tunnel forwards to localhost:8080, so it only works while the app is up.'
    }
}
Write-Log 'Useful commands:'
Write-Log '  docker compose logs -f openpost   # live app logs'
Write-Log '  docker compose ps                 # container status'
Write-Log '  docker compose down               # stop everything'
Write-Log '=============================================================='

if ($Open) {
    Write-Log 'Opening http://localhost:8080 in the default browser...'
    Start-Process 'http://localhost:8080'
}
