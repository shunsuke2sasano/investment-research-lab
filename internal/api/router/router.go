package router

import (
	"net/http"
	"strings"

	"investment_committee/internal/api/handlers"
	"investment_committee/internal/db/queries"
)

type Router struct {
	server *handlers.Server
	apiKey string
}

func New(repo *queries.Repository, apiKey string) *Router {
	return &Router{
		server: handlers.NewServer(handlers.NewStoreAdapter(repo)),
		apiKey: apiKey,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.apiKey != "" && !handlers.AuthOK(req, r.apiKey) {
		handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "api" || parts[1] != "v1" {
		handlers.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	parts = parts[2:]

	if len(parts) == 2 && parts[0] == "universe" && parts[1] == "items" {
		r.server.HandleUniverseItems(w, req)
		return
	}
	if len(parts) == 3 && parts[0] == "universe" && parts[1] == "items" {
		r.server.HandleUniverseItem(w, req, parts[2])
		return
	}

	if len(parts) >= 2 && parts[0] == "phase1" && parts[1] == "runs" {
		r.server.HandlePhase1Runs(w, req, parts[2:])
		return
	}

	if len(parts) == 1 && parts[0] == "handoffs" {
		r.server.HandleHandoffs(w, req)
		return
	}
	if len(parts) >= 2 && parts[0] == "handoffs" {
		r.server.HandleHandoff(w, req, parts[1:])
		return
	}

	if len(parts) == 1 && parts[0] == "cases" {
		r.server.HandleCases(w, req)
		return
	}
	if len(parts) >= 2 && parts[0] == "cases" {
		r.server.HandleCase(w, req, parts[1:])
		return
	}

	if len(parts) >= 2 && parts[0] == "decisions" {
		r.server.HandleDecision(w, req, parts[1:])
		return
	}

	if len(parts) >= 2 && parts[0] == "monitoring-plans" {
		r.server.HandleMonitoringPlan(w, req, parts[1:])
		return
	}

	if len(parts) >= 2 && parts[0] == "alerts" {
		r.server.HandleAlert(w, req, parts[1:])
		return
	}

	handlers.WriteError(w, http.StatusNotFound, "not found")
}
