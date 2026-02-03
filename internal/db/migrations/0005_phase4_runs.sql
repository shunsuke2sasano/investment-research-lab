CREATE TABLE IF NOT EXISTS phase4_runs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  input_packet jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_phase4_runs_created_at
ON phase4_runs(created_at DESC);
