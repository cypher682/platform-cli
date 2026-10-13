package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type Format string

const (
	Text Format = "text"
	JSON Format = "json"
)

type Options struct {
	Format Format
	Quiet  bool
	NoColor bool
}

type Formatter interface {
	Print(v any) error
	Printf(format string, args ...any) error
	Println(v any) error
	Fprintf(w io.Writer, format string, args ...any) error
	WriteJSON(v any) error
}

func New(opts Options) Formatter {
	if opts.NoColor {
		color.NoColor = true
	}
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
	return &formatter{opts: opts}
}

type formatter struct {
	opts Options
}

func (f *formatter) WriteJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (f *formatter) Fprintf(w io.Writer, format string, args ...any) error {
	if f.opts.Quiet && f.opts.Format != JSON {
		return nil
	}
	if f.opts.Format == JSON {
		m := map[string]any{"message": fmt.Sprintf(format, args...)}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	}
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

func (f *formatter) Print(v any) error {
	if f.opts.Quiet && f.opts.Format != JSON {
		return nil
	}
	if f.opts.Format == JSON {
		return f.WriteJSON(v)
	}
	_, err := fmt.Print(v)
	return err
}

func (f *formatter) Printf(format string, args ...any) error {
	if f.opts.Quiet && f.opts.Format != JSON {
		return nil
	}
	if f.opts.Format == JSON {
		m := map[string]any{"message": fmt.Sprintf(format, args...)}
		return f.WriteJSON(m)
	}
	_, err := fmt.Printf(format, args...)
	return err
}

func (f *formatter) Println(v any) error {
	if f.opts.Quiet && f.opts.Format != JSON {
		return nil
	}
	if f.opts.Format == JSON {
		return f.WriteJSON(v)
	}
	_, err := fmt.Println(v)
	return err
}
