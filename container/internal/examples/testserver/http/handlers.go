// Copyright (c) 2023–present Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package http

import (
	"database/sql"
	"fmt"
	"net/http"
)

// ErrorAwareHandler in contrast to [http.Handler] may return an error.
// It lets us create endpoints that may return errors.
// Such an endpoint may by simply wrapped by another one that automatically handles the transaction
// (see example [NewAutoCloseTxEndpoint]) and implements [http.Handler] from stdlib.
type ErrorAwareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type ErrorAwareHandlerFunc func(http.ResponseWriter, *http.Request) error

func (e ErrorAwareHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return e(w, r)
}

// NewAutoCloseTxEndpoint wraps the provided handler with another one that automatically commits/rollbacks transactions.
// NOTE:
// This function requires *sql.Tx and [ErrorAwareHandler].
// Since *sql.Tx has the contextual scope in our container,
// the same instance of it will be used in this function and in [ErrorAwareHandler] :)
func NewAutoCloseTxEndpoint(tx *sql.Tx, handler ErrorAwareHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
