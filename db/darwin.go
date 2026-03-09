package db

import (
	"encoding/json"
	"fmt"
)

func loadDarwin(data []byte) (map[string]DarwinVersions, error) {
	var m map[string]DarwinVersions

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("parsing darwin JSON: %w", err)
	}

	return m, nil
}
