# investment-research-lab

Minimal Go API service.

## Quick Start (Docker)
```powershell
docker compose up -d --build
```

## One-command E2E (clean reset + migrate + E2E, non-interactive)
```powershell
powershell -ExecutionPolicy Bypass -File scripts\reset_and_e2e.ps1
```

## Troubleshooting
- Port 8080 in use: stop other services or change `docker-compose.yml` port mapping.
- 405 on endpoints: ensure you are using the correct URL and method (see API examples).
- FK errors: run the reset script to start from a clean DB (`docker compose down -v`).
- Mojibake/BOM: PowerShell scripts are saved as UTF-8 with BOM; avoid editing with tools that strip BOM.

## API Examples (PowerShell)
```powershell
$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

# Universe create
$u = Invoke-RestMethod -Method Post -Uri "$base/universe/items" -Headers $headers -Body (@{
  entity_type="ticker"; entity_id="AAPL-20260101"; name="Apple"; keywords=@("iphone"); priority=50; is_active=$true
} | ConvertTo-Json -Depth 10)

# Case create (align entity_id with Universe)
$case = Invoke-RestMethod -Method Post -Uri "$base/cases" -Headers $headers -Body (@{
  title="Apple case AAPL-20260101"; case_type="ticker"; entity_id="AAPL-20260101"; priority=50
} | ConvertTo-Json -Depth 10)

# Run create (Phase1)
$run = Invoke-RestMethod -Method Post -Uri "$base/phase1/runs" -Headers $headers -Body (@{
  mode="manual"; config=@{ sources=@("ir"); max_items_per_source=1 }
} | ConvertTo-Json -Depth 10)

# Handoff create (heavy 1->3) + case_id
$handoff = Invoke-RestMethod -Method Post -Uri "$base/handoffs" -Headers $headers -Body (@{
  run_id=$run.run_id
  case_id=$case.id
  handoff_type="heavy"
  from_phase=1
  to_phase=3
  packet=@{
    handoff_type="heavy"
    from_phase=1
    to_phase=3
    universe_item_ids=@($u.id)
    event_ids=@()
    trigger_decision_id=("td-"+$run.run_id)
    created_at=(Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    payload=@{
      summary_md="## Handoff (heavy)"
      industry_scope="consumer electronics"
      value_pool_notes="devices/services"
      key_questions=@("who has pricing power?","regulatory risk?")
    }
  }
} | ConvertTo-Json -Depth 20)

# Case snapshot (handoffs should be an array)
Invoke-RestMethod -Uri "$base/cases/$($case.id)" -Headers @{ "X-API-Key"="devkey" } | ConvertTo-Json -Depth 20

# Monitoring plans list (case scoped)
Invoke-RestMethod -Uri "$base/cases/$($case.id)/monitoring-plans" -Headers @{ "X-API-Key"="devkey" } | ConvertTo-Json -Depth 20
```

## Tests
### Host
```powershell
go test ./...
```

### Docker (builder image)
```powershell
docker build --target builder -t investment-research-lab-builder .
docker run --rm -v ${PWD}:/app -w /app investment-research-lab-builder go test ./... -count=1
```

### Smoke (Phase1 events -> handoff)
```powershell
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase1_events_handoff.ps1
# already running (reuse existing)
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase1_events_handoff.ps1 -NoReset
```

### Smoke (Phase2 bootstrap)
```powershell
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase2_bootstrap.ps1
# already running (reuse existing)
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase2_bootstrap.ps1 -NoReset
```

## Notes
- `case_id` in handoff creation links the handoff to the case for `/cases/{caseId}` snapshots.
- `/cases/{caseId}/monitoring-plans` supports GET list + POST create.

## Phase1 event_type registry
- run.finalized
- doc.fetched
- note.added
- signal.detected
- universe.member_added

## Phase1 event source registry
- manual
- system
- sec
- edinet
- other

## Phase1 Run Events API (examples)
```powershell
$base = "http://localhost:8080/api/v1"
$headers = @{ "X-API-Key"="devkey"; "Content-Type"="application/json" }

# POST event
Invoke-RestMethod -Method Post -Uri "$base/phase1/runs/{run_id}/events" -Headers $headers -Body (@{
  event_type="note.added"
  source="manual"
  payload=@{ note="example" }
} | ConvertTo-Json -Depth 10)

# GET events
Invoke-RestMethod -Method Get -Uri "$base/phase1/runs/{run_id}/events?limit=50" -Headers @{ "X-API-Key"="devkey" }
```

## Handoff packet schema (versioned)
```json
{
  "version": 1,
  "phases": {
    "phase1": { "run_id": "uuid", "events": [], "meta": {} }
  },
  "run_events": [],
  "phase1": { "run_id": "uuid", "events": [], "meta": {} }
}
```

## Phase1 Run config (sources)
```powershell
Invoke-RestMethod -Method Post -Uri "$base/phase1/runs" -Headers $headers -Body (@{
  mode="manual"
  config=@{ sources=@("ir"); max_items_per_source=1 }
} | ConvertTo-Json -Depth 10)
```
Note: `ir` fetcher is a stub implementation for now (no external calls).

## Phase2 Run bootstrap (Phase1 handoff -> Phase2 run)
```powershell
Invoke-RestMethod -Method Post -Uri "$base/phase2/runs" -Headers $headers -Body (@{
  packet = $handoff.packet
} | ConvertTo-Json -Depth 20)
```

### Smoke (Phase1 fetch -> doc.fetched)
```powershell
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase1_fetch.ps1
# already running (reuse existing)
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase1_fetch.ps1 -NoReset
# alternate sources
powershell -ExecutionPolicy Bypass -File scripts\smoke_phase1_fetch.ps1 -NoReset -Sources sec,edinet
```
