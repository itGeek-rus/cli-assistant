package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type Format string

const (
	FormatHuman Format = "human"
	FormatJSON  Format = "json"
)

func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	default:
		return FormatHuman
	}
}

type Printer struct {
	format Format
	w      io.Writer
}

func NewPrinter(format Format, w io.Writer) *Printer {
	if w == nil {
		w = os.Stdout
	}
	return &Printer{format: format, w: w}
}

func (p *Printer) Print(v any) error {
	switch p.format {
	case FormatJSON:
		enc := json.NewEncoder(p.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	default:
		return fmt.Errorf("human printer: unsupported type %T", v)
	}
}

func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	if p.format == FormatJSON {
		return fmt.Errorf("user Print with struct for JSON")
	}
	tw := tabwriter.NewWriter(p.w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "%s\n", strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\n", strings.Join(row, "\t"))
	}
	return tw.Flush()
}
