package formatter

import (
	"fmt"
	"sort"
)

type DbObject struct {
	Name string
	Type string
}

func PrintDatabaseStructure(structure map[string][]DbObject) {
	if len(structure) == 0 {
		fmt.Println("ℹ️ No objects in schemas")
		return
	}
	fmt.Println("\n=== 📊 Restored Database Structure ===")

	schemas := make([]string, 0, len(structure))
	for schema := range structure {
		schemas = append(schemas, schema)
	}
	sort.Strings(schemas)

	for _, schema := range schemas {
		fmt.Printf("\n📂 Schema: [%s]\n", schema)
		objects := structure[schema]

		for i, obj := range objects {
			prefix := "  ├─"
			if i == len(objects)-1 {
				prefix = "  └─"
			}
			fmt.Printf("%s %s [%s]\n", prefix, obj.Name, obj.Type)
		}
	}

	fmt.Println("\n======================================")
}
