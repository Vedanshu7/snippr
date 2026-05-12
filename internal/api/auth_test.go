package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func post(t *testing.T, r http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegister(t *testing.T) {
	database := setupDB(t)
	r, _ := setupRouter(t, database)

	w := post(t, r, "/api/register", `{"email":"test@example.com","password":"password123"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["email"] != "test@example.com" {
		t.Errorf("expected email in response, got %v", resp)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	database := setupDB(t)
	r, _ := setupRouter(t, database)

	body := `{"email":"dup@example.com","password":"password123"}`
	w1 := post(t, r, "/api/register", body)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", w1.Code)
	}

	w2 := post(t, r, "/api/register", body)
	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestLogin(t *testing.T) {
	database := setupDB(t)
	r, _ := setupRouter(t, database)

	post(t, r, "/api/register", `{"email":"login@example.com","password":"password123"}`)

	w := post(t, r, "/api/login", `{"email":"login@example.com","password":"password123"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("expected token in response")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	database := setupDB(t)
	r, _ := setupRouter(t, database)

	post(t, r, "/api/register", `{"email":"wp@example.com","password":"password123"}`)

	w := post(t, r, "/api/login", `{"email":"wp@example.com","password":"wrongpassword"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
