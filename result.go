package herald

// ClientType indicates the category of the user agent.
// Each type implies which fields in Result are populated:
//
//	ClientTypeBrowser   — Browser + OS + Device
//	ClientTypeIAB       — Browser + OS + Device + IAB
//	ClientTypeNativeApp — OS + Device + Native
//	ClientTypeHttpClient — Native only
//	ClientTypeBot       — Bot only
//	ClientTypeUnknown   — all fields empty
type ClientType uint8

const (
	ClientTypeBrowser    ClientType = iota // Browser + OS + Device
	ClientTypeIAB                          // Browser + OS + Device + IAB
	ClientTypeNativeApp                    // OS + Device + Native
	ClientTypeHttpClient                   // Native only
	ClientTypeBot                          // Bot only
	ClientTypeUnknown                      // all fields empty
)

// BotCategory classifies what kind of bot this is.
type BotCategory string

const (
	BotCategorySearch  BotCategory = "search"
	BotCategorySocial  BotCategory = "social"
	BotCategoryAI      BotCategory = "ai"
	BotCategoryMonitor BotCategory = "monitor"
	BotCategoryScraper BotCategory = "scraper"
)

// OS name constants.
const (
	OSiOS      = "iOS"
	OSAndroid  = "Android"
	OSWindows  = "Windows"
	OSmacOS    = "macOS"
	OSLinux    = "Linux"
	OSDarwin   = "Darwin"
	OSChromeOS = "Chrome OS"
)

// Device type constants.
const (
	DeviceMobile  = "mobile"
	DeviceTablet  = "tablet"
	DeviceDesktop = "desktop"
)

// Browser engine constants.
const (
	EngineBlink   = "Blink"
	EngineWebKit  = "WebKit"
	EngineGecko   = "Gecko"
	EngineTrident = "Trident"
)

// Well-known product/browser token names.
const (
	ProductCFNetwork = "CFNetwork"
	ProductDarwin    = "Darwin"
	ProductMozilla   = "Mozilla"
)

// Result is the parsed representation of a User-Agent string.
// All fields are value types — use IsEmpty() to check if a section was populated.
type Result struct {
	Raw        string
	ClientType ClientType
	Browser    Browser
	OS         OS
	Device     Device
	IAB        IABInfo
	Bot        BotInfo
	Native     NativeApp
}

// Browser holds parsed browser information.
type Browser struct {
	Name    string // "Chrome", "Safari", "Firefox"
	Version string // "120.0.0.0"
	Engine  string // "Blink", "WebKit", "Gecko"
}

// IsEmpty returns true if no browser was detected.
func (b Browser) IsEmpty() bool { return b.Name == "" }

// OS holds parsed operating system information.
type OS struct {
	Name    string // "iOS", "Android", "Windows", "macOS"
	Version string // "18.7", "14", "11"
}

// IsEmpty returns true if no OS was detected.
func (o OS) IsEmpty() bool { return o.Name == "" }

// Device holds parsed device information.
type Device struct {
	Type     string // "mobile", "tablet", "desktop", "tv", "console"
	Model    string // "iPhone 13 Pro" — may be empty
	ModelRaw string // "iPhone14,2", "SM-G991B" — raw identifier
}

// IsEmpty returns true if no device was detected.
func (d Device) IsEmpty() bool { return d.Type == "" }

// IABInfo holds in-app browser metadata from Facebook/Instagram/TikTok UAs.
type IABInfo struct {
	App         string
	AppVersion  string
	Locale      string
	Region      string
	NetType     string
	ScreenScale float64
	Resolution  string
}

// IsEmpty returns true if no IAB info was detected.
func (i IABInfo) IsEmpty() bool { return i.App == "" }

// BotInfo holds information about a detected bot.
type BotInfo struct {
	Name       string
	Owner      string
	Category   BotCategory
	Confidence float64 // 0.0-1.0
}

// IsEmpty returns true if no bot was detected.
func (b BotInfo) IsEmpty() bool { return b.Name == "" && b.Confidence == 0 }

// NativeApp holds information about a native HTTP client or app.
type NativeApp struct {
	Name    string // "curl", "Dart", "CFNetwork", "ut-1"
	Version string
	Runtime string // "dart:io", "CFNetwork", "OkHttp"
}

// IsEmpty returns true if no native app was detected.
func (n NativeApp) IsEmpty() bool { return n.Name == "" }
