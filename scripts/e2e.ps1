# scripts/e2e.ps1
$ErrorActionPreference = "Stop"

$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey" }
$headersJson = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

function Invoke-Api {
  param(
    [Parameter(Mandatory=$true)][string]$Method,
    [Parameter(Mandatory=$true)][string]$Uri,
    [Parameter(Mandatory=$true)][hashtable]$Headers,
    [string]$Body,
    [string]$Label
  )

  if ($Label) {
    Write-Host ("== " + $Label + " ==")
  }
  try {
    if ($Body) {
      return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $Headers -Body $Body
    }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $Headers
  }
  catch {
    $resp = $_.Exception.Response
    $status = $null
    $respBody = ""
    $path = $Uri
    try { $path = (New-Object System.Uri($Uri)).PathAndQuery } catch {}
    if ($resp) {
      try { $status = $resp.StatusCode.value__ } catch {}
      try {
        $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
        $respBody = $reader.ReadToEnd()
      } catch {}
    }
    Write-Host ("[HTTP FAIL] " + $Method + " " + $path) -ForegroundColor Red
    if ($status) { Write-Host ("status = " + $status) }
    if ($respBody) { Write-Host ("body = " + $respBody) }
    throw
  }
}

$ids = [ordered]@{}

Write-Host "== 1) Universe: create =="

$eid = "AAPL-" + (Get-Date -Format "yyyyMMddHHmmssfff")

$uBody = @{
  entity_type="ticker"
  entity_id=$eid
  name="Apple"
  keywords=@("iphone")
  priority=50
  is_active=$true
} | ConvertTo-Json -Depth 20


$u = Invoke-Api -Method Post -Uri "$base/universe/items" -Headers $headersJson -Body $uBody -Label "1) Universe: create"
$universeId = $u.id
$ids.universeId = $universeId
Write-Host ("universeId = " + $universeId)

$uList = Invoke-Api -Method Get -Uri "$base/universe/items?limit=3" -Headers $headers -Label "2) Universe: list"
$uList | ConvertTo-Json -Depth 20

Write-Host "== 2.5) Case: create =="

$caseBody = @{
  title    = ("Apple case " + $eid)
  case_type= "ticker"
  entity_id= $eid        # Universeとそろえる
  priority = 50
} | ConvertTo-Json -Depth 10

$case = Invoke-Api -Method Post -Uri "$base/cases" -Headers $headersJson -Body $caseBody -Label "2.5) Case: create"
$caseId = $case.id
$ids.caseId = $caseId
Write-Host ("caseId = " + $caseId)

Write-Host "== 2.6) Run: create (phase1) =="

$runBody = @{
  mode = "manual"
  config = @{
    sources = @("ir")
    max_items_per_source = 1
  }
} | ConvertTo-Json -Depth 10

$run = Invoke-Api -Method Post -Uri "$base/phase1/runs" -Headers $headersJson -Body $runBody -Label "2.6) Run: create (phase1)"
$runId = $run.run_id
$ids.runId = $runId
Write-Host ("runId = " + $runId)

Write-Host "== 2.7) Handoff: create (heavy 1->3) =="

$handoffPacket = @{
  handoff_type = "heavy"
  from_phase = 1
  to_phase = 3
  universe_item_ids = @($universeId)
  event_ids = @()
  trigger_decision_id = ("td-" + $runId)
  created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  payload = @{
    summary_md = "## Handoff (heavy)`n- reason: sample`n"
    industry_scope = "consumer electronics"
    value_pool_notes = "devices/services"
    key_questions = @("who has pricing power?","regulatory risk?")
  }
}

$handoffBody = @{
  run_id = $runId
  case_id = $caseId
  handoff_type = "heavy"
  from_phase = 1
  to_phase = 3
  packet = $handoffPacket
} | ConvertTo-Json -Depth 20

$handoff = Invoke-Api -Method Post -Uri "$base/handoffs" -Headers $headersJson -Body $handoffBody -Label "2.7) Handoff: create (heavy 1->3)"
$handoffId = $handoff.id
$ids.handoffId = $handoffId
Write-Host ("handoffId = " + $handoffId)

Write-Host "== 3) Artifacts: create phase2/3/4/5 =="

$art2 = @{
  phase = 2
  artifact_type = "industry_review"
  content_md = "## Industry Review`n- market structure`n- value pool`n- barriers`n- disruption risk`n"
  content_json = @{
    industry_scope="consumer electronics + services"
    value_pool=@("devices","services","ads")
    barriers=@("ecosystem","brand","distribution")
    disruption_risks=@("regulation","platform shift")
  }
} | ConvertTo-Json -Depth 20

$art3 = @{
  phase = 3
  artifact_type = "positioning_result"
  content_md = "## Positioning (AAPL)`n- who/what`n- differentiation`n- substitutes`n- pricing power`n"
  content_json = @{
    market="consumer electronics + services"
    moat_hypothesis=@("ecosystem lock-in","brand","distribution")
    key_risks=@("regulation","china supply chain")
  }
} | ConvertTo-Json -Depth 20

$art4 = @{
  phase = 4
  artifact_type = "stock_hypothesis"
  content_md = "## Stock Hypothesis`n- hypothesis`n- research`n- KPI`n"
  content_json = @{
    hypotheses=@("services mix improves margin")
    research_questions=@("what are key regulatory topics?","how big is china dependency?")
    kpis=@("gross_margin","services_growth","china_sales_ratio")
  }
} | ConvertTo-Json -Depth 20

$art5 = @{
  phase = 5
  artifact_type = "screening_result"
  content_md = "## Screening`n- result: PASS`n- reason: moat strong / finance OK`n"
  content_json = @{
    result="PASS"
    checks=@{
      debt_ok=$true
      cashflow_ok=$true
      moat_ok=$true
      redflags=@("regulation","china_supply_chain")
    }
  }
} | ConvertTo-Json -Depth 20

Invoke-Api -Method Post -Uri "$base/cases/$caseId/artifacts" -Headers $headersJson -Body $art2 -Label "3) Artifacts: create phase2" | Out-Null
Invoke-Api -Method Post -Uri "$base/cases/$caseId/artifacts" -Headers $headersJson -Body $art3 -Label "3) Artifacts: create phase3" | Out-Null
Invoke-Api -Method Post -Uri "$base/cases/$caseId/artifacts" -Headers $headersJson -Body $art4 -Label "3) Artifacts: create phase4" | Out-Null
Invoke-Api -Method Post -Uri "$base/cases/$caseId/artifacts" -Headers $headersJson -Body $art5 -Label "3) Artifacts: create phase5" | Out-Null

Write-Host "artifacts created"

Write-Host "== 4) Decision: create =="

$decBody = @{
  overall_score = 65
  final_label = "Watch"
  constraints = @("regulation","china_supply_chain")
  judge_results = @{
    verdict="monitor"
    confidence=0.65
    key_reasons=@("ecosystem lock-in","pricing power","regulatory risk")
  }
  decision_md = "## Decision (draft)`n- moat strong`n- watch regulation and china risk`n`nConclusion: monitor"
} | ConvertTo-Json -Depth 20

$dec = Invoke-Api -Method Post -Uri "$base/cases/$caseId/decisions" -Headers $headersJson -Body $decBody -Label "4) Decision: create"
$decisionId = $dec.id
$ids.decisionId = $decisionId
Write-Host ("decisionId = " + $decisionId)

Write-Host "== 5) Monitoring plan: create =="

$planBody = @{
  decision_id = $decisionId
  plan = @{
    kpis = @(
      @{ name="gross_margin"; rule="below"; threshold=0.35; window_days=90 },
      @{ name="services_growth"; rule="below"; threshold=0.08; window_days=180 }
    )
    events = @(
      @{ category="regulation"; severity="high"; note="alert on major regulation news" }
    )
    cadence = @{ review="weekly"; deep_review="quarterly" }
  }
} | ConvertTo-Json -Depth 20

$plan = Invoke-Api -Method Post -Uri "$base/cases/$caseId/monitoring-plans" -Headers $headersJson -Body $planBody -Label "5) Monitoring plan: create"
$planId = $plan.id
$ids.planId = $planId
Write-Host ("planId = " + $planId)

Write-Host "== 6) Alert: create =="

$alertBody = @{
  severity="high"
  type="regulation"
  message="possible EU regulation tightening"
  refs=@{ url="https://example.com"; note="source memo" }
} | ConvertTo-Json -Depth 10

$a = Invoke-Api -Method Post -Uri "$base/monitoring-plans/$planId/alerts" -Headers $headersJson -Body $alertBody -Label "6) Alert: create"
$alertId = $a.id
$ids.alertId = $alertId
Write-Host ("alertId = " + $alertId)

Write-Host "== 7) Alerts: list =="
$alerts = Invoke-Api -Method Get -Uri "$base/monitoring-plans/$planId/alerts" -Headers $headers -Label "7) Alerts: list"
$alerts | ConvertTo-Json -Depth 30

Write-Host "== 8) Alert: ack =="
$ack = Invoke-Api -Method Post -Uri "$base/alerts/$alertId/ack" -Headers $headers -Label "8) Alert: ack"
$ack | ConvertTo-Json -Depth 10

Write-Host "== 9) Case: get (final snapshot) =="
$snapshot = Invoke-Api -Method Get -Uri "$base/cases/$caseId" -Headers $headers -Label "9) Case: get (final snapshot)"
$handoffCount = ($snapshot.handoffs | Measure-Object).Count
$ids.handoffCount = $handoffCount
Write-Host ("handoffs count = " + $handoffCount)
if ($handoffCount -lt 1) { throw "handoffs count is ${handoffCount}" }
$snapshot | ConvertTo-Json -Depth 50

Write-Host "== 10) Case artifacts: list =="
$artifacts = Invoke-Api -Method Get -Uri "$base/cases/$caseId/artifacts" -Headers $headers -Label "10) Case artifacts: list"
$artifacts | ConvertTo-Json -Depth 50

Write-Host "== 11) Case decisions: list =="
$decisions = Invoke-Api -Method Get -Uri "$base/cases/$caseId/decisions" -Headers $headers -Label "11) Case decisions: list"
$decisions | ConvertTo-Json -Depth 50

Write-Host "== 12) Case monitoring plans: list =="
$plans = Invoke-Api -Method Get -Uri "$base/cases/$caseId/monitoring-plans" -Headers $headers -Label "12) Case monitoring plans: list"
$plans | ConvertTo-Json -Depth 50

Write-Host "== Summary ==" -ForegroundColor Cyan
$ids.GetEnumerator() | ForEach-Object { Write-Host ("{0} = {1}" -f $_.Key, $_.Value) }

