CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS universe_items (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  entity_type text NOT NULL CHECK (entity_type IN ('ticker','industry','theme','macro')),
  entity_id text NOT NULL,
  name text NOT NULL,
  keywords jsonb,
  priority int NOT NULL DEFAULT 50,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_universe_active
ON universe_items(is_active, priority);

CREATE TABLE IF NOT EXISTS runs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  phase int NOT NULL DEFAULT 1 CHECK (phase = 1),
  mode text NOT NULL DEFAULT 'manual' CHECK (mode IN ('manual','scheduled','event_driven')),
  status text NOT NULL DEFAULT 'running' CHECK (status IN ('running','success','failed')),
  config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  error text
);

CREATE INDEX IF NOT EXISTS idx_runs_started_at
ON runs(started_at DESC);

CREATE TABLE IF NOT EXISTS raw_items (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  source_type text NOT NULL,
  source_name text,
  url text NOT NULL,
  title text NOT NULL,
  published_at timestamptz,
  raw_text text NOT NULL,
  hash text NOT NULL UNIQUE,
  fetched_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_raw_items_run
ON raw_items(run_id);

CREATE INDEX IF NOT EXISTS idx_raw_items_published
ON raw_items(published_at DESC);

CREATE TABLE IF NOT EXISTS events (
  event_id text PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  observed_at timestamptz NOT NULL,
  entity_type text NOT NULL CHECK (entity_type IN ('ticker','industry','theme','macro')),
  entity_id text NOT NULL,
  category text NOT NULL CHECK (category IN (
    'earnings','ir','regulation','technology','macro','supply_demand','competition','price_action','other'
  )),
  title text NOT NULL,
  facts_json jsonb NOT NULL,
  impact_json jsonb,
  sources_json jsonb NOT NULL,
  confidence double precision NOT NULL CHECK (confidence BETWEEN 0 AND 1),
  dedupe_key text NOT NULL UNIQUE,
  tags_json jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_run
ON events(run_id);

CREATE INDEX IF NOT EXISTS idx_events_entity
ON events(entity_type, entity_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_category
ON events(category, observed_at DESC);

CREATE TABLE IF NOT EXISTS anomaly_summaries (
  run_id uuid PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
  summary_json jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS trigger_decisions (
  run_id uuid PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
  decision_json jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cases (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_type text NOT NULL CHECK (case_type IN ('ticker','industry','theme')),
  entity_id text NOT NULL,
  title text NOT NULL,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','paused','closed')),
  priority int NOT NULL DEFAULT 50,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cases_status
ON cases(status, priority);

CREATE TABLE IF NOT EXISTS handoff_packets (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  case_id uuid REFERENCES cases(id) ON DELETE SET NULL,
  handoff_type text NOT NULL CHECK (handoff_type IN ('light','heavy')),
  from_phase int NOT NULL DEFAULT 1 CHECK (from_phase = 1),
  to_phase int NOT NULL CHECK (to_phase IN (3,5)),
  packet_json jsonb NOT NULL,
  status text NOT NULL DEFAULT 'created' CHECK (status IN ('created','consumed','archived')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_handoff_run
ON handoff_packets(run_id);

CREATE INDEX IF NOT EXISTS idx_handoff_case
ON handoff_packets(case_id, created_at DESC);

CREATE TABLE IF NOT EXISTS phase_artifacts (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_id uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  phase int NOT NULL CHECK (phase BETWEEN 2 AND 7),
  artifact_type text NOT NULL,
  content_md text,
  content_json jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_artifacts_case_phase
ON phase_artifacts(case_id, phase, created_at DESC);

CREATE TABLE IF NOT EXISTS decisions (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_id uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  decision_date date NOT NULL DEFAULT CURRENT_DATE,
  overall_score int NOT NULL CHECK (overall_score BETWEEN 0 AND 100),
  final_label text NOT NULL CHECK (final_label IN ('BuyCandidate','Watch','Pass')),
  constraints_json jsonb NOT NULL DEFAULT '[]'::jsonb,
  judge_results_json jsonb NOT NULL,
  decision_md text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_decisions_case
ON decisions(case_id, decision_date DESC);

CREATE TABLE IF NOT EXISTS monitoring_plans (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_id uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  decision_id uuid REFERENCES decisions(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','closed')),
  plan_json jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_monitoring_active
ON monitoring_plans(status, updated_at DESC);

CREATE TABLE IF NOT EXISTS alerts (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  monitoring_plan_id uuid NOT NULL REFERENCES monitoring_plans(id) ON DELETE CASCADE,
  severity text NOT NULL CHECK (severity IN ('low','mid','high')),
  type text NOT NULL,
  message text NOT NULL,
  refs_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  acknowledged_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_alerts_plan
ON alerts(monitoring_plan_id, created_at DESC);
