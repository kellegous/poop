package poop

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

const (
	unknownPathText = "??"
)

type PathFormatter func(string) string
type FuncFormatter func(string) string

var DefaultReporter = NewDefaultReporter(PathLastNSegments(2), OmitImportPath)

func PathBase(path string) string {
	if path == "" {
		return unknownPathText
	}
	return filepath.Base(path)
}

func PathLastNSegments(n int) PathFormatter {
	return func(path string) string {
		if path == "" {
			return unknownPathText
		}
		pfx := path
		for i := 0; i < n; i++ {
			idx := strings.LastIndex(pfx, "/")
			if idx == -1 {
				break
			}
			pfx = pfx[:idx]
		}
		if len(pfx) == len(path) {
			return path
		}
		return path[len(pfx)+1:]
	}
}

func OmitImportPath(f string) string {
	ix := strings.LastIndex(f, "/")
	if ix == -1 {
		return f
	}
	return f[ix+1:]
}

type table struct {
	cols [3]int
	rows []*link
}

func (t *table) render(
	w io.Writer,
	fmtFunc func(...any) string,
	fmtPath func(...any) string,
	fmtLine func(...any) string,
	fmtMsg func(...any) string,
) error {

	for _, row := range t.rows {
		msg := row.message
		if len(row.message) > t.cols[2] {
			msg = msg[:t.cols[2]] + "..."
		}
		if !row.hasFrame() {
			if _, err := fmt.Fprintf(
				w,
				"%s %s %s\n",
				strings.Repeat(" ", t.cols[0]),
				strings.Repeat(" ", t.cols[1]),
				fmtMsg(msg),
			); err != nil {
				return err
			}
		} else {
			pathLineN := len(row.path) + len(row.line) + 1
			pathLine := fmt.Sprintf("%s:%s", fmtPath(row.path), fmtLine(row.line))
			if _, err := fmt.Fprintf(
				w,
				"%s %s %s\n",
				fmtFunc(row.function+strings.Repeat(" ", t.cols[0]-len(row.function))),
				pathLine+strings.Repeat(" ", t.cols[1]-pathLineN),
				fmtMsg(msg),
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildTable(err error, pathFormatter PathFormatter, funcFormatter FuncFormatter) *table {
	var links []*link
	for e := range IterChain(err) {
		if chErr, ok := e.(*chainedError); ok {
			var msg string
			if cur := chErr.current; cur != nil {
				msg = cur.Error()
			} else {
				msg = "↓"
			}

			f := chErr.frame()
			links = append(links, &link{
				function: funcFormatter(f.function),
				path:     pathFormatter(f.file),
				line:     fmt.Sprintf("%d", f.line),
				message:  msg,
			})
		} else {
			links = append(links, &link{
				function: "",
				path:     pathFormatter(""),
				line:     "",
				message:  e.Error(),
			})
		}
	}

	t := table{
		rows: links,
	}

	for _, l := range links {
		t.cols[0] = max(t.cols[0], len(l.function))
		t.cols[1] = max(t.cols[1], len(l.path)+len(l.line)+1)
		t.cols[2] = max(t.cols[2], len(l.message))
	}

	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		// if we're in a tty, try to clip the message column. However, do
		// not clip it under 10 characters.
		if pw := w - t.cols[0] - t.cols[1] - 2; pw > 10 {
			t.cols[2] = pw
		}
	}

	return &t
}

type link struct {
	function string
	path     string
	line     string
	message  string
}

func (l *link) hasFrame() bool {
	return l.path != ""
}

func NewDefaultReporter(
	pathFormatter PathFormatter,
	funcFormatter FuncFormatter,
) func(w io.Writer, err error) error {
	return func(w io.Writer, err error) error {
		t := buildTable(err, pathFormatter, funcFormatter)
		forFunc := color.New(color.FgCyan).SprintFunc()
		forPath := color.New(color.FgGreen).SprintFunc()
		forLine := color.New(color.FgYellow).SprintFunc()
		forMsg := color.New().SprintFunc()
		return t.render(w, forFunc, forPath, forLine, forMsg)
	}
}
