package herald

import (
	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/scoring"
	"github.com/withoutasecondthought/herald/tokenizer"
)

const botScoreThreshold = 0.8

// detectBot runs bot detection: trie lookup for known bots, then scoring for unknowns.
// Returns true if a bot was detected (caller should stop pipeline).
func detectBot(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	entry, found := database.BotTrie.Search(result.Raw)
	if found {
		result.ClientType = ClientTypeBot
		result.Bot = BotInfo{
			Name:       entry.Name,
			Owner:      entry.Owner,
			Category:   BotCategory(entry.Category),
			Confidence: 1.0,
		}

		return true
	}

	features := scoring.ExtractFeatures(tokens, result.Raw)
	score := scoring.Score(features)

	if score >= botScoreThreshold {
		result.ClientType = ClientTypeBot
		result.Bot = BotInfo{
			Confidence: score,
		}

		return true
	}

	return false
}
