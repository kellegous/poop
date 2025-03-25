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
	path    string
	line    string
	message string
}

func (l *linkLine) pathWidth() int {
	return len(l.path) + len(l.line) + 2
}

func (l *linkLine) render(
	pathWidth int,
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
		"%s:%s%s %s",
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

				lines = append(lines, &linkLine{
					path:    pathFormatter(chErr.File),
					line:    fmt.Sprintf("%d", chErr.Line),
					message: msg,
				})
			} else {
				lines = append(lines, &linkLine{
					path:    pathFormatter(""),
					line:    "",
					message: e.Error(),
				})
			}
		}

		var offset int
		for _, l := range lines {
			offset = max(offset, l.pathWidth())
		}

		banner := color.New(color.FgBlack, color.BgRed).SprintFunc()
		if _, err := fmt.Fprintf(w, "%s\n", banner(" Fatal Error ")); err != nil {
		}

		forPath := color.New(color.FgGreen).SprintFunc()
		forLine := color.New(color.FgYellow).SprintFunc()
		forMsg := color.New().SprintFunc()

		for _, l := range lines {
			if _, err := fmt.Fprintf(w, "%s\n", l.render(offset, forPath, forLine, forMsg)); err != nil {
				return err
			}
		}

		return nil
	}
}
