package herald

// ClientHints represents the Client Hints headers sent by modern browsers
// as a replacement for the frozen User-Agent string.
// These values take priority over data parsed from the UA string.
type ClientHints struct {
	UA              string // Sec-CH-UA: "Chromium";v="120", "Google Chrome";v="120"
	Mobile          bool   // Sec-CH-UA-Mobile: ?0
	Platform        string // Sec-CH-UA-Platform: "Windows"
	PlatformVersion string // Sec-CH-UA-Platform-Version: "15.0.0"
	FullVersionList string // Sec-CH-UA-Full-Version-List
	Model           string // Sec-CH-UA-Model
	Architecture    string // Sec-CH-UA-Arch
}

func (ch ClientHints) IsEmpty() bool {
	return ch.UA == "" && ch.Platform == "" && ch.Model == ""
}
