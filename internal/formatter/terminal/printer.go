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
	headerLine := strings.Repeat("-", 64)
	subLine := strings.Repeat(".", 64)

	fmt.Fprintf(p.out, "%s%s%s\n", cyan, headerLine, reset)
	fmt.Fprintf(p.out, "%s CLUSTER :%s %s\n", cyan, reset, cluster.Version)
	fmt.Fprintf(p.out, "%s ROLES   :%s %s\n", cyan, reset, strings.Join(cluster.Roles, ", "))
	fmt.Fprintf(p.out, "%s%s%s\n", cyan, headerLine, reset)

	for _, db := range cluster.Databases {
		fmt.Fprintf(p.out, "\n%sDATABASE:%s [%s]\n", yellow, reset, db.Name)

		if len(db.Schemas) == 0 {
			fmt.Fprintf(p.out, "  (no schemas found)\n")
			continue
		}

		sNames := make([]string, 0, len(db.Schemas))
		for s := range db.Schemas {
			sNames = append(sNames, s)
		}
		sort.Strings(sNames)

		for _, sName := range sNames {
			tables := db.Schemas[sName]
			fmt.Fprintf(p.out, "  %sSCHEMA:%s %s %s(%d tables)%s\n",
				cyan, reset, sName, blue, len(tables), reset)

			if len(tables) > 0 {
				p.printTableGrid(tables, 4, "    ")
			} else {
				fmt.Fprintf(p.out, "    (empty)\n")
			}
			fmt.Fprintf(p.out, "  %s%s%s\n", blue, subLine, reset)
		}
	}
}

// printTableGrid displays a list of rows in a compact grid
func (p *Printer) printTableGrid(items []string, columns int, indent string) {
	if len(items) == 0 {
		return
	}

	maxLen := 0
	for _, item := range items {
		if len(item) > maxLen {
			maxLen = len(item)
		}
	}

	cellWidth := maxLen + 2

	for i := 0; i < len(items); i++ {
		if i%columns == 0 {
			fmt.Print(indent)
		}

		fmt.Fprintf(p.out, "%-*s", cellWidth, items[i])

		if (i+1)%columns == 0 || i == len(items)-1 {
			fmt.Fprintln(p.out)
		}
	}
}
