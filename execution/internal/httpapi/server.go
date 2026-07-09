package httpapi

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"trading-bot/execution/internal/engine"
)

type Server struct {
	engine *engine.Engine
}

//go:embed dashboard.html
var dashboardHTML []byte

func NewServer(eng *engine.Engine) *Server {
	return &Server{engine: eng}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/positions", s.handlePositions)
	mux.HandleFunc("/pause", s.handlePause)
	mux.HandleFunc("/resume", s.handleResume)
	mux.HandleFunc("/kill", s.handleKill)
	return mux
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	dailyPnL, trades, wins := s.engine.GetDailyStats()
	resp := map[string]any{
		"running":    s.engine.IsRunning(),
		"breached":   s.engine.IsBreached(),
		"equity":     s.engine.GetEquity(),
		"daily_pnl":  dailyPnL,
		"trades":     trades,
		"wins":       wins,
		"positions":  s.engine.GetPositions(),
		"signals":    len(s.engine.GetActiveSignals()),
		"strategies": s.engine.GetStrategyNames(),
	}
	writeJSON(w, resp)
}

func (s *Server) handlePositions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.engine.GetPositions())
}

func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.engine.PauseAll()
	writeJSON(w, map[string]string{"status": "paused"})
}

func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.engine.ResumeAll()
	writeJSON(w, map[string]string{"status": "resumed"})
}

func (s *Server) handleKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.engine.KillAll()
	writeJSON(w, map[string]string{"status": "killed"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
