package terminal

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Robcenster/restore-assert/internal/formatter"
)

// Цвета ANSI (без жирного шрифта, как договаривались)
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	blue   = "\033[34m"
)

// Printer реализует интерфейс formatter.Formatter для терминала.
type Printer struct {
	out io.Writer
}

// NewPrinter создает новый форматер для терминала.
func NewPrinter(w io.Writer) *Printer {
	return &Printer{
		out: w,
	}
}

// Info выводит обычное информационное сообщение
func (p *Printer) Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "  %s\n", msg) // Отступ для красоты
}

// Step визуально выделяет начало нового этапа работы
func (p *Printer) Step(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	// Используем синий цвет и стрелочку для индикации шага
	fmt.Fprintf(p.out, "%s==>%s %s\n", blue, reset, msg)
}

// Success выводит сообщение об успехе (зеленым)
func (p *Printer) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sSUCCESS:%s %s\n", green, reset, msg)
}

// Error выводит сообщение об ошибке (красным)
func (p *Printer) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sERROR:%s %s\n", red, reset, msg)
}

// Warning выводит предупреждение (желтым)
func (p *Printer) Warning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(p.out, "%sWARNING:%s %s\n", yellow, reset, msg)
}

// PrintClusterReport выводит дерево базы данных
func (p *Printer) PrintClusterReport(cluster *formatter.ClusterSnapshot) {
	p.Info("") // Пустая строка
	separator := strings.Repeat("=", 60)

	fmt.Fprintf(p.out, "%s%s%s\n", cyan, separator, reset)
	fmt.Fprintf(p.out, "🏗️  %sPOSTGRES CLUSTER:%s %s\n", yellow, reset, cluster.Version)
	fmt.Fprintf(p.out, "👥 %sROLES:%s            %s\n", yellow, reset, strings.Join(cluster.Roles, ", "))
	fmt.Fprintf(p.out, "%s%s%s\n", cyan, separator, reset)

	for _, db := range cluster.Databases {
		if len(db.Schemas) == 0 {
			fmt.Fprintf(p.out, "\n📦 DB: %s[%s]%s (no tables)\n", cyan, db.Name, reset)
			continue
		}

		fmt.Fprintf(p.out, "\n📊 DB: %s[%s]%s\n", cyan, db.Name, reset)

		// Сортируем схемы
		sNames := make([]string, 0, len(db.Schemas))
		for s := range db.Schemas {
			sNames = append(sNames, s)
		}
		sort.Strings(sNames)

		for _, sName := range sNames {
			fmt.Fprintf(p.out, "  📂 Schema: %s%s%s\n", yellow, sName, reset)
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
