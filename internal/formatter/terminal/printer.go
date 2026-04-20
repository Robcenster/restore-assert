package terminal

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Robcenster/restore-assert/internal/formatter"
)

// Colors ANSI
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	blue   = "\033[34m"
)

type Printer struct {
	out io.Writer
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{
		out: w,
	}
}

// Info displays a standard message
func (p *Printer) Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "  %s\n", msg)
}

// Step visually marks the start of a new phase of work
func (p *Printer) Step(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%s==>%s %s\n", blue, reset, msg)
}

// Success displays a success message (in green)
func (p *Printer) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sSUCCESS:%s %s\n", green, reset, msg)
}

// Error displays an error message (in red)
func (p *Printer) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sERROR:%s %s\n", red, reset, msg)
}

// Warning displays a warning (in yellow)
func (p *Printer) Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sWARNING:%s %s\n", yellow, reset, msg)
}

// PrintClusterReport displays the database tree
func (p *Printer) PrintClusterReport(cluster *formatter.ClusterSnapshot) {
	p.Info("")
	separator := strings.Repeat("=", 60)

	fmt.Fprintf(p.out, "%s%s%s\n", cyan, separator, reset)
	fmt.Fprintf(p.out, "%sPOSTGRES CLUSTER:%s %s\n", yellow, reset, cluster.Version)
	fmt.Fprintf(p.out, "%sROLES:%s            %s\n", yellow, reset, strings.Join(cluster.Roles, ", "))
	fmt.Fprintf(p.out, "%s%s%s\n", cyan, separator, reset)

	for _, db := range cluster.Databases {
		if len(db.Schemas) == 0 {
			fmt.Fprintf(p.out, "\nDB: %s[%s]%s (no tables)\n", cyan, db.Name, reset)
			continue
		}

		fmt.Fprintf(p.out, "\nDB: %s[%s]%s\n", cyan, db.Name, reset)

		sNames := make([]string, 0, len(db.Schemas))
		for s := range db.Schemas {
			sNames = append(sNames, s)
		}
		sort.Strings(sNames)

		for _, sName := range sNames {
			fmt.Fprintf(p.out, "  Schema: %s%s%s\n", yellow, sName, reset)
			tables := db.Schemas[sName]

			for i, tName := range tables {
				prefix := "  ├── "
				if i == len(tables)-1 {
					prefix = "  └── "
				}
				fmt.Fprintf(p.out, "%s%s\n", prefix, tName)
			}
		}
	}
	fmt.Fprintf(p.out, "\n%s%s%s\n\n", cyan, separator, reset)
}
