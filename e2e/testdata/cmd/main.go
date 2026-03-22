package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"urlserver/service"
)

func parseUserID(r *http.Request) (int, error) {
	raw := r.Header.Get("x-user-id")
	if raw == "" {
		return 0, errors.New("missing x-user-id header")
	}
	return strconv.Atoi(raw)
}

func encodeHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input := r.URL.Query().Get("input")
		if input == "" {
			http.Error(w, "missing input parameter", http.StatusBadRequest)
			return
		}
		encoded, err := svc.Encode(userID, input)
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		fmt.Fprint(w, encoded)
	}
}

func decodeHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		input := r.URL.Query().Get("input")
		if input == "" {
			http.Error(w, "missing input parameter", http.StatusBadRequest)
			return
		}
		decoded, err := svc.Decode(userID, input)
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, decoded)
	}
}

//nocover:block
func main() {
	svc := service.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/encode", encodeHandler(svc))
	mux.HandleFunc("/decode", decodeHandler(svc))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Println("service starting at port: 8080")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "listen error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
		os.Exit(1)
	}
}
