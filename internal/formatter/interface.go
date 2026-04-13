package formatter

// ClusterSnapshot и DatabaseSnapshot описывают структуру БД.
// Они лежат здесь, чтобы любой реализатор Formatter знал, какие данные он получит.
type ClusterSnapshot struct {
	Version   string
	Roles     []string
	Databases []DatabaseSnapshot
}

type DatabaseSnapshot struct {
	Name    string
	Schemas map[string][]string
}

// Formatter определяет контракт для вывода информации пользователю.
type Formatter interface {
	Info(format string, args ...any)
	Success(format string, args ...any)
	Error(format string, args ...any)
	Warning(format string, args ...any)

	// Step используется для выделения основных этапов работы (например, "Восстановление...")
	Step(format string, args ...any)

	PrintClusterReport(cluster *ClusterSnapshot)
}
