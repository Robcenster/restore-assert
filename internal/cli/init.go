package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigTemplate = `engine: "postgres" # DBType: postgres/mssql/oracle etc

docker:
  image: "postgres:17-alpine" # Docker image for the temporary environment
  container_name: "restore-assert-vp" # Name of the spawned container
  memory_limit: "512MB" # RAM limit for the container
  cpu_limit: "1.0" # CPU core limit (e.g., 0.5, 1.0)

database:
  db_name: "postgres" # Target database name
  user: "admin" # Database administrative user
  password: "very_secret_password" # Administrative password
  extensions: # List of required PostgreSQL extensions
    - "uuid-ossp"
    - "pg_trgm"
  roles: # Roles to be created before restoration
    - "postgres"
    - "warehouse_analyst"
    - "warehouse_app_user"

  settings: # Custom postgresql.conf parameters (via -c flags)
    max_connections: "50" # Maximum concurrent connections
    shared_buffers: "128MB" # Memory for caching data
    fsync: "off" # Disables disk sync for faster testing (unsafe for prod)

restore:
  parallel_jobs: 1 # Number of threads for pg_restore
  analyze: true # Run ANALYZE to update statistics after restore
  on_error_stop: true # Halt if an error occurs
  single_transaction: false # Use a single transaction for the whole restore
  no_owner: true # Skip restoration of object ownership
  no_privileges: true # Skip restoration of access privileges (GRANT/REVOKE)
  full_restore_logs: false # Print verbose logs from restore utilities
  show_db_info: true # Display database summary after restore
  hide_success_tests: false # Only show failed assertions in the report

# Logical validation checks (Commented out by default)
# asserts:
#   tables:
#     - name: "users"
#       metrics:
#         - type: "row_count"
#           condition: "gt"
#           expected: 100
`

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new restore-config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "restore-config.yaml"

			if _, err := os.Stat(filename); err == nil {
				return fmt.Errorf("config file '%s' already exists", filename)
			}

			err := os.WriteFile(filename, []byte(defaultConfigTemplate), 0644)
			if err != nil {
				return fmt.Errorf("failed to create config: %w", err)
			}

			fmt.Printf("✨ Created %s with default settings\n", filename)
			return nil
		},
	}
}
