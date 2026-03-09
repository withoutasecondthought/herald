package herald

import (
	"fmt"
	"sync"

	"github.com/withoutasecondthought/herald/db"
)

// Parser holds the in-memory database and pipeline for UA parsing.
// Create with NewParser, then call Parse/ParseWithHints/DetectType.
type Parser struct {
	db *db.Database
}

// Option configures how NewParser loads data.
type Option func(*parserConfig)

type parserConfig struct {
	dataDir     string // load only from this directory (no embedded defaults)
	overrideDir string // merge overrides on top of embedded defaults
}

// WithDataDir loads data exclusively from the given directory, ignoring embedded defaults.
// Use this when you maintain your own complete set of data files.
func WithDataDir(dir string) Option {
	return func(c *parserConfig) {
		c.dataDir = dir
	}
}

// WithOverrides merges user data on top of embedded defaults.
// The override directory may contain any subset of the data files (bots.json, apple.json,
// android.json, darwin.json). Only files that exist are merged; missing files are skipped.
// For device/bot data, override entries are added to or replace the built-in entries.
func WithOverrides(dir string) Option {
	return func(c *parserConfig) {
		c.overrideDir = dir
	}
}

// NewParser creates a ready-to-use parser.
//
// Without options, it uses the embedded default data (zero-config):
//
//	p, err := herald.NewParser()
//
// With WithDataDir, it loads data exclusively from the given directory:
//
//	p, err := herald.NewParser(herald.WithDataDir("path/to/data"))
//
// With WithOverrides, it uses embedded defaults and merges user data on top:
//
//	p, err := herald.NewParser(herald.WithOverrides("path/to/overrides"))
func NewParser(opts ...Option) (*Parser, error) {
	var cfg parserConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	database, err := loadDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("herald: loading data: %w", err)
	}

	return &Parser{db: database}, nil
}

func loadDatabase(cfg parserConfig) (*db.Database, error) {
	if cfg.dataDir != "" {
		database, err := db.LoadAll(cfg.dataDir)
		if err != nil {
			return nil, fmt.Errorf("loading from dir: %w", err)
		}

		return database, nil
	}

	database, err := db.LoadFromFS(defaultDataFS)
	if err != nil {
		return nil, fmt.Errorf("loading embedded data: %w", err)
	}

	if cfg.overrideDir != "" {
		mergeErr := db.MergeDir(database, cfg.overrideDir)
		if mergeErr != nil {
			return nil, fmt.Errorf("merging overrides: %w", mergeErr)
		}
	}

	return database, nil
}

// Database returns the loaded database for direct access if needed.
func (p *Parser) Database() *db.Database {
	return p.db
}

//nolint:gochecknoglobals // package-level default parser for convenience API
var (
	defaultParser *Parser
	defaultMu     sync.RWMutex
)

// Init initializes the default parser for the package-level Parse/ParseWithHints/DetectType.
//
// Without options, it uses embedded defaults:
//
//	herald.Init()
//
// With options:
//
//	herald.Init(herald.WithOverrides("path/to/extra"))
func Init(opts ...Option) error {
	p, err := NewParser(opts...)
	if err != nil {
		return err
	}

	defaultMu.Lock()
	defaultParser = p
	defaultMu.Unlock()

	return nil
}

func getDefault() *Parser {
	defaultMu.RLock()

	p := defaultParser

	defaultMu.RUnlock()

	return p
}
