package idempotency

import (
	"context"
	"log"
	"net/http"
)

type httpWriter struct {
	http.ResponseWriter
	Instance

	idemKey string
	ctx     context.Context
}

func (w *httpWriter) WriteHeader(status int) {
	if status >= http.StatusBadRequest {
		if err := w.Instance.DeleteIdempotencyKey(w.ctx, w.idemKey); err != nil {
			log.Print("Couldn't delete the idempotent Key")
		}
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *httpWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func HTTPMiddleware(state Instance, headerName string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			idemKey := r.Header.Get(headerName)
			if len(idemKey) == 0 {
				next(w, r)
				return
			}

			dw := &httpWriter{ResponseWriter: w, Instance: state, idemKey: idemKey, ctx: ctx}
			err := state.CheckAndSet(ctx, idemKey)

			if err == ErrKeyAlreadyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}

			next(dw, r)
		}
	}
}
