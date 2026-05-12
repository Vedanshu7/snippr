package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	dbpkg "github.com/vedanshu/snippr/internal/db"
)

// BoardEvent is broadcast to all SSE clients watching a workspace.
type BoardEvent struct {
	Type    string          `json:"type"`    // "snippet_added" | "snippet_updated" | "snippet_deleted"
	Payload json.RawMessage `json:"payload"` // full Snippet JSON, or {"id":N} for deleted
}

// Hub manages SSE client channels keyed by workspace ID.
// TODO(*): Replace Hub with Redis pub/sub for multi-instance deployments.
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[chan BoardEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[chan BoardEvent]struct{})}
}

func (h *Hub) subscribe(workspaceID int64) chan BoardEvent {
	ch := make(chan BoardEvent, 32)
	h.mu.Lock()
	if h.clients[workspaceID] == nil {
		h.clients[workspaceID] = make(map[chan BoardEvent]struct{})
	}
	h.clients[workspaceID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) unsubscribe(workspaceID int64, ch chan BoardEvent) {
	h.mu.Lock()
	delete(h.clients[workspaceID], ch)
	if len(h.clients[workspaceID]) == 0 {
		delete(h.clients, workspaceID)
	}
	h.mu.Unlock()
	close(ch)
}

// Broadcast sends an event to all clients watching workspaceID.
// Sends are non-blocking — slow clients miss events rather than blocking the broadcaster.
func (h *Hub) Broadcast(workspaceID int64, event BoardEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[workspaceID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// MakeBoardEvent marshals v into a BoardEvent payload.
func MakeBoardEvent(eventType string, v any) (BoardEvent, error) {
	p, err := json.Marshal(v)
	if err != nil {
		return BoardEvent{}, err
	}
	return BoardEvent{Type: eventType, Payload: json.RawMessage(p)}, nil
}

// MakeDeleteEvent creates a BoardEvent for a deleted snippet.
func MakeDeleteEvent(id int64) BoardEvent {
	return BoardEvent{
		Type:    "snippet_deleted",
		Payload: json.RawMessage(fmt.Sprintf(`{"id":%d}`, id)),
	}
}

// EditingPayload is the SSE payload for editing_start / editing_stop events.
type EditingPayload struct {
	SnippetID int64  `json:"snippet_id"`
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Editing   bool   `json:"editing"`
}

type announceEditingReq struct {
	SnippetID int64 `json:"snippet_id"`
	Editing   bool  `json:"editing"`
}

// AnnounceEditingHandler broadcasts editing presence events to workspace board watchers.
func AnnounceEditingHandler(db *sql.DB, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid workspace id")
			return
		}

		userID := GetUserID(r)
		ok, err := dbpkg.IsMember(db, workspaceID, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "not a workspace member")
			return
		}

		var req announceEditingReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.SnippetID <= 0 {
			writeError(w, http.StatusBadRequest, "snippet_id required")
			return
		}

		email, err := dbpkg.GetUserEmail(db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		eventType := "editing_stop"
		if req.Editing {
			eventType = "editing_start"
		}
		event, err := MakeBoardEvent(eventType, EditingPayload{
			SnippetID: req.SnippetID,
			UserID:    userID,
			Email:     email,
			Editing:   req.Editing,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		hub.Broadcast(workspaceID, event)
		w.WriteHeader(http.StatusNoContent)
	}
}

func BoardSSEHandler(db *sql.DB, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid workspace id")
			return
		}

		userID := GetUserID(r)
		ok, err := dbpkg.IsMember(db, workspaceID, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "not a workspace member")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "streaming unsupported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		// Initial heartbeat flushes headers to the browser immediately.
		fmt.Fprint(w, ": ok\n\n")
		flusher.Flush()

		ch := hub.subscribe(workspaceID)
		defer hub.unsubscribe(workspaceID, ch)

		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, open := <-ch:
				if !open {
					return
				}
				data, _ := json.Marshal(event)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}
