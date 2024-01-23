//go:build production
// +build production

package assets

import "embed"

//go:embed css/*.css
//go:embed js/*.js
//go:embed fonts
var fsys embed.FS
