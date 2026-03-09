package db

import (
	"encoding/json"
	"fmt"
)

func loadApple(data []byte) (map[string]string, error) {
	var models map[string]string

	err := json.Unmarshal(data, &models)
	if err != nil {
		return nil, fmt.Errorf("parsing apple JSON: %w", err)
	}

	return models, nil
}
