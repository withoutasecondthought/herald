package herald

import "embed"

// defaultDataFS holds the built-in JSON data files embedded at compile time.
// Used as the default data source when NewParser is called without WithDataDir.
//
//go:embed data/bots.json data/apple.json data/android.json data/darwin.json
var defaultDataFS embed.FS
