package restoreassert

import _ "embed"

// DefaultTemplate holds the raw byte content of the default configuration file.
// Using the 'go:embed' directive, the template is compiled directly into the 
// binary executable. This ensures the CLI tool remains portable and standalone, 
// allowing users to generate a fresh config file anywhere without needing 
// external template files.
//go:embed config/template/restore-config.yaml
var DefaultTemplate []byte