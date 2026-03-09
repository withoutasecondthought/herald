package db

import (
	"encoding/json"
	"fmt"
)

func loadAndroid(data []byte) (map[string]AndroidDevice, error) {
	var devices map[string]AndroidDevice

	err := json.Unmarshal(data, &devices)
	if err != nil {
		return nil, fmt.Errorf("parsing android JSON: %w", err)
	}

	return devices, nil
}
