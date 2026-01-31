param(
  [switch]$NoReset,
  [string[]]$Sources = @("ir")
)

$ErrorActionPreference = "Stop"

if (-not $NoReset) {
  Write-Host "== reset_and_e2e ==" -ForegroundColor Cyan
  powershell -ExecutionPolicy Bypass -File scripts\reset_and_e2e.ps1
}

$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey" }
$headersJson = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

Write-Host "== smoke: Phase1 fetch -> doc.fetched ==" -ForegroundColor Cyan

$run = Invoke-RestMethod -Method Post -Uri "$base/phase1/runs" -Headers $headersJson -Body (@{
  mode = "manual"
  config = @{ sources=$Sources; max_items_per_source=1 }
} | ConvertTo-Json -Depth 10)

$runId = $run.run_id

$events = Invoke-RestMethod -Method Get -Uri "$base/phase1/runs/$runId/events?limit=50" -Headers $headers

$found = $false
foreach ($e in $events.items) {
  if ($e.event_type -eq "doc.fetched") { $found = $true; break }
}

if (-not $found) { throw "doc.fetched not found" }

Write-Host "[OK] doc.fetched found" -ForegroundColor Green
