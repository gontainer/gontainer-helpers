package logger

import (
	"io"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

type logger struct {
	output     io.Writer
	prefix     string
	fileColumn string
}

var (
	reFileNameSuffix = regexp.MustCompile(" \\+0x[a-z0-9]+")
)

const (
	prefixLen = 40
)

func New(output io.Writer, prefix string) *logger {
	if len([]rune(prefix)) > prefixLen {
		prefix = string([]rune(prefix)[:prefixLen-3]) + "..."
	} else if len([]rune(prefix)) < prefixLen {
		prefix = prefix + strings.Repeat(" ", prefixLen-len([]rune(prefix)))
	}

	lines := strings.Split(string(debug.Stack()), "\n")
	line := lines[8][1:] // 8th is the desired line, and we have to remove the leading tab
	line = reFileNameSuffix.ReplaceAllString(line, "")

	return &logger{
		output:     output,
		prefix:     prefix,
		fileColumn: adjustLen(line, 30, "..."),
	}
}

func adjustLen(s string, le int, prefix string) string {
	if len([]rune(s)) > le {
		s = prefix + string([]rune(s)[len([]rune(s))-le+len([]rune(prefix)):])
	} else if len([]rune(s)) < le {
		s = s + strings.Repeat(" ", le-len([]rune(s)))
	}
	return s
}

func (l *logger) Info(s string) {
	l.mustPrint("\033[1;37m" + time.Now().Format("2006-01-02 15:04:05") + "\033[0m")
	l.mustPrint(" | ")
	l.mustPrint("\033[1;37m" + l.fileColumn + "\033[0m")
	l.mustPrint(" | ")
	l.mustPrint(l.prefix)
	l.mustPrint(" | ")
	l.mustPrint("\033[1;33m" + s + "\033[0m\n")
}

func (l *logger) Error(err error) {
	l.mustPrint("\033[1;37m" + time.Now().Format("2006-01-02 15:04:05") + "\033[0m")
	l.mustPrint(" | ")
	l.mustPrint("\033[1;37m" + l.fileColumn + "\033[0m")
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
