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
Run the API examples script (single source of truth):

```powershell
powershell -ExecutionPolicy Bypass -File scripts\README_api_examples.ps1
# already running (reuse existing)
powershell -ExecutionPolicy Bypass -File scripts\README_api_examples.ps1 -NoReset
```

The script `scripts/README_api_examples.ps1` is the only authoritative source for API examples.

## Tips
- If `Select-String -Recurse` is unavailable, use: `Get-ChildItem -Recurse | Select-String -Pattern "..."`.
- 401 unauthorized: ensure `X-API-Key` is included.
- 404 not found: confirm `/api/v1` path (health is `/api/v1/health`).

## Verification (Quick)
```powershell
powershell -ExecutionPolicy Bypass -File scripts\reset_and_e2e.ps1
docker compose up -d --build
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/health" -Headers @{ "X-API-Key"="devkey" }
```
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
