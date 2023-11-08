package http

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
)

type ErrorAwareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type ErrorAwareHandlerFunc func(http.ResponseWriter, *http.Request) error

func (e ErrorAwareHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return e(w, r)
}

// txProvider is an interface that our container implements.
type txProvider interface {
	Tx(context.Context) *sql.Tx
}

// NewAutoCloseTxEndpoint wraps the provided handler with another one that automatically commits/rollbacks transactions.
func NewAutoCloseTxEndpoint(t txProvider, handler ErrorAwareHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := t.Tx(r.Context())
		ok := false

		defer func() {
			if !ok {
				// handler panics, so we have to rollback the transaction
				_ = tx.Rollback()
			}
		}()

		s := "Decorator AutoCloseTxEndpoint:\n"
		s += fmt.Sprintf("\tTxID: %p\n", tx)
		_, _ = w.Write([]byte(s))

		err := handler.ServeHTTP(w, r)
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
		ok = true
	})
}
