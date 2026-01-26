$ErrorActionPreference = "Stop"

function Write-FailLog($message) {
  $msg = $message
  if ($message -and $message.Exception) {
    $msg = $message.Exception.Message
  }
  Write-Host "[FAIL] $msg" -ForegroundColor Red
  Write-Host "== docker compose ps =="
  try { docker compose ps | Out-Host } catch {}
  Write-Host "== docker compose logs (api) =="
  try { docker compose logs --tail=200 api | Out-Host } catch {}
  Write-Host "== docker compose logs (db) =="
  try { docker compose logs --tail=200 db | Out-Host } catch {}
}

try {
  Set-Location (Split-Path $MyInvocation.MyCommand.Path) | Out-Null
  Set-Location .. | Out-Null

  Write-Host "== docker compose down -v ==" -ForegroundColor Cyan
  docker compose down -v

  Write-Host "== docker compose up -d --build ==" -ForegroundColor Cyan
  docker compose up -d --build

  Write-Host "== wait db ready ==" -ForegroundColor Cyan
  $maxWaitSeconds = 60
  $intervalSeconds = 2
  $deadline = (Get-Date).AddSeconds($maxWaitSeconds)
  $cid = $null
  $ready = $false

  while ((Get-Date) -lt $deadline) {
    $cid = docker ps -q --filter "name=investment-research-lab-db-1"
    if (-not $cid) {
      Start-Sleep -Seconds $intervalSeconds
      continue
    }
    docker exec $cid psql -U app -d investment -c "SELECT 1" | Out-Null
    if ($LASTEXITCODE -eq 0) {
      $ready = $true
      break
    }
    Start-Sleep -Seconds $intervalSeconds
  }

  if (-not $ready) {
    throw "db not ready within ${maxWaitSeconds}s"
  }

  Write-Host "== apply migration ==" -ForegroundColor Cyan
  Get-Content internal\db\migrations\0001_init.sql | docker exec -i $cid psql -U app -d investment | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "migration failed"
  }

  Write-Host "== wait api ready ==" -ForegroundColor Cyan
  $apiReady = $false
  $deadline = (Get-Date).AddSeconds($maxWaitSeconds)
  while ((Get-Date) -lt $deadline) {
    try {
      Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/universe/items?limit=1" -Headers @{ "X-API-Key"="devkey" } | Out-Null
      if ($LASTEXITCODE -eq 0) { $apiReady = $true; break }
    } catch {}
    Start-Sleep -Seconds $intervalSeconds
  }
  if (-not $apiReady) {
    throw "api not ready within ${maxWaitSeconds}s"
  }

  Write-Host "== run e2e ==" -ForegroundColor Cyan
  powershell -ExecutionPolicy Bypass -File scripts\e2e.ps1
  if ($LASTEXITCODE -ne 0) {
    throw "e2e failed"
  }

  Write-Host "[OK] E2E completed successfully" -ForegroundColor Green
  Write-Host "Next: try API quick examples in README.md" -ForegroundColor Cyan
}
catch {
  Write-FailLog $_
  throw
}
