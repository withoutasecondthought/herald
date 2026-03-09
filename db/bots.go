package db

import (
	"encoding/json"
	"fmt"

	"github.com/withoutasecondthought/herald/trie"
)

type botRecord struct {
	Pattern  string `json:"pattern"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Category string `json:"category"`
}

func loadBots(data []byte) (*trie.Trie, error) {
	records, err := parseBotRecords(data)
	if err != nil {
		return nil, err
	}

	t := trie.New()

	insertBotRecords(t, records)

	return t, nil
}

func parseBotRecords(data []byte) ([]botRecord, error) {
	var records []botRecord

	err := json.Unmarshal(data, &records)
	if err != nil {
		return nil, fmt.Errorf("parsing bots JSON: %w", err)
	}

	return records, nil
}

func insertBotRecords(t *trie.Trie, records []botRecord) {
	for i := range records {
		r := &records[i]

		entry := &trie.BotEntry{
			Pattern:  r.Pattern,
			Name:     r.Name,
			Owner:    r.Owner,
			Category: r.Category,
		}

		t.Insert(r.Pattern, entry)
	}
}
