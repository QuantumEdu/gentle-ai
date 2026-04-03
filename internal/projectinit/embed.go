package projectinit

import "embed"

// FS holds all project template files embedded into the binary at compile time.
// Templates live under internal/projectinit/templates/<template-id>/.
//
//go:embed templates
var FS embed.FS
