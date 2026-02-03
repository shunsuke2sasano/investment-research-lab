CREATE TABLE IF NOT EXISTS phase3_runs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  input_packet jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_phase3_runs_created_at
ON phase3_runs(created_at DESC);
