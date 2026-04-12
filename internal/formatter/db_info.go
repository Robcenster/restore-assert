package formatter

import (
	"fmt"
	"sort"
	"strings"
)

func PrintSimpleReport(cluster *ClusterSnapshot) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("CLUSTER: %s\n", cluster.Version)
	fmt.Printf("ROLES: %s\n", strings.Join(cluster.Roles, ", "))
	fmt.Println(strings.Repeat("=", 50))

	for _, db := range cluster.Databases {
		if len(db.Schemas) == 0 {
			fmt.Printf("\nDB: [%s] (no tables)\n", db.Name)
			continue
		}

		fmt.Printf("\nDB: [%s]\n", db.Name)

		schemas := make([]string, 0, len(db.Schemas))
		for s := range db.Schemas {
			schemas = append(schemas, s)
		}
		sort.Strings(schemas)

		for _, sName := range schemas {
			fmt.Printf("📂 Schema: %s\n", sName)
			tables := db.Schemas[sName]

			for i, tName := range tables {
				prefix := "  ├── "
				if i == len(tables)-1 {
					prefix = "  └── "
				}
				fmt.Printf("%s%s\n", prefix, tName)
			}
		}
	}
	fmt.Println("\n" + strings.Repeat("=", 50))
}
