package handlers

import (
	"context"
	"encoding/json"
	"time"

	"investment_committee/internal/db/models"
	"investment_committee/internal/db/queries"
	"investment_committee/internal/domain"
)

type StoreAdapter struct {
	repo *queries.Repository
}

func NewStoreAdapter(repo *queries.Repository) *StoreAdapter {
	return &StoreAdapter{repo: repo}
}

func (s *StoreAdapter) CreateUniverseItem(ctx context.Context, in UniverseItemInput) (string, error) {
	item := models.UniverseItem{
		EntityType: in.EntityType,
		EntityID:   in.EntityID,
		Name:       in.Name,
		Priority:   in.Priority,
		IsActive:   in.IsActive,
	}
	raw, _ := json.Marshal(in.Keywords)
	item.Keywords = raw
	return s.repo.CreateUniverseItem(ctx, item)
}

func (s *StoreAdapter) ListUniverseItems(ctx context.Context, f UniverseFilterInput) ([]UniverseItemOutput, *string, error) {
	cursor, err := queries.ParseCursor(f.Cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListUniverseItems(ctx, queries.UniverseFilter{
		Active:     f.Active,
		EntityType: f.EntityType,
		Limit:      f.Limit,
		Cursor:     cursor,
	})
	if err != nil {
		return nil, nil, err
	}
	var out []UniverseItemOutput
	for _, it := range items {
		var keywords []string
		_ = json.Unmarshal(it.Keywords, &keywords)
		out = append(out, UniverseItemOutput{
			ID:         it.ID,
			EntityType: it.EntityType,
			EntityID:   it.EntityID,
			Name:       it.Name,
			Keywords:   keywords,
			Priority:   it.Priority,
			IsActive:   it.IsActive,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) UpdateUniverseItem(ctx context.Context, id string, u UniverseUpdateInput) error {
	return s.repo.UpdateUniverseItem(ctx, id, queries.UniverseUpdate{
		Priority: u.Priority,
		IsActive: u.IsActive,
		Keywords: u.Keywords,
	})
}

func (s *StoreAdapter) CreateRun(ctx context.Context, mode string, configJSON []byte) (string, error) {
	return s.repo.CreateRun(ctx, mode, configJSON)
}

func (s *StoreAdapter) GetRun(ctx context.Context, id string) (RunOutput, error) {
	run, err := s.repo.GetRun(ctx, id)
	if err != nil {
		return RunOutput{}, err
	}
	cfg := map[string]any{}
	_ = json.Unmarshal(run.ConfigJSON, &cfg)
	events, _, err := s.ListPhase1RunEventsByRunID(ctx, id, 200, "")
	if err != nil {
		return RunOutput{}, err
	}
	if events == nil {
		events = []Phase1RunEvent{}
	}
	return RunOutput{
		ID:         run.ID,
		Status:     run.Status,
		StartedAt:  run.StartedAt,
		FinishedAt: run.FinishedAt,
		Error:      run.Error,
		Config:     cfg,
		Events:     events,
	}, nil
}

func (s *StoreAdapter) ListEventsByRun(ctx context.Context, runID string, limit int, cursor string) ([]EventOutput, *string, error) {
	cur, err := queries.ParseCursor(cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListEventsByRun(ctx, runID, limit, cur)
	if err != nil {
		return nil, nil, err
	}
	var out []EventOutput
	for _, e := range items {
		var facts any
		var impact any
		var sources any
		_ = json.Unmarshal(e.FactsJSON, &facts)
		if len(e.ImpactJSON) > 0 {
			_ = json.Unmarshal(e.ImpactJSON, &impact)
		}
		_ = json.Unmarshal(e.Sources, &sources)
		out = append(out, EventOutput{
			EventID:    e.EventID,
			Category:   e.Category,
			ObservedAt: e.ObservedAt,
			Title:      e.Title,
			Facts:      facts,
			Impact:     impact,
			Sources:    sources,
			Confidence: e.Confidence,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) AppendEventToRun(ctx context.Context, runID string, input RunEventInput) (int, error) {
	source := input.Source
	if source == "" {
		source = "manual"
	}
	source = domain.NormalizePhase1EventSource(source)
	occurredAt := time.Now().UTC()
	if input.OccurredAt != nil {
		occurredAt = input.OccurredAt.UTC()
	}
	payload := []byte("{}")
	if input.Payload != nil {
		b, err := json.Marshal(input.Payload)
		if err != nil {
			return 0, err
		}
		payload = b
	}
	e := models.Phase1RunEvent{
		RunID:      runID,
		EventType:  input.EventType,
		Source:     source,
		OccurredAt: occurredAt,
		Payload:    payload,
	}
	return s.repo.CreatePhase1RunEvent(ctx, e)
}

func (s *StoreAdapter) ListPhase1RunEventsByRunID(ctx context.Context, runID string, limit int, cursor string) ([]Phase1RunEvent, *string, error) {
	cur, err := queries.ParseCursor(cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListPhase1RunEventsByRunID(ctx, runID, limit, cur)
	if err != nil {
		return nil, nil, err
	}
	out := []Phase1RunEvent{}
	for _, e := range items {
		payload := map[string]any{}
		if len(e.Payload) > 0 && string(e.Payload) != "null" {
			if err := json.Unmarshal(e.Payload, &payload); err != nil {
				payload = map[string]any{}
			}
		}
		out = append(out, Phase1RunEvent{
			RunID:      e.RunID,
			Seq:        e.Seq,
			EventType:  e.EventType,
			Source:     e.Source,
			OccurredAt: e.OccurredAt,
			Payload:    payload,
			CreatedAt:  e.CreatedAt,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) GetAnomalySummaryByRun(ctx context.Context, runID string) (AnomalySummaryOutput, error) {
	a, err := s.repo.GetAnomalySummaryByRun(ctx, runID)
	if err != nil {
		return AnomalySummaryOutput{}, err
	}
	var summary any
	_ = json.Unmarshal(a.Summary, &summary)
	return AnomalySummaryOutput{RunID: a.RunID, Summary: summary}, nil
}

func (s *StoreAdapter) GetTriggerDecisionByRun(ctx context.Context, runID string) (TriggerDecisionOutput, error) {
	t, err := s.repo.GetTriggerDecisionByRun(ctx, runID)
	if err != nil {
		return TriggerDecisionOutput{}, err
	}
	var decision any
	_ = json.Unmarshal(t.Decision, &decision)
	return TriggerDecisionOutput{RunID: t.RunID, Decision: decision}, nil
}

func (s *StoreAdapter) ListHandoffsByRun(ctx context.Context, runID string) ([]HandoffOutput, error) {
	items, err := s.repo.ListHandoffsByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	var out []HandoffOutput
	for _, h := range items {
		var packet map[string]any
		_ = json.Unmarshal(h.PacketJSON, &packet)
		out = append(out, HandoffOutput{
			ID:          h.ID,
			RunID:       h.RunID,
			CaseID:      h.CaseID,
			HandoffType: h.HandoffType,
			FromPhase:   h.FromPhase,
			ToPhase:     h.ToPhase,
			Packet:      packet,
			Status:      h.Status,
		})
	}
	return out, nil
}

func (s *StoreAdapter) CreateHandoff(ctx context.Context, input HandoffInput) (string, error) {
	packet, err := json.Marshal(input.Packet)
	if err != nil {
		return "", err
	}
	h := models.HandoffPacket{
		RunID:       input.RunID,
		CaseID:      input.CaseID,
		HandoffType: input.HandoffType,
		FromPhase:   input.FromPhase,
		ToPhase:     input.ToPhase,
		PacketJSON:  packet,
		Status:      "created",
	}
	return s.repo.CreateHandoff(ctx, h)
}

func (s *StoreAdapter) GetHandoff(ctx context.Context, id string) (HandoffOutput, error) {
	h, err := s.repo.GetHandoff(ctx, id)
	if err != nil {
		return HandoffOutput{}, err
	}
	var packet map[string]any
	_ = json.Unmarshal(h.PacketJSON, &packet)
	return HandoffOutput{
		ID:          h.ID,
		RunID:       h.RunID,
		CaseID:      h.CaseID,
		HandoffType: h.HandoffType,
		FromPhase:   h.FromPhase,
		ToPhase:     h.ToPhase,
		Packet:      packet,
		Status:      h.Status,
	}, nil
}

func (s *StoreAdapter) AttachCaseToHandoff(ctx context.Context, handoffID string, caseInput CaseInput) (string, error) {
	caseID, err := s.CreateCase(ctx, caseInput)
	if err != nil {
		return "", err
	}
	if err := s.repo.AttachCaseToHandoff(ctx, handoffID, caseID); err != nil {
		return "", err
	}
	return caseID, nil
}

func (s *StoreAdapter) CreateCase(ctx context.Context, input CaseInput) (string, error) {
	c := models.Case{
		CaseType: input.CaseType,
		EntityID: input.EntityID,
		Title:    input.Title,
		Priority: input.Priority,
	}
	return s.repo.CreateCase(ctx, c)
}

func (s *StoreAdapter) ListCases(ctx context.Context, f CaseFilterInput) ([]CaseOutput, *string, error) {
	cur, err := queries.ParseCursor(f.Cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListCases(ctx, queries.CaseFilter{
		Status: f.Status,
		Limit:  f.Limit,
		Cursor: cur,
	})
	if err != nil {
		return nil, nil, err
	}
	var out []CaseOutput
	for _, c := range items {
		out = append(out, CaseOutput{
			ID:       c.ID,
			Title:    c.Title,
			Status:   c.Status,
			CaseType: c.CaseType,
			EntityID: c.EntityID,
			Priority: c.Priority,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) GetCaseDetail(ctx context.Context, id string) (CaseDetailOutput, error) {
	d, err := s.repo.GetCaseDetail(ctx, id)
	if err != nil {
		return CaseDetailOutput{}, err
	}
	out := CaseDetailOutput{
		Case: CaseOutput{
			ID:       d.Case.ID,
			Title:    d.Case.Title,
			Status:   d.Case.Status,
			CaseType: d.Case.CaseType,
			EntityID: d.Case.EntityID,
			Priority: d.Case.Priority,
		},
		Handoffs:        []HandoffOutput{},
		Artifacts:       []ArtifactOutput{},
		Decisions:       []DecisionOutput{},
		MonitoringPlans: []MonitoringPlanOutput{},
	}
	for _, h := range d.Handoffs {
		var packet map[string]any
		_ = json.Unmarshal(h.PacketJSON, &packet)
		out.Handoffs = append(out.Handoffs, HandoffOutput{
			ID:          h.ID,
			RunID:       h.RunID,
			CaseID:      h.CaseID,
			HandoffType: h.HandoffType,
			FromPhase:   h.FromPhase,
			ToPhase:     h.ToPhase,
			Packet:      packet,
			Status:      h.Status,
		})
	}
	for _, a := range d.Artifacts {
		var content map[string]any
		if len(a.ContentJSON) > 0 {
			_ = json.Unmarshal(a.ContentJSON, &content)
		}
		out.Artifacts = append(out.Artifacts, ArtifactOutput{
			ID:           a.ID,
			Phase:        a.Phase,
			ArtifactType: a.Artifact,
			ContentMD:    a.ContentMD,
			ContentJSON:  content,
		})
	}
	for _, dec := range d.Decisions {
		var constraints []string
		var judge map[string]any
		_ = json.Unmarshal(dec.Constraints, &constraints)
		_ = json.Unmarshal(dec.JudgeResults, &judge)
		out.Decisions = append(out.Decisions, DecisionOutput{
			ID:           dec.ID,
			CaseID:       dec.CaseID,
			DecisionDate: dec.DecisionDate,
			OverallScore: dec.OverallScore,
			FinalLabel:   dec.FinalLabel,
			Constraints:  constraints,
			JudgeResults: judge,
			DecisionMD:   dec.DecisionMD,
		})
	}
	for _, m := range d.MonitoringPlans {
		var plan map[string]any
		_ = json.Unmarshal(m.PlanJSON, &plan)
		out.MonitoringPlans = append(out.MonitoringPlans, MonitoringPlanOutput{
			ID:         m.ID,
			CaseID:     m.CaseID,
			DecisionID: m.DecisionID,
			Status:     m.Status,
			Plan:       plan,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return out, nil
}

func (s *StoreAdapter) CreateArtifact(ctx context.Context, input ArtifactInput) (string, error) {
	var contentJSON []byte
	if input.ContentJSON != nil {
		contentJSON, _ = json.Marshal(input.ContentJSON)
	}
	a := models.PhaseArtifact{
		CaseID:      input.CaseID,
		Phase:       input.Phase,
		Artifact:    input.ArtifactType,
		ContentMD:   input.ContentMD,
		ContentJSON: contentJSON,
	}
	return s.repo.CreateArtifact(ctx, a)
}

func (s *StoreAdapter) ListArtifacts(ctx context.Context, caseID string, f ArtifactFilterInput) ([]ArtifactOutput, error) {
	items, err := s.repo.ListArtifacts(ctx, caseID, queries.ArtifactFilter{
		Phase:  f.Phase,
		Latest: f.Latest,
	})
	if err != nil {
		return nil, err
	}
	var out []ArtifactOutput
	for _, a := range items {
		var content map[string]any
		if len(a.ContentJSON) > 0 {
			_ = json.Unmarshal(a.ContentJSON, &content)
		}
		out = append(out, ArtifactOutput{
			ID:           a.ID,
			Phase:        a.Phase,
			ArtifactType: a.Artifact,
			ContentMD:    a.ContentMD,
			ContentJSON:  content,
		})
	}
	return out, nil
}

func (s *StoreAdapter) CreateDecision(ctx context.Context, input DecisionInput) (string, error) {
	constraints, _ := json.Marshal(input.Constraints)
	judge, _ := json.Marshal(input.JudgeResults)
	d := models.Decision{
		CaseID:       input.CaseID,
		OverallScore: input.OverallScore,
		FinalLabel:   input.FinalLabel,
		Constraints:  constraints,
		JudgeResults: judge,
		DecisionMD:   input.DecisionMD,
	}
	return s.repo.CreateDecision(ctx, d)
}

func (s *StoreAdapter) ListDecisionsByCase(ctx context.Context, caseID string) ([]DecisionOutput, error) {
	items, err := s.repo.ListDecisionsByCase(ctx, caseID)
	if err != nil {
		return nil, err
	}
	var out []DecisionOutput
	for _, d := range items {
		var constraints []string
		var judge map[string]any
		_ = json.Unmarshal(d.Constraints, &constraints)
		_ = json.Unmarshal(d.JudgeResults, &judge)
		out = append(out, DecisionOutput{
			ID:           d.ID,
			CaseID:       d.CaseID,
			DecisionDate: d.DecisionDate,
			OverallScore: d.OverallScore,
			FinalLabel:   d.FinalLabel,
			Constraints:  constraints,
			JudgeResults: judge,
			DecisionMD:   d.DecisionMD,
		})
	}
	return out, nil
}

func (s *StoreAdapter) GetDecision(ctx context.Context, id string) (DecisionOutput, error) {
	d, err := s.repo.GetDecision(ctx, id)
	if err != nil {
		return DecisionOutput{}, err
	}
	var constraints []string
	var judge map[string]any
	_ = json.Unmarshal(d.Constraints, &constraints)
	_ = json.Unmarshal(d.JudgeResults, &judge)
	return DecisionOutput{
		ID:           d.ID,
		CaseID:       d.CaseID,
		DecisionDate: d.DecisionDate,
		OverallScore: d.OverallScore,
		FinalLabel:   d.FinalLabel,
		Constraints:  constraints,
		JudgeResults: judge,
		DecisionMD:   d.DecisionMD,
	}, nil
}

func (s *StoreAdapter) CreateMonitoringPlan(ctx context.Context, input MonitoringPlanInput) (string, error) {
	plan, _ := json.Marshal(input.Plan)
	m := models.MonitoringPlan{
		CaseID:     input.CaseID,
		DecisionID: input.DecisionID,
		Status:     "active",
		PlanJSON:   plan,
	}
	return s.repo.CreateMonitoringPlan(ctx, m)
}

func (s *StoreAdapter) GetMonitoringPlan(ctx context.Context, id string) (MonitoringPlanOutput, error) {
	m, err := s.repo.GetMonitoringPlan(ctx, id)
	if err != nil {
		return MonitoringPlanOutput{}, err
	}
	var plan map[string]any
	_ = json.Unmarshal(m.PlanJSON, &plan)
	return MonitoringPlanOutput{
		ID:         m.ID,
		CaseID:     m.CaseID,
		DecisionID: m.DecisionID,
		Status:     m.Status,
		Plan:       plan,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

func (s *StoreAdapter) ListMonitoringPlansByCase(ctx context.Context, caseID string, limit int, cursor string) ([]MonitoringPlanOutput, *string, error) {
	cur, err := queries.ParseCursor(cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListMonitoringPlansByCase(ctx, caseID, limit, cur)
	if err != nil {
		return nil, nil, err
	}
	var out []MonitoringPlanOutput
	for _, m := range items {
		var plan map[string]any
		_ = json.Unmarshal(m.PlanJSON, &plan)
		out = append(out, MonitoringPlanOutput{
			ID:         m.ID,
			CaseID:     m.CaseID,
			DecisionID: m.DecisionID,
			Status:     m.Status,
			Plan:       plan,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) CreateAlert(ctx context.Context, input AlertInput) (string, error) {
	refs, _ := json.Marshal(input.Refs)
	a := models.Alert{
		MonitoringPlanID: input.MonitoringPlanID,
		Severity:         input.Severity,
		Type:             input.Type,
		Message:          input.Message,
		RefsJSON:         refs,
	}
	return s.repo.CreateAlert(ctx, a)
}

func (s *StoreAdapter) ListAlertsByPlan(ctx context.Context, planID string, limit int, cursor string) ([]AlertOutput, *string, error) {
	cur, err := queries.ParseCursor(cursor)
	if err != nil {
		return nil, nil, err
	}
	items, next, err := s.repo.ListAlertsByPlan(ctx, planID, limit, cur)
	if err != nil {
		return nil, nil, err
	}
	var out []AlertOutput
	for _, a := range items {
		var refs map[string]any
		_ = json.Unmarshal(a.RefsJSON, &refs)
		out = append(out, AlertOutput{
			ID:        a.ID,
			Severity:  a.Severity,
			Type:      a.Type,
			Message:   a.Message,
			Refs:      refs,
			CreatedAt: a.CreatedAt,
			AckAt:     a.AcknowledgedAt,
		})
	}
	var nextCursor *string
	if next != nil {
		s := next.Format(time.RFC3339Nano)
		nextCursor = &s
	}
	return out, nextCursor, nil
}

func (s *StoreAdapter) AckAlert(ctx context.Context, id string) error {
	return s.repo.AckAlert(ctx, id)
}
