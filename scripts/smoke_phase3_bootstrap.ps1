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

Write-Host "== smoke: Phase3 bootstrap ==" -ForegroundColor Cyan

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
  $entityId = "SMOKE3-" + $suffix

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
        summary_md = "## phase3"
        industry_scope = "phase3"
        value_pool_notes = "phase3"
        key_questions = @("phase3")
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

$version = $resp3.packet.version
if ($version -ne 1) { throw "packet.version must be 1" }

$phase3 = $resp3.packet.phases.phase3
if (-not $phase3.run_id) { throw "phase3.run_id missing" }

if (-not (
  $phase3.positioning -is [System.Collections.IDictionary] -or
  $phase3.positioning -is [System.Management.Automation.PSCustomObject]
)) { throw "phase3.positioning not object" }

if (-not ($phase3.positioning.target_customers -is [System.Array])) { throw "target_customers not array" }
if (-not ($phase3.positioning.key_competitors -is [System.Array])) { throw "key_competitors not array" }
if (-not ($phase3.positioning.differentiators -is [System.Array])) { throw "differentiators not array" }
if (-not ($phase3.positioning.notes -is [System.Array])) { throw "positioning.notes not array" }
if (-not ($phase3.positioning.value_prop -is [string])) { throw "value_prop not string" }

if (-not (
  $phase3.meta -is [System.Collections.IDictionary] -or
  $phase3.meta -is [System.Management.Automation.PSCustomObject]
)) { throw "phase3.meta not object" }

if ($phase3.meta.phase2_industry_candidates_count -ne $null -and -not ($phase3.meta.phase2_industry_candidates_count -is [int] -or $phase3.meta.phase2_industry_candidates_count -is [double])) {
  throw "phase2_industry_candidates_count must be number"
}
if ($phase3.meta.phase2_template_present -ne $null -and -not ($phase3.meta.phase2_template_present -is [bool])) {
  throw "phase2_template_present must be bool"
}
if ($phase3.meta.phase2_template_present -ne $true) { throw "phase2_template_present must be true" }
if ($phase3.meta.phase2_industry_candidates_count -lt 1) { throw "phase2_industry_candidates_count must be >= 1" }

Write-Host "[OK] phase3 bootstrap passed" -ForegroundColor Green
