package formatter

// internal/formatter/types.go
type ClusterSnapshot struct {
	Version   string
	Roles     []string
	Databases []DatabaseSnapshot
}

type DatabaseSnapshot struct {
	Name    string
	Schemas map[string][]string // Имя схемы -> список таблиц
}
