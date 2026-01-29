CREATE TABLE IF NOT EXISTS phase1_run_events (
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  seq int NOT NULL,
  event_type text NOT NULL,
  source text NOT NULL DEFAULT 'manual',
  occurred_at timestamptz NOT NULL DEFAULT now(),
  payload_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(run_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_phase1_run_events_run
ON phase1_run_events(run_id);
