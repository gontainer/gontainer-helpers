// Copyright (c) 2023 Bartłomiej Krukowski
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

package container_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_Server(t *testing.T) {
	re := regexp.MustCompile("TxID: 0x[0-9a-f]+")

	c := container.BuildContainer()
	mux, err := c.Get("mux")
	require.NoError(t, err)
	s := httptest.NewServer(mux.(http.Handler))
	defer s.Close()
	client := s.Client()

	// we cannot run it in different goroutines due to limitations in github.com/DATA-DOG/go-sqlmock
	prev := ""
	for i := 0; i < 100; i++ {
		r, err := client.Get(s.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, r.StatusCode)
		buff, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		matches := re.FindAllString(string(buff), -1)
		require.Len(t, matches, 2)
		require.Equal(t, matches[0], matches[1]) // make we use a single transaction in the scope of the given request
		require.NotEqual(t, matches[0], prev)    // make sure we don't share the same transaction between different requests

		prev = matches[0]
	}
}
