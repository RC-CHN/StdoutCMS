package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"stdoutcms/internal/config"
	"stdoutcms/internal/store"
)

type loginAdminStub struct {
	admin *store.Admin
}

func (s loginAdminStub) GetAdminByUsername(_ context.Context, username string) (*store.Admin, error) {
	if username != s.admin.Username {
		return nil, store.ErrNotFound
	}
	return s.admin, nil
}

type loginSessionStub struct {
	err      error
	calls    int
	ctx      context.Context
	token    string
	username string
	ttl      int
}

func (s *loginSessionStub) SetSession(ctx context.Context, token, username string, ttl int) error {
	s.calls++
	s.ctx, s.token, s.username, s.ttl = ctx, token, username, ttl
	return s.err
}

func TestLoginSessionCreation(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	pg := loginAdminStub{admin: &store.Admin{Username: "root", PasswordHash: string(hash)}}
	cfg := &config.Config{SessionMaxAge: 3600}

	for _, tc := range []struct {
		name       string
		body       string
		sessionErr error
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "session stored",
			body:       `{"username":"root","password":"correct-password"}`,
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
		{
			name:       "session write fails",
			body:       `{"username":"root","password":"correct-password"}`,
			sessionErr: errors.New("redis write failed"),
			wantStatus: http.StatusServiceUnavailable,
			wantCalls:  1,
		},
		{
			name:       "wrong password",
			body:       `{"username":"root","password":"wrong-password"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unknown user",
			body:       `{"username":"unknown","password":"correct-password"}`,
			wantStatus: http.StatusUnauthorized,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rd := &loginSessionStub{err: tc.sessionErr}
			r := gin.New()
			r.POST("/login", Login(pg, rd, cfg))

			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-Proto", "https")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if rd.calls != tc.wantCalls {
				t.Fatalf("session writes = %d, want %d", rd.calls, tc.wantCalls)
			}
			if rd.calls > 0 && (rd.ctx != req.Context() || rd.username != "root" || rd.ttl != cfg.SessionMaxAge) {
				t.Fatal("session must use the request context, authenticated username and configured TTL")
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.wantStatus != http.StatusOK {
				if _, ok := body["token"]; ok {
					t.Error("failed login returned a token")
				}
				if len(w.Header().Values("Set-Cookie")) != 0 {
					t.Error("failed login set a cookie")
				}
				if tc.sessionErr != nil {
					if body["error"] != "authentication service unavailable" || body["code"] != "AUTH_BACKEND_DOWN" {
						t.Errorf("unexpected session failure response: %v", body)
					}
				}
				return
			}

			if len(body["token"]) != 64 || body["token"] != rd.token {
				t.Error("response token must match the stored session token")
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("cookies = %d, want 1", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Name != "blog_session_token" || cookie.Value != rd.token ||
				!cookie.HttpOnly || !cookie.Secure || cookie.Path != "/" || cookie.MaxAge != cfg.SessionMaxAge {
				t.Error("cookie must carry the stored token and preserve session attributes")
			}
		})
	}
}
