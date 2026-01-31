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

Write-Host "== smoke: Phase2 bootstrap ==" -ForegroundColor Cyan

$packet = $null

# Prefer existing case -> handoff packet
try {
  $cases = Invoke-RestMethod -Method Get -Uri "$base/cases?limit=1" -Headers $headers
  if ($cases.items -and $cases.items.Count -gt 0) {
    $caseId = $cases.items[0].id
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
  $entityId = "SMOKE-" + $suffix

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
        summary_md = "## phase2"
        industry_scope = "phase2"
        value_pool_notes = "phase2"
        key_questions = @("phase2")
      }
    }
  } | ConvertTo-Json -Depth 20)

  $hid = $handoff.id
  $got = Invoke-RestMethod -Method Get -Uri "$base/handoffs/$hid" -Headers $headers
  $packet = $got.packet
}

if (-not $packet) { throw "handoff packet not found" }

$resp = Invoke-RestMethod -Method Post -Uri "$base/phase2/runs" -Headers $headersJson -Body (@{
  packet = $packet
} | ConvertTo-Json -Depth 20)

$phase2 = $resp.packet.phases.phase2
if (-not $phase2.run_id) { throw "phase2.run_id missing" }
if (-not ($phase2.industry_candidates -is [System.Array])) { throw "phase2.industry_candidates not array" }
if (-not (
  $phase2.meta -is [System.Collections.IDictionary] -or
  $phase2.meta -is [System.Management.Automation.PSCustomObject]
)) { throw "phase2.meta not object" }

if ($phase2.industry_candidates.Count -gt 0) {
  foreach ($c in $phase2.industry_candidates) {
    if (-not $c.industry_id) { throw "industry_id missing" }
    if (-not $c.source) { throw "source missing" }
    if (-not $c.derived_from) { throw "derived_from missing" }
    if ($c.confidence -ne $null) { throw "confidence must be null" }
  }
}

if ($phase2.meta.phase1_total_events -ne $null -and -not ($phase2.meta.phase1_total_events -is [int] -or $phase2.meta.phase1_total_events -is [double])) {
  throw "phase1_total_events must be number"
}
if ($phase2.meta.phase1_last_seq -ne $null -and -not ($phase2.meta.phase1_last_seq -is [int] -or $phase2.meta.phase1_last_seq -is [double])) {
  throw "phase1_last_seq must be number"
}
if ($phase2.meta.phase1_finalized_present -ne $null -and -not ($phase2.meta.phase1_finalized_present -is [bool])) {
  throw "phase1_finalized_present must be bool"
}

Write-Host "[OK] phase2 bootstrap passed" -ForegroundColor Green
