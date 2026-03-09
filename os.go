package herald

import (
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

//nolint:gochecknoglobals // read-only lookup table
var windowsVersions = map[string]string{
	"10.0": "10",
	"6.3":  "8.1",
	"6.2":  "8",
	"6.1":  "7",
	"6.0":  "Vista",
	"5.1":  "XP",
	"5.0":  "2000",
}

const macOSXPrefix = "Mac OS X "

// detectOS parses OS information from comment tokens and product tokens.
func detectOS(tokens []tokenizer.Token, result *Result, database *db.Database) {
	if detectOSFromComments(tokens, result) {
		return
	}

	detectOSFromDarwin(tokens, result, database)
}

func detectOSFromComments(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindComment {
			continue
		}

		for _, attr := range tok.Attrs {
			attr = strings.TrimSpace(attr)

			if found := matchOSAttr(attr, tok.Attrs, result); found {
				return true
			}
		}
	}

	return false
}

func matchOSAttr(attr string, allAttrs []string, result *Result) bool {
	if strings.Contains(attr, "CPU iPhone OS") || strings.Contains(attr, "CPU OS") {
		if v := parseIOSVersion(attr); v != "" {
			result.OS = OS{Name: OSiOS, Version: v}

			return true
		}
	}

	if v, ok := strings.CutPrefix(attr, OSAndroid); ok {
		result.OS = OS{Name: OSAndroid, Version: strings.TrimSpace(v)}

		return true
	}

	if ntVer, ok := strings.CutPrefix(attr, "Windows NT"); ok {
		ntVer = strings.TrimSpace(ntVer)

		winVer := ntVer
		if mapped, ok := windowsVersions[ntVer]; ok {
			winVer = mapped
		}

		result.OS = OS{Name: OSWindows, Version: winVer}

		return true
	}

	if strings.HasPrefix(attr, "Macintosh") {
		result.OS = OS{Name: OSmacOS}

		for _, a2 := range allAttrs {
			a2 = strings.TrimSpace(a2)

			if v := parseMacOSVersion(a2); v != "" {
				result.OS.Version = v
			}
		}

		return true
	}

	if strings.HasPrefix(attr, OSLinux) && result.OS.Name == "" {
		result.OS = OS{Name: OSLinux}
	}

	return false
}

func detectOSFromDarwin(tokens []tokenizer.Token, result *Result, database *db.Database) {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct || tok.Name != ProductDarwin {
			continue
		}

		if v := tok.Version; v != "" {
			major := strings.SplitN(v, ".", 2)[0] //nolint:mnd // split major.minor

			if dv, ok := database.DarwinMap[major]; ok {
				result.OS = OS{Name: OSDarwin, Version: dv.IOS}
			}
		}

		return
	}
}

func parseIOSVersion(s string) string {
	_, rest, found := strings.Cut(s, "OS ")
	if !found {
		return ""
	}

	if likeIdx := strings.Index(rest, " like"); likeIdx >= 0 {
		rest = rest[:likeIdx]
	}

	rest = strings.TrimSpace(rest)

	return strings.ReplaceAll(rest, "_", ".")
}

func parseMacOSVersion(s string) string {
	_, v, found := strings.Cut(s, macOSXPrefix)
	if !found {
		return ""
	}

	v = strings.TrimSpace(v)

	return strings.ReplaceAll(v, "_", ".")
}
