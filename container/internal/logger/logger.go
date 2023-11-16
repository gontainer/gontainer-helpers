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

package logger

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gontainer/gontainer-helpers/v3/grouperror"
)

type logger struct {
	output io.Writer
	prefix string
}

const (
	prefixLen = 40
)

func New(output io.Writer, prefix string) *logger {
	if len([]rune(prefix)) > prefixLen {
		prefix = string([]rune(prefix)[:prefixLen-3]) + "..."
	} else if len([]rune(prefix)) < prefixLen {
		prefix = prefix + strings.Repeat(" ", prefixLen-len([]rune(prefix)))
	}

	return &logger{
		output: output,
		prefix: prefix,
	}
}

// TODO: color only when l.output == os.Stdout || l.output == os.Stderr

func (l *logger) Info(s string) {
	l.mustPrint("\033[1;37m" + time.Now().Format("2006-01-02 15:04:05") + "\033[0m")
	l.mustPrint(" | ")
	l.mustPrint(l.prefix)
	l.mustPrint(" | ")
	l.mustPrint("\033[1;33m" + s + "\033[0m\n")
}

func (l *logger) Error(err error) {
	errs := grouperror.Collection(err)
	if len(errs) > 1 {
		l.Error(fmt.Errorf("%d errors:", len(errs)))
		for _, x := range errs {
			l.Error(x)
		}
		return
	}
	l.mustPrint("\033[1;37m" + time.Now().Format("2006-01-02 15:04:05") + "\033[0m")
	l.mustPrint(" | ")
	l.mustPrint(l.prefix)
	l.mustPrint(" | ")
	l.mustPrint("\033[1;37;41m" + err.Error() + "\033[0m\n")
}

func (l *logger) mustPrint(s string) {
	_, err := l.output.Write([]byte(s))
	if err != nil {
		panic("container.logger.Info: " + err.Error())
	}
}
