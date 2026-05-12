package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

func authHeader(t *testing.T, ja *jwtauth.JWTAuth, userID int64) string {
	t.Helper()
	claims := map[string]any{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	_, tok, err := ja.Encode(claims)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func authedPost(t *testing.T, r http.Handler, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func authedGet(t *testing.T, r http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func authedDelete(t *testing.T, r http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func registerAndGetID(t *testing.T, r http.Handler, email string) int64 {
	t.Helper()
	w := post(t, r, "/api/register", fmt.Sprintf(`{"email":%q,"password":"password123"}`, email))
	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	return int64(resp["id"].(float64))
}

func TestCreateAndGetSnippet(t *testing.T) {
	database := setupDB(t)
	r, ja := setupRouter(t, database)

	userID := registerAndGetID(t, r, "s1@example.com")
	tok := authHeader(t, ja, userID)

	w := authedPost(t, r, "/api/snippets",
		`{"title":"Hello","content":"fmt.Println()","language":"go","is_public":false}`, tok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created map[string]any
	json.NewDecoder(w.Body).Decode(&created)
	id := int64(created["id"].(float64))

	w2 := authedGet(t, r, fmt.Sprintf("/api/snippets/%d", id), tok)
	if w2.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w2.Code)
	}
}

func TestListSnippets(t *testing.T) {
	database := setupDB(t)
	r, ja := setupRouter(t, database)

	userID := registerAndGetID(t, r, "s2@example.com")
	tok := authHeader(t, ja, userID)

	authedPost(t, r, "/api/snippets", `{"title":"First","content":"a","language":"text"}`, tok)
	authedPost(t, r, "/api/snippets", `{"title":"Second","content":"b","language":"text"}`, tok)

	w := authedGet(t, r, "/api/snippets", tok)
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	var snippets []map[string]any
	json.NewDecoder(w.Body).Decode(&snippets)
	if len(snippets) != 2 {
		t.Errorf("expected 2 snippets, got %d", len(snippets))
	}
}

func TestPublicSnippet(t *testing.T) {
	database := setupDB(t)
	r, ja := setupRouter(t, database)

	userID := registerAndGetID(t, r, "s3@example.com")
	tok := authHeader(t, ja, userID)

	w := authedPost(t, r, "/api/snippets",
		`{"title":"Public","content":"hello","language":"text","is_public":true}`, tok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create public: %d %s", w.Code, w.Body.String())
	}

	var created map[string]any
	json.NewDecoder(w.Body).Decode(&created)
	slug := created["share_slug"].(string)
	if slug == "" {
		t.Fatal("expected share_slug for public snippet")
	}

	w2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/s/"+slug, nil)
	r.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Fatalf("public slug: expected 200, got %d", w2.Code)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	database := setupDB(t)
	r, _ := setupRouter(t, database)

	req := httptest.NewRequest(http.MethodGet, "/api/snippets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
}

func TestDeleteSnippet(t *testing.T) {
	database := setupDB(t)
	r, ja := setupRouter(t, database)

	userID := registerAndGetID(t, r, "s4@example.com")
	tok := authHeader(t, ja, userID)

	w := authedPost(t, r, "/api/snippets", `{"title":"Delete me","content":"x","language":"text"}`, tok)
	var created map[string]any
	json.NewDecoder(w.Body).Decode(&created)
	id := int64(created["id"].(float64))

	wd := authedDelete(t, r, fmt.Sprintf("/api/snippets/%d", id), tok)
	if wd.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", wd.Code)
	}

	wg := authedGet(t, r, fmt.Sprintf("/api/snippets/%d", id), tok)
	if wg.Code != http.StatusNotFound {
		t.Fatalf("after delete get: expected 404, got %d", wg.Code)
	}
}
