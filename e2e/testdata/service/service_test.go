package service

import (
	"errors"
	"testing"
)

func TestURLService_Encode(t *testing.T) {
	svc := New()

	tests := []struct {
		name    string
		userID  int
		input   string
		want    string
		wantErr error
	}{
		{
			name:   "encodes spaces and special chars",
			userID: 1,
			input:  "hello world&foo=bar",
			want:   "hello+world%26foo%3Dbar",
		},
		{
			name:   "plain string unchanged",
			userID: 1,
			input:  "hello",
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Encode(tt.userID, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestURLService_Decode(t *testing.T) {
	svc := New()

	tests := []struct {
		name    string
		userID  int
		input   string
		want    string
		wantErr error
	}{
		{
			name:   "decodes encoded string",
			userID: 1,
			input:  "hello+world%26foo%3Dbar",
			want:   "hello world&foo=bar",
		},
		{
			name:   "plain string unchanged",
			userID: 1,
			input:  "hello",
			want:   "hello",
		},
		{
			name:    "invalid escape sequence returns error",
			userID:  1,
			input:   "hello%ZZ",
			wantErr: errors.New("decode error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Decode(tt.userID, tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Decode() expected error, got nil")
					return
				}
				if tt.wantErr == ErrForbidden && !errors.Is(err, ErrForbidden) {
					t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Decode() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() = %q, want %q", got, tt.want)
			}
		})
	}
}
