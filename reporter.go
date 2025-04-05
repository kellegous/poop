package poop

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const (
	unknownPathText = "??"
)

var DefaultReporter = NewDefaultReporter(PathLastNSegments(2))

func PathBase(path string) string {
	if path == "" {
		return unknownPathText
	}
	return filepath.Base(path)
}

func PathLastNSegments(n int) func(string) string {
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

type linkLine struct {
	function string
	path     string
	line     string
	message  string
}

func (l *linkLine) pathWidth() int {
	return len(l.function) + len(l.path) + len(l.line) + 3
}

func (l *linkLine) render(
	pathWidth int,
	forFunc func(...any) string,
	forPath func(...any) string,
	forLine func(...any) string,
	forMsg func(...any) string,
) string {
	if l.line == "" {
		return fmt.Sprintf(
			"%s%s  %s",
			forPath(l.path),
			strings.Repeat(" ", pathWidth-l.pathWidth()),
			forMsg(l.message),
		)
	}

	return fmt.Sprintf(
		"%s %s:%s%s %s",
		forFunc(l.function),
		forPath(l.path),
		forLine(l.line),
		strings.Repeat(" ", pathWidth-l.pathWidth()),
		forMsg(l.message),
	)
}

func NewDefaultReporter(
	pathFormatter func(string) string,
) func(w io.Writer, err error) error {
	return func(w io.Writer, err error) error {

		var lines []*linkLine
		for e := range IterChain(err) {
			if chErr, ok := e.(*chainedError); ok {
				var msg string
				if cur := chErr.current; cur != nil {
					msg = cur.Error()
				} else {
					msg = "↓"
				}

				// TODO(kellegous): Need to add function to reporter
				f := chErr.frame()
				lines = append(lines, &linkLine{
					function: f.function,
					path:     pathFormatter(f.file),
					line:     fmt.Sprintf("%d", f.line),
					message:  msg,
				})
			} else {
				lines = append(lines, &linkLine{
					function: "",
					path:     pathFormatter(""),
					line:     "",
					message:  e.Error(),
				})
			}
		}

		var offset int
		for _, l := range lines {
			offset = max(offset, l.pathWidth())
		}

		banner := color.New(color.FgWhite, color.BgRed).SprintFunc()
		if _, err := fmt.Fprintf(w, "%s\n", banner(" Fatal Error ")); err != nil {
			return err
		}

		forFunc := color.New(color.FgCyan).SprintFunc()
		forPath := color.New(color.FgGreen).SprintFunc()
		forLine := color.New(color.FgYellow).SprintFunc()
		forMsg := color.New().SprintFunc()

		for _, l := range lines {
			if _, err := fmt.Fprintf(w, "%s\n", l.render(offset, forFunc, forPath, forLine, forMsg)); err != nil {
				return err
			}
		}

		return nil
	}
}
