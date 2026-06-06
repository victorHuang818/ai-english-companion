<#
.SYNOPSIS
    Starts all Go backend services for the AI English Companion project on Windows.
.DESCRIPTION
    Loads environment variables from .env and runs user.rpc, core.rpc, ai.rpc, companion_api, and gateway in background jobs.
#>

# 1. Ensure we are in the backend directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# 2. Check for .env file and load it
$EnvFile = Join-Path $ScriptDir ".env"
if (Test-Path $EnvFile) {
    Write-Host "[Info] Loading environment variables from $EnvFile..." -ForegroundColor Cyan
    Get-Content $EnvFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#")) {
            $name, $value = $line -split '=', 2
            if ($name -and $value) {
                $varName = $name.Trim()
                $varVal = $value.Trim().Trim('"').Trim("'")
                [System.Environment]::SetEnvironmentVariable($varName, $varVal, [System.EnvironmentVariableTarget]::Process)
                Write-Host "       Set $varName" -ForegroundColor DarkGray
            }
        }
    }
} else {
    Write-Error "[Warning] .env file not found at $EnvFile. Services might fail to start if env vars are missing."
}

# 3. Check for dependencies (basic socket check)
function Check-Port ($Port, $Name) {
    $connection = New-Object System.Net.Sockets.TcpClient
    try {
        $connection.Connect("127.0.0.1", $Port)
        Write-Host "[OK] Dependency $Name is running on port $Port." -ForegroundColor Green
        $connection.Close()
        return $true
    } catch {
        Write-Host "[WARN] Dependency $Name (port $Port) seems offline. Please make sure it is running." -ForegroundColor Yellow
        return $false
    }
}

Write-Host "`n[1/3] Checking Infrastructure Dependencies..." -ForegroundColor Yellow
Check-Port 3306 "MySQL"
Check-Port 6379 "Redis"
Check-Port 9000 "Minio"

Write-Host "`n[2/3] Starting Backend Services in background jobs..." -ForegroundColor Yellow

$Services = @(
    @{ Name = "ai-rpc";       Path = "rpc/ai";       File = "ai.go";       Config = "etc/ai.yaml" },
    @{ Name = "user-rpc";     Path = "rpc/user";     File = "user.go";     Config = "etc/user.yaml" },
    @{ Name = "core-rpc";     Path = "rpc/core";     File = "core.go";     Config = "etc/core.yaml" },
    @{ Name = "companion-api";Path = "companion_api";File = "companion.go";Config = "etc/companion-api.yaml" },
    @{ Name = "gateway";      Path = "gateway";      File = "main.go";     Config = "etc/gateway.yaml" }
)

foreach ($svc in $Services) {
    Write-Host "       Launching service: $($svc.Name)..." -ForegroundColor Cyan
    Start-Job -Name $svc.Name -ScriptBlock {
        param($s, $dir)
        Set-Location (Join-Path $dir $s.Path)
        go run $s.File -f $s.Config
    } -ArgumentList $svc, $ScriptDir
    
    # Wait briefly between starting RPCs to allow ports to bind
    Start-Sleep -Seconds 1
}

Write-Host "`n[3/3] Backend Services Started!" -ForegroundColor Green
Write-Host "--------------------------------------------------------"
Write-Host "Services status (Jobs):"
Get-Job | Format-Table -Property ID, Name, State

Write-Host "`nTips:"
Write-Host " - Use 'Receive-Job -Name <name> -Keep' to view a service log."
Write-Host " - Use 'Stop-Job -Name *' or 'Get-Job | Remove-Job -Force' to stop all background services."
Write-Host "--------------------------------------------------------"
