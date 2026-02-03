param(
  [switch]$NoReset
)

$ErrorActionPreference = "Stop"

if (-not $NoReset) {
  Write-Host "== README API Examples (no reset) ==" -ForegroundColor Cyan
}

$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey" }
$headersJson = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

# Health
Invoke-RestMethod -Method Get -Uri "$base/health" -Headers $headers

# Phase2/3/4 bootstrap examples (verbatim from scripts/smoke_phase4_bootstrap.ps1)
$packet = $null

# Prefer existing case -> handoff packet
try {
  $cases = Invoke-RestMethod -Method Get -Uri "$base/cases?limit=1" -Headers $headers
  $items = $null
  if ($cases.items) { $items = $cases.items }
  elseif ($cases -is [System.Array]) { $items = $cases }

  if ($items -and $items.Count -gt 0) {
    $caseId = $items[0].id
    try {
      $detail = Invoke-RestMethod -Method Get -Uri "$base/cases/$caseId" -Headers $headers
      if ($detail.handoffs -and $detail.handoffs.Count -gt 0) {
        $handoff = $detail.handoffs | Where-Object { $_.packet } | Select-Object -First 1
        if ($handoff -and $handoff.packet) { $packet = $handoff.packet }
      }
    } catch {
      # fallback
    }
  }
} catch {
  # fallback
}

# Fallback: create minimal resources -> create handoff -> get packet
if (-not $packet) {
  $suffix = (Get-Date).ToUniversalTime().ToString("yyyyMMddHHmmss")
  $entityId = "SMOKE4-" + $suffix

  $u = Invoke-RestMethod -Method Post -Uri "$base/universe/items" -Headers $headersJson -Body (@{
    entity_type = "ticker"
    entity_id = $entityId
    name = "SmokeCo"
    keywords = @("smoke")
    priority = 50
    is_active = $true
  } | ConvertTo-Json -Depth 10)

  $case = Invoke-RestMethod -Method Post -Uri "$base/cases" -Headers $headersJson -Body (@{
    case_type = "ticker"
    entity_id = $entityId
    title = "Smoke case $entityId"
    priority = 50
  } | ConvertTo-Json -Depth 10)

  $run = Invoke-RestMethod -Method Post -Uri "$base/phase1/runs" -Headers $headersJson -Body (@{
    mode = "manual"
    config = @{ sources=@("ir"); max_items_per_source=1 }
  } | ConvertTo-Json -Depth 10)
  $runId = $run.run_id

  $handoff = Invoke-RestMethod -Method Post -Uri "$base/handoffs" -Headers $headersJson -Body (@{
    run_id = $runId
    case_id = $case.id
    handoff_type = "heavy"
    from_phase = 1
    to_phase = 3
    packet = @{
      handoff_type = "heavy"
      from_phase = 1
      to_phase = 3
      universe_item_ids = @($u.id)
      event_ids = @()
      trigger_decision_id = ("td-" + $runId)
      created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
      payload = @{
        summary_md = "## phase4"
        industry_scope = "phase4"
        value_pool_notes = "phase4"
        key_questions = @("phase4")
      }
    }
  } | ConvertTo-Json -Depth 20)

  $hid = $handoff.id
  $got = Invoke-RestMethod -Method Get -Uri "$base/handoffs/$hid" -Headers $headers
  $packet = $got.packet
}

if (-not $packet) { throw "handoff packet not found" }

$resp2 = Invoke-RestMethod -Method Post -Uri "$base/phase2/runs" -Headers $headersJson -Body (@{
  packet = $packet
} | ConvertTo-Json -Depth 20)

$resp3 = Invoke-RestMethod -Method Post -Uri "$base/phase3/runs" -Headers $headersJson -Body (@{
  packet = $resp2.packet
} | ConvertTo-Json -Depth 20)

$resp4 = Invoke-RestMethod -Method Post -Uri "$base/phase4/runs" -Headers $headersJson -Body (@{
  packet = $resp3.packet
} | ConvertTo-Json -Depth 20)

Write-Host "[OK] README API examples completed" -ForegroundColor Green
