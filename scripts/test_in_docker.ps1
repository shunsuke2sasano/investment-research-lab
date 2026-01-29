$ErrorActionPreference = "Stop"

Write-Host "== build builder image ==" -ForegroundColor Cyan
$tag = "investment-research-lab-builder"
docker build --target builder -t $tag .

Write-Host "== go test ./... ==" -ForegroundColor Cyan
$pwdPath = (Get-Location).Path

docker run --rm -v "${pwdPath}:/app" -w /app $tag go test ./... -count=1

Write-Host "Smoke test: run scripts\\smoke_phase1_events_handoff.ps1" -ForegroundColor Cyan
