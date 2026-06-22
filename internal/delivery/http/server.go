package http

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"cli-assistant/internal/domain/history"
	"cli-assistant/internal/usecase"
	"cli-assistant/pkg/version"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	Addr    string
	Token   string
	Profile string
}

type Server struct {
	dep     *usecase.Deployment
	obs     *usecase.Observability
	inspect *usecase.Inspector
	hist    history.Store
	token   string
	addr    string
	profile string
}

func New(
	dep *usecase.Deployment,
	obs *usecase.Observability,
	inspect *usecase.Inspector,
	hist history.Store,
	opts Options,
) *Server {
	addr := opts.Addr
	if addr == "" {
		addr = ":8080"
	}
	return &Server{
		dep:     dep,
		obs:     obs,
		inspect: inspect,
		hist:    hist,
		token:   opts.Token,
		addr:    addr,
		profile: opts.Profile,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	var shutdownOnce sync.Once
	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		shutdownOnce.Do(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx)
			close(done)
		})
	}()

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	<-done
	return nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /version", s.version)
	mux.HandleFunc("GET /v1/deploy/applications", s.auth(s.listApps))
	mux.HandleFunc("GET /v1/deploy/applications/{name}", s.auth(s.getApp))
	mux.HandleFunc("GET /v1/observe/health", s.auth(s.observeHealth))
	mux.HandleFunc("GET /v1/observe/query", s.auth(s.query))
	mux.HandleFunc("GET /v1/deploy/applications/{name}/inspect", s.auth(s.inspectApp))
	mux.HandleFunc("GET /v1/history/commands", s.auth(s.commandHistory))
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
		"go_version": runtime.Version(),
	})
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next(w, r)
			return
		}
		got := r.Header.Get("Authorization")
		if got != "Bearer "+s.token {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func (s *Server) listApps(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	res, err := s.dep.List(r.Context(), deploy.ListFilter{
		Namespace:  q.Get("namespace"),
		NamePrefix: q.Get("name_prefix"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res.Applications)
}

func (s *Server) getApp(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	res, err := s.dep.Status(r.Context(), name)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res.Application)
}

func (s *Server) observeHealth(w http.ResponseWriter, r *http.Request) {
	res, err := s.obs.Health(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res.Health)
}

func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	expr := r.URL.Query().Get("expr")
	if expr == "" {
		writeError(w, http.StatusBadRequest, "expr query parameter is required")
		return
	}
	res, err := s.obs.Query(r.Context(), expr)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res.Result)
}

func (s *Server) inspectApp(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	res, err := s.inspect.Inspect(r.Context(), name)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	if s.hist != nil {
		payload, mErr := json.Marshal(res)
		if mErr == nil {
			_ = s.hist.SaveInspectSnapshot(r.Context(), history.InspectSnapshot{
				AppName:     name,
				Profile:     s.profile,
				PayloadJSON: payload,
			})
		}
	}

	writeJSON(w, http.StatusOK, res)
}

func (s *Server) commandHistory(w http.ResponseWriter, r *http.Request) {
	if s.hist == nil {
		writeError(w, http.StatusServiceUnavailable, "history store is not configured")
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}

	runs, err := s.hist.ListCommandRuns(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUnavailable):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
