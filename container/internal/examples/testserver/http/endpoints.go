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
	"fmt"
	"net/http"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/repos"
)

// NewMyEndpoint depends on
//   - [repos.UserRepo]
//   - [repos.ImageRepo]
//
// These two objects depends on *sql.Tx.
// Container automatically builds *sql.Tx and injects it properly in the given scope.
func NewMyEndpoint(userRepo repos.UserRepo, imageRepo repos.ImageRepo) ErrorAwareHandler {
	return ErrorAwareHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		s := "MyEndpoint:\n"
		s += fmt.Sprintf("\tTxID: %p\n", userRepo.Tx)
		s += fmt.Sprintf("\tuserRepo.Tx == imageRepo.Tx: %t\n", userRepo.Tx == imageRepo.Tx)

		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		_, _ = w.Write([]byte(s))

		return nil
	})
}
