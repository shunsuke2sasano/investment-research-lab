param(
  [switch]$NoReset
)

$ErrorActionPreference = "Stop"

if (-not $NoReset) {
  Write-Host "== reset_and_e2e ==" -ForegroundColor Cyan
  powershell -ExecutionPolicy Bypass -File scripts\reset_and_e2e.ps1
}

$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey" }
$headersJson = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

Write-Host "== smoke: Run -> Events -> Handoff ==" -ForegroundColor Cyan

$run = Invoke-RestMethod -Method Post -Uri "$base/phase1/runs" -Headers $headersJson -Body (@{
  mode = "manual"
  config = @{ sources=@("ir"); max_items_per_source=1 }
} | ConvertTo-Json -Depth 10)

$runId = $run.run_id

Invoke-RestMethod -Method Post -Uri "$base/phase1/runs/$runId/events" -Headers $headersJson -Body (@{
  event_type = "note.added"
  source = "manual"
  payload = @{ note="smoke" }
} | ConvertTo-Json -Depth 10) | Out-Null

$handoff = Invoke-RestMethod -Method Post -Uri "$base/handoffs" -Headers $headersJson -Body (@{
  run_id = $runId
  handoff_type = "heavy"
  from_phase = 1
  to_phase = 3
  packet = @{
    handoff_type = "heavy"
    from_phase = 1
    to_phase = 3
    universe_item_ids = @()
    event_ids = @()
    trigger_decision_id = ("td-" + $runId)
    created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    payload = @{
      summary_md = "## smoke"
      industry_scope = "smoke"
      value_pool_notes = "smoke"
      key_questions = @("smoke")
    }
  }
} | ConvertTo-Json -Depth 20)

$hid = $handoff.id
$got = Invoke-RestMethod -Method Get -Uri "$base/handoffs/$hid" -Headers $headers

if (-not $got.packet.version) { throw "handoff packet.version missing" }
if (-not $got.packet.phases.phase1.run_id) { throw "handoff packet.phases.phase1.run_id missing" }
if (-not $got.packet.phases.phase1.events) { throw "handoff packet.phases.phase1.events missing" }

Write-Host "[OK] smoke passed" -ForegroundColor Green
