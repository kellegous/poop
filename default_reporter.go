package poop

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const unknownValueText = "??"

type FormattedValue interface {
	// Format returns the colored, formatted value.
	Format() string

	// Width returns the character width of the value.
	// TODO(kellegous): This doesn't take into account
	// glyphs that are multi-byte but display as a single
	// character.
	Width() int
}

// PathFormatter formats a path and line number into a formatted value.
type PathFormatter func(path string, line int) FormattedValue

// FuncFormatter formats a function name into a formatted value.
type FuncFormatter func(function string) FormattedValue

type pathValue struct {
	pathColorFn func(string) string
	lineColorFn func(string) string
	path        string
	line        string
}

func (p *pathValue) Format() string {
	return fmt.Sprintf("%s:%s", p.pathColorFn(p.path), p.lineColorFn(p.line))
}

func (p *pathValue) Width() int {
	return len(p.path) + len(p.line) + 1
}

// NewPathFormatter creates a new PathFormatter that formats a path and line number into a formatted value.
func NewPathFormatter(
	pathStringFn func(string) string,
	lineStringFn func(int) string,
	lineColorFn func(string) string,
	pathColorFn func(string) string,
) PathFormatter {
	return func(path string, line int) FormattedValue {
		return &pathValue{
			pathColorFn: pathColorFn,
			lineColorFn: lineColorFn,
			path:        pathStringFn(path),
			line:        lineStringFn(line),
		}
	}
}

type funcValue struct {
	colorFn  func(string) string
	function string
}

func (f *funcValue) Format() string {
	return f.colorFn(f.function)
}

func (f *funcValue) Width() int {
	return len(f.function)
}

// NewFuncFormatter creates a new FuncFormatter that formats a function name into a formatted value.
func NewFuncFormatter(
	stringFn func(string) string,
	colorFn func(string) string,
) FuncFormatter {
	return func(function string) FormattedValue {
		return &funcValue{
			colorFn:  colorFn,
			function: stringFn(function),
		}
	}
}

// WithColor creates a new color function that wraps the given color attribute.
func WithColor(c color.Attribute) func(string) string {
	return func(s string) string {
		return color.New(c).Sprint(s)
	}
}

// DefaultReporter is the default reporter that is used when no reporter is provided.
var DefaultReporter = NewDefaultReporter()

// PathBase returns the base name of the given path.
func PathBase(path string) string {
	if path == "" {
		return unknownValueText
	}
	return filepath.Base(path)
}

// PathLastNSegments returns a function that returns the last n segments of the given path.
func PathLastNSegments(n int) func(path string) string {
	return func(path string) string {
		if path == "" {
			return unknownValueText
		}
		pfx := path
		for range n {
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

// OmitImportPath returns the last segment of the given function name.
func OmitImportPath(f string) string {
	if f == "" {
		return unknownValueText
	}

	ix := strings.LastIndex(f, "/")
	if ix == -1 {
		return f
	}
	return f[ix+1:]
}

type link struct {
	funcValue FormattedValue
	pathValue FormattedValue
	message   string
	isChain   bool
}

func (l *link) getMessage() string {
	if l.isChain && l.message == "" {
		return "↓"
	}
	return l.message
}

// DefaultReporterOption is a function that modifies the default reporter options.
type DefaultReporterOption func(*DefaultReporterOptions)

// DefaultReporterOptions is the options for the default reporter.
type DefaultReporterOptions struct {
	pathFormatter PathFormatter
	funcFormatter FuncFormatter
}

func (o *DefaultReporterOptions) applyDefaults() {
	if o.pathFormatter == nil {
		o.pathFormatter = NewPathFormatter(
			PathLastNSegments(2),
			func(line int) string {
				if line == 0 {
					return unknownValueText
				}
				return fmt.Sprintf("%d", line)
			},
			WithColor(color.FgGreen),
			WithColor(color.FgYellow),
		)
	}
	if o.funcFormatter == nil {
		o.funcFormatter = NewFuncFormatter(
			OmitImportPath,
			WithColor(color.FgCyan),
		)
	}
}

func compressLinks(links []*link) []*link {
	if len(links) > 1 {
		last := links[len(links)-1]
		prev := links[len(links)-2]

		if !last.isChain && prev.isChain && prev.message == "" {
			prev.message = last.message
			return links[:len(links)-1]
		}
	}

	return links
}

func (o *DefaultReporterOptions) render(w io.Writer, err error) error {
	var links []*link
	for e := range IterChain(err) {
		if chErr, ok := e.(*ChainedError); ok {
			var msg string
			if cur := chErr.message; cur != "" {
				msg = cur
			}

			f := chErr.frame()
			links = append(links, &link{
				funcValue: o.funcFormatter(f.Function),
				pathValue: o.pathFormatter(f.File, f.Line),
				message:   msg,
				isChain:   true,
			})
		} else {
			links = append(links, &link{
				funcValue: o.funcFormatter(""),
				pathValue: o.pathFormatter("", 0),
				message:   e.Error(),
			})
		}
	}

	links = compressLinks(links)

	var widths [2]int

	for _, l := range links {
		widths[0] = max(widths[0], l.funcValue.Width())
		widths[1] = max(widths[1], l.pathValue.Width())
	}

	for _, link := range links {
		if _, err := fmt.Fprintf(
			w,
			"%s %s %s\n",
			link.funcValue.Format()+strings.Repeat(" ", widths[0]-link.funcValue.Width()),
			link.pathValue.Format()+strings.Repeat(" ", widths[1]-link.pathValue.Width()),
			link.getMessage(),
		); err != nil {
			return err
		}
	}

	return nil
}

// WithPathFormatter sets the path formatter for the default reporter.
func WithPathFormatter(formatter PathFormatter) DefaultReporterOption {
	return func(o *DefaultReporterOptions) {
		o.pathFormatter = formatter
	}
}

// WithFuncFormatter sets the function formatter for the default reporter.
func WithFuncFormatter(formatter FuncFormatter) DefaultReporterOption {
	return func(o *DefaultReporterOptions) {
		o.funcFormatter = formatter
	}
}

// NewDefaultReporter creates a new default reporter with the given options.
func NewDefaultReporter(opts ...DefaultReporterOption) func(w io.Writer, err error) error {
	o := DefaultReporterOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	o.applyDefaults()
	return o.render
}
