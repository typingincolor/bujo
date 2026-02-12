package http

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
	"github.com/typingincolor/bujo/internal/service"
)

//go:embed install.html
var installPage []byte

type Handler struct {
	bujo *service.BujoService
}

func NewHandler(bujo *service.BujoService) *Handler {
	return &Handler{bujo: bujo}
}

var allowedOrigins = map[string]bool{
	"https://mail.google.com": true,
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.handleHealth)
	mux.HandleFunc("POST /api/entries", h.handleCreateEntries)
	mux.HandleFunc("GET /install", h.handleInstall)
	return corsMiddleware(mux)
}

func (h *Handler) handleInstall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(installPage)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type entryInput struct {
	Type     string       `json:"type"`
	Content  string       `json:"content"`
	Children []entryInput `json:"children,omitempty"`
}

type createEntriesRequest struct {
	Entries []entryInput `json:"entries"`
}

type entryResult struct {
	ID       int64         `json:"id"`
	Children []entryResult `json:"children,omitempty"`
}

type createEntriesResponse struct {
	Success bool          `json:"success"`
	Entries []entryResult `json:"entries,omitempty"`
	Error   string        `json:"error,omitempty"`
}

func (h *Handler) handleCreateEntries(w http.ResponseWriter, r *http.Request) {
	var req createEntriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if len(req.Entries) == 0 {
		writeError(w, http.StatusBadRequest, "No entries provided")
		return
	}

	input, childCounts, err := buildInput(req.Entries)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	today := time.Now()
	ids, err := h.bujo.LogEntries(r.Context(), input, service.LogEntriesOptions{
		Date: today,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create entries")
		return
	}

	results := buildResults(ids, childCounts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createEntriesResponse{
		Success: true,
		Entries: results,
	})
}

func buildInput(entries []entryInput) (string, []int, error) {
	var lines []string
	var childCounts []int

	for _, e := range entries {
		if e.Content == "" {
			return "", nil, fmt.Errorf("Missing required field: content")
		}

		entryType, err := domain.ParseEntryTypeFromString(e.Type)
		if err != nil {
			return "", nil, fmt.Errorf("Invalid entry type: %s", e.Type)
		}

		symbol := symbolForType(entryType)
		lines = append(lines, symbol+" "+sanitizeContent(e.Content))
		childCounts = append(childCounts, len(e.Children))

		for _, child := range e.Children {
			if child.Content == "" {
				return "", nil, fmt.Errorf("Missing required field: content")
			}

			childType, err := domain.ParseEntryTypeFromString(child.Type)
			if err != nil {
				return "", nil, fmt.Errorf("Invalid entry type: %s", child.Type)
			}

			childSymbol := symbolForType(childType)
			lines = append(lines, "  "+childSymbol+" "+sanitizeContent(child.Content))
		}
	}

	return strings.Join(lines, "\n"), childCounts, nil
}

func buildResults(ids []int64, childCounts []int) []entryResult {
	var results []entryResult
	idx := 0

	for _, count := range childCounts {
		if idx >= len(ids) {
			break
		}

		result := entryResult{ID: ids[idx]}
		idx++

		for j := 0; j < count && idx < len(ids); j++ {
			result.Children = append(result.Children, entryResult{ID: ids[idx]})
			idx++
		}

		results = append(results, result)
	}

	return results
}

func sanitizeContent(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func symbolForType(et domain.EntryType) string {
	switch et {
	case domain.EntryTypeTask:
		return "."
	case domain.EntryTypeNote:
		return "-"
	case domain.EntryTypeEvent:
		return "o"
	case domain.EntryTypeDone:
		return "x"
	case domain.EntryTypeMigrated:
		return ">"
	case domain.EntryTypeQuestion:
		return "?"
	case domain.EntryTypeAnswer:
		return "A"
	default:
		return "."
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(createEntriesResponse{
		Success: false,
		Error:   message,
	})
}
