package httpsproxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHandlers_ServeHTTP(t *testing.T) {
	tests := []struct {
		name string
		h    Handlers

		wantStatus int
		wantBody   string
	}{
		{
			name: "empty",
			h:    Handlers{},

			wantStatus: http.StatusExpectationFailed,
		},
		{
			name: "one handler, success",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hello"))
				}),
			},

			wantBody: "hello",
		},
		{
			name: "one handler, failure",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}),
			},

			wantStatus: http.StatusTeapot,
		},
		{
			name: "two handlers, get fastest success",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hello"))
				}),
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.Write([]byte("world"))
				}),
			},

			wantBody: "hello",
		},
		{
			name: "two handlers, first failure",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}),
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.Write([]byte("world"))
				}),
			},

			wantBody: "world",
		},
		{
			name: "two handlers, both failure, get fastest failure",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}),
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.WriteHeader(http.StatusBadRequest)
				}),
			},

			wantStatus: http.StatusTeapot,
		},
		{
			name: "two handlers, ignore failed with timeout, return slow success",
			h: Handlers{
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusRequestTimeout)
				}),
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.Write([]byte("world"))
				}),
			},

			wantBody: "world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bufferedResponseWriter
			var r http.Request

			tt.h.ServeHTTP(&b, &r)

			if b.statusCode != tt.wantStatus {
				t.Errorf("statusCode = %v, want %v", b.statusCode, tt.wantStatus)
			}
			if diff := cmp.Diff(b.response.String(), tt.wantBody); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBufferedResponseWriter(t *testing.T) {
	var b bufferedResponseWriter

	header := b.Header()
	if header == nil {
		t.Errorf("header is nil")
	}

	b.WriteHeader(http.StatusTeapot)
	if b.statusCode != http.StatusTeapot {
		t.Errorf("statusCode = %v, want %v", b.statusCode, http.StatusTeapot)
	}

	b.Write([]byte("hello"))
	if b.response.String() != "hello" {
		t.Errorf("response = %v, want %v", b.response.String(), "hello")
	}
}
