package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"urlserver/service"
)

func newTestServer() *http.ServeMux {
	svc := service.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/encode", encodeHandler(svc))
	mux.HandleFunc("/decode", decodeHandler(svc))
	return mux
}

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    int
		wantErr bool
	}{
		{"valid id", "1", 1, false},
		{"zero", "0", 0, false},
		{"missing header", "", 0, true},
		{"non-numeric", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("x-user-id", tt.header)
			}
			got, err := parseUserID(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseUserID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEncodeHandler(t *testing.T) {
	mux := newTestServer()

	tests := []struct {
		name       string
		userID     string
		input      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "encodes input",
			userID:     "1",
			input:      "hello world",
			wantStatus: http.StatusOK,
			wantBody:   "hello+world",
		},
		{
			name:       "missing input returns 400",
			userID:     "1",
			input:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing user-id returns 400",
			userID:     "",
			input:      "hello",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid user-id returns 400",
			userID:     "abc",
			input:      "hello",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/encode"
			if tt.input != "" {
				target += "?input=" + url.QueryEscape(tt.input)
			}
			r := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.userID != "" {
				r.Header.Set("x-user-id", tt.userID)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestDecodeHandler(t *testing.T) {
	mux := newTestServer()

	tests := []struct {
		name       string
		userID     string
		input      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "decodes input",
			userID:     "1",
			input:      "hello+world",
			wantStatus: http.StatusOK,
			wantBody:   "hello world",
		},
		{
			name:       "missing input returns 400",
			userID:     "1",
			input:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing user-id returns 400",
			userID:     "",
			input:      "hello",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid escape returns 400",
			userID:     "1",
			input:      "hello%ZZ",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/decode"
			if tt.input != "" {
				target += "?input=" + tt.input
			}
			r := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.userID != "" {
				r.Header.Set("x-user-id", tt.userID)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}
