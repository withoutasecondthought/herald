package db

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"github.com/withoutasecondthought/herald/trie"
)

// AndroidDevice holds brand and model name for an Android device.
type AndroidDevice struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
}

// DarwinVersions maps a Darwin kernel major version to iOS and macOS versions.
type DarwinVersions struct {
	IOS   string `json:"ios"`
	MacOS string `json:"macos"`
}

// Database holds all in-memory lookup data loaded from JSON files.
// Built once at startup, used read-only during parsing.
type Database struct {
	BotTrie     *trie.Trie
	AppleModels map[string]string         // "iPhone14,2" → "iPhone 13 Pro"
	AndroidDB   map[string]AndroidDevice  // "SM-G991B" → {Samsung, Galaxy S21}
	DarwinMap   map[string]DarwinVersions // "24" → {ios:"18", macos:"15"}
}

//nolint:gochecknoglobals // constant list of expected data file names
var dataFiles = [...]string{
	"bots.json",
	"apple.json",
	"android.json",
	"darwin.json",
}

// LoadAll reads all JSON data files from dataDir and builds the in-memory database.
func LoadAll(dataDir string) (*Database, error) {
	files, err := readAllFiles(dataDir, "")
	if err != nil {
		return nil, err
	}

	return buildDatabase(files)
}

// LoadFromFS reads data files from an fs.FS (typically an embedded filesystem).
func LoadFromFS(fsys fs.FS) (*Database, error) {
	var files [len(dataFiles)][]byte

	for i, name := range dataFiles {
		data, err := fs.ReadFile(fsys, filepath.Join("data", name))
		if err != nil {
			return nil, fmt.Errorf("reading embedded %s: %w", name, err)
		}

		files[i] = data //nolint:gosec // i is bounded by dataFiles range
	}

	return buildDatabase(files)
}

func readAllFiles(dir, prefix string) ([len(dataFiles)][]byte, error) {
	var files [len(dataFiles)][]byte

	for i, name := range dataFiles {
		path := filepath.Join(dir, prefix+name)

		data, err := os.ReadFile(path) //nolint:gosec // path from trusted dataDir
		if err != nil {
			return files, fmt.Errorf("reading %s: %w", name, err)
		}

		files[i] = data //nolint:gosec // i is bounded by dataFiles range
	}

	return files, nil
}

func buildDatabase(files [len(dataFiles)][]byte) (*Database, error) {
	db := &Database{}

	var err error

	db.BotTrie, err = loadBots(files[0])
	if err != nil {
		return nil, fmt.Errorf("loading bots: %w", err)
	}

	db.AppleModels, err = loadApple(files[1])
	if err != nil {
		return nil, fmt.Errorf("loading apple models: %w", err)
	}

	db.AndroidDB, err = loadAndroid(files[2])
	if err != nil {
		return nil, fmt.Errorf("loading android models: %w", err)
	}

	db.DarwinMap, err = loadDarwin(files[3])
	if err != nil {
		return nil, fmt.Errorf("loading darwin map: %w", err)
	}

	return db, nil
}

// MergeDir loads override files from overrideDir and merges them into the database.
// Only files that exist in overrideDir are merged; missing files are silently skipped.
// For maps (apple, android, darwin), override entries replace existing ones.
// For bots, override patterns are added to the existing trie.
func MergeDir(database *Database, overrideDir string) error {
	err := mergeBots(database, overrideDir)
	if err != nil {
		return err
	}

	err = mergeApple(database, overrideDir)
	if err != nil {
		return err
	}

	err = mergeAndroid(database, overrideDir)
	if err != nil {
		return err
	}

	return mergeDarwin(database, overrideDir)
}

func mergeBots(database *Database, dir string) error {
	data, err := readFileIfExists(filepath.Join(dir, "bots.json"))
	if err != nil {
		return fmt.Errorf("reading override bots: %w", err)
	}

	if data == nil {
		return nil
	}

	records, err := parseBotRecords(data)
	if err != nil {
		return fmt.Errorf("parsing override bots: %w", err)
	}

	insertBotRecords(database.BotTrie, records)

	return nil
}

func mergeApple(database *Database, dir string) error {
	data, err := readFileIfExists(filepath.Join(dir, "apple.json"))
	if err != nil {
		return fmt.Errorf("reading override apple: %w", err)
	}

	if data == nil {
		return nil
	}

	models, err := loadApple(data)
	if err != nil {
		return fmt.Errorf("merging override apple: %w", err)
	}

	maps.Copy(database.AppleModels, models)

	return nil
}

func mergeAndroid(database *Database, dir string) error {
	data, err := readFileIfExists(filepath.Join(dir, "android.json"))
	if err != nil {
		return fmt.Errorf("reading override android: %w", err)
	}

	if data == nil {
		return nil
	}

	devices, err := loadAndroid(data)
	if err != nil {
		return fmt.Errorf("merging override android: %w", err)
	}

	maps.Copy(database.AndroidDB, devices)

	return nil
}

func mergeDarwin(database *Database, dir string) error {
	data, err := readFileIfExists(filepath.Join(dir, "darwin.json"))
	if err != nil {
		return fmt.Errorf("reading override darwin: %w", err)
	}

	if data == nil {
		return nil
	}

	versions, err := loadDarwin(data)
	if err != nil {
		return fmt.Errorf("merging override darwin: %w", err)
	}

	maps.Copy(database.DarwinMap, versions)

	return nil
}

// readFileIfExists returns file contents or nil if the file does not exist.
func readFileIfExists(path string) ([]byte, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path from trusted overrideDir
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	return data, nil
}
