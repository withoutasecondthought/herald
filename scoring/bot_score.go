package scoring

import (
	"math"
	"strconv"
	"strings"

	"github.com/withoutasecondthought/herald/tokenizer"
)

// Scoring weights and thresholds.
const (
	weightBotWord        = 0.35
	weightLibSignature   = 0.30
	weightSingleProduct  = 0.15
	weightInconsistentOS = 0.10
	weightInvalidChrome  = 0.10
	weightLowEntropy     = 0.10

	scoreEmptyUA     = 0.9
	entropyThreshold = 3.0

	chromeMinVersion = 60
	chromeMaxVersion = 170

	freqMapCap = 64
)

// BotFeatures holds extracted signals for bot classification.
type BotFeatures struct {
	IsEmpty            bool
	HasBotWord         bool
	HasLibSignature    bool
	LooksLikeBrowser   bool
	ChromeVersionValid bool
	HasConsistentOS    bool
	HasOnlyOneProduct  bool
	StringEntropy      float64
}

//nolint:gochecknoglobals // package-level lookup tables, read-only
var (
	botWords = []string{
		"bot", "crawl", "spider", "scrape",
		"fetch", "monitor", "check", "scan",
	}
	libSignatures = []string{
		"python-requests", "python-urllib", "okhttp", "dart:io",
		"go-http-client", "java/", "libwww-perl", "curl/",
		"wget/", "httpie/", "node-fetch", "axios/",
	}
)

// ExtractFeatures analyzes tokens and the raw UA string for bot-like signals.
func ExtractFeatures(tokens []tokenizer.Token, raw string) BotFeatures {
	if len(raw) == 0 {
		return BotFeatures{IsEmpty: true}
	}

	f := BotFeatures{}
	rawLower := strings.ToLower(raw)

	f.HasBotWord = containsAny(rawLower, botWords)
	f.HasLibSignature = containsAny(rawLower, libSignatures)

	hasMozillaPrefix, chromeVersion, hasOSComment, productCount := analyzeTokens(tokens)

	f.LooksLikeBrowser = hasMozillaPrefix
	f.HasOnlyOneProduct = productCount == 1
	f.ChromeVersionValid = isValidChromeVersion(chromeVersion, hasMozillaPrefix)
	f.HasConsistentOS = !hasMozillaPrefix || hasOSComment
	f.StringEntropy = shannonEntropy(raw)

	return f
}

func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}

	return false
}

func analyzeTokens(tokens []tokenizer.Token) (hasMozilla bool, chromeVer string, hasOS bool, products int) {
	for _, tok := range tokens {
		switch tok.Kind {
		case tokenizer.KindProduct:
			products++

			if tok.Name == "Mozilla" && tok.Version == "5.0" {
				hasMozilla = true
			}

			if tok.Name == "Chrome" || tok.Name == "Chromium" {
				chromeVer = tok.Version
			}
		case tokenizer.KindComment:
			if hasOSInAttrs(tok.Attrs) {
				hasOS = true
			}
		}
	}

	return
}

func hasOSInAttrs(attrs []string) bool {
	osKeywords := []string{"windows", "macintosh", "linux", "android", "iphone", "ipad"}

	for _, attr := range attrs {
		a := strings.ToLower(strings.TrimSpace(attr))

		for _, kw := range osKeywords {
			if strings.Contains(a, kw) {
				return true
			}
		}
	}

	return false
}

func isValidChromeVersion(chromeVersion string, hasMozillaPrefix bool) bool {
	if chromeVersion == "" {
		return hasMozillaPrefix
	}

	parts := strings.SplitN(chromeVersion, ".", 2) //nolint:mnd // split major.minor

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	return major >= chromeMinVersion && major <= chromeMaxVersion
}

// Score computes a bot probability score from 0.0 to 1.0.
func Score(f BotFeatures) float64 {
	if f.IsEmpty {
		return scoreEmptyUA
	}

	score := 0.0

	if f.HasBotWord {
		score += weightBotWord
	}

	if f.HasLibSignature {
		score += weightLibSignature
	}

	if f.HasOnlyOneProduct && !f.LooksLikeBrowser {
		score += weightSingleProduct
	}

	if !f.HasConsistentOS {
		score += weightInconsistentOS
	}

	if f.LooksLikeBrowser && !f.ChromeVersionValid {
		score += weightInvalidChrome
	}

	if f.StringEntropy < entropyThreshold {
		score += weightLowEntropy
	}

	return min(score, 1.0)
}

// shannonEntropy computes the Shannon entropy of a string in bits.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[byte]int, freqMapCap)

	for i := range len(s) {
		freq[s[i]]++
	}

	n := float64(len(s))

	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}
