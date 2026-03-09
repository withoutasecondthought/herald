# herald

Fast, zero-dependency Go library for parsing User-Agent strings.

Detects browsers, operating systems, devices, bots, in-app browsers (Facebook, Instagram, TikTok), and native HTTP clients. Supports [Client Hints](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Client_hints) for modern browsers with frozen UA strings.

## Features

- **Browser detection** — Chrome, Safari, Firefox, Edge, Opera, Yandex, Samsung Browser, IE, and more
- **OS detection** — iOS, Android, Windows (NT version mapping), macOS, Linux, Chrome OS, Darwin
- **Device detection** — type (mobile/tablet/desktop) + model resolution via Apple and Android lookup tables
- **Bot detection** — byte-level trie for ~100 known bots + feature scoring for unknown bots
- **In-app browser parsing** — Facebook (`FBAN`), Instagram (positional), Threads (Barcelona), TikTok (`musical_ly`), Meta IAB
- **Client Hints** — `Sec-CH-UA` headers enrich or override frozen UA data
- **Data-driven** — bot patterns and device models embedded as JSON, extensible via `WithOverrides`
- **Zero external dependencies** — stdlib only

## Install

```bash
go get github.com/withoutasecondthought/herald
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/withoutasecondthought/herald"
)

func main() {
    // Zero-config: uses embedded data, no file paths needed.
    p, err := herald.NewParser()
    if err != nil {
        log.Fatal(err)
    }

    ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    r := p.Parse(ua)

    fmt.Println(r.Browser.Name)    // Chrome
    fmt.Println(r.Browser.Version) // 120.0.0.0
    fmt.Println(r.Browser.Engine)  // Blink
    fmt.Println(r.OS.Name)         // Windows
    fmt.Println(r.OS.Version)      // 10
    fmt.Println(r.Device.Type)     // desktop
    fmt.Println(r.ClientType)      // 0 (ClientTypeBrowser)
}
```

## Client Hints

Modern browsers freeze the UA string. Use `ParseWithHints` to get accurate data from `Sec-CH-UA` headers:

```go
hints := herald.ClientHints{
    UA:              `"Chromium";v="120", "Google Chrome";v="120"`,
    Platform:        "Windows",
    PlatformVersion: "15.0.0", // Windows 11
    Mobile:          false,
}

r := p.ParseWithHints(ua, hints)
fmt.Println(r.OS.Version) // 11 (resolved from PlatformVersion)
```

## Fast Type Detection

When you only need to know *what kind* of client it is (bot vs browser vs native app) without full parsing:

```go
switch p.DetectType(ua) {
case herald.ClientTypeBot:
    // block or rate-limit
case herald.ClientTypeBrowser:
    // serve page
case herald.ClientTypeHttpClient:
    // API client
}
```

`DetectType` skips browser/OS/device resolution — bot detection runs in ~160ns with zero allocations.

## Custom Data

By default, `NewParser()` uses built-in data embedded at compile time. You can extend or replace it:

```go
// Add your own bot patterns or device models on top of built-in data.
// The override directory may contain any subset of the data files.
// Only files that exist are merged; missing files are skipped.
p, err := herald.NewParser(herald.WithOverrides("path/to/overrides"))

// Use only your own data files (ignores built-in data entirely).
p, err := herald.NewParser(herald.WithDataDir("path/to/data"))
```

### File Formats

**bots.json** — array of bot patterns (added to the built-in trie):

```json
[
  {"pattern": "MyBot", "name": "My Bot", "owner": "Acme", "category": "scraper"}
]
```

`pattern` is matched as a substring in the UA string (case-sensitive). `category` is one of: `search`, `ai`, `social`, `monitor`, `scraper`, `other`.

**android.json** — map of model identifier to brand and name:

```json
{
  "SM-G991B": {"brand": "Samsung", "model": "Galaxy S21"},
  "Pixel 8":  {"brand": "Google",  "model": "Pixel 8"}
}
```

**apple.json** — map of internal identifier to device name:

```json
{
  "iPhone14,2": "iPhone 13 Pro",
  "iPad13,1":   "iPad Air (4th generation)"
}
```

**darwin.json** — map of Darwin kernel major version to OS versions:

```json
{
  "24": {"ios": "18", "macos": "15"},
  "25": {"ios": "19", "macos": "16"}
}
```

## Package-Level API

For simpler setups, use the package-level functions with a shared default parser:

```go
herald.Init() // uses embedded data

r := herald.Parse(ua)
r := herald.ParseWithHints(ua, hints)
t := herald.DetectType(ua)
```

## Result Types

`Parse` returns a `*Result` with these fields:

```go
type Result struct {
    Raw        string
    ClientType ClientType
    Browser    Browser   // Name, Version, Engine
    OS         OS        // Name, Version
    Device     Device    // Type, Model, ModelRaw
    IAB        IABInfo   // App, AppVersion, Locale, ScreenScale, ...
    Bot        BotInfo   // Name, Owner, Category, Confidence
    Native     NativeApp // Name, Version, Runtime
}
```

All sub-types are values (not pointers) with an `IsEmpty()` method:

```go
if !r.Browser.IsEmpty() {
    fmt.Println(r.Browser.Name)
}
```

### Browser

| Field | Example |
|---|---|
| `Name` | `"Chrome"`, `"Safari"`, `"Firefox"`, `"Edge"` |
| `Version` | `"120.0.0.0"` |
| `Engine` | `"Blink"`, `"WebKit"`, `"Gecko"`, `"Trident"` |

### OS

| Field | Example |
|---|---|
| `Name` | `"iOS"`, `"Android"`, `"Windows"`, `"macOS"`, `"Linux"` |
| `Version` | `"18.7"`, `"14"`, `"10"` |

### Device

| Field | Example |
|---|---|
| `Type` | `"mobile"`, `"tablet"`, `"desktop"` |
| `Model` | `"iPhone 13 Pro"`, `"Galaxy S24 Ultra"` (resolved from DB) |
| `ModelRaw` | `"iPhone14,2"`, `"SM-S928B"` (raw identifier from UA) |

### BotInfo

| Field | Example |
|---|---|
| `Name` | `"Googlebot"`, `"GPTBot"`, `"ClaudeBot"` |
| `Owner` | `"Google"`, `"OpenAI"`, `"Anthropic"` |
| `Category` | `"search"`, `"ai"`, `"social"`, `"monitor"`, `"scraper"` |
| `Confidence` | `1.0` (trie match) or `0.0-1.0` (scoring) |

### IABInfo (In-App Browser)

| Field | Description |
|---|---|
| `App` | App name: `"Facebook"`, `"Instagram"`, `"Threads"`, `"TikTok"`, `"Meta"` |
| `AppVersion` | App version string |
| `Locale` | e.g. `"en_US"` |
| `ScreenScale` | Screen density factor |
| `Resolution` | Screen resolution (Instagram/TikTok) |
| `Region` | Region override (Instagram, Threads, TikTok) |
| `NetType` | Network type (TikTok) |

### NativeApp

| Field | Example |
|---|---|
| `Name` | `"curl"`, `"Dart"`, `"OkHttp"`, `"CFNetwork"` |
| `Version` | `"7.68.0"` |
| `Runtime` | `"dart:io"`, `"OkHttp"`, `"CFNetwork"` |

## Client Types

| ClientType | Populated fields | Example |
|---|---|---|
| `ClientTypeBrowser` | Browser, OS, Device | Chrome, Safari, Firefox |
| `ClientTypeIAB` | Browser, OS, Device, IAB | Facebook app, Instagram app |
| `ClientTypeNativeApp` | Native, OS, Device | CFNetwork, OkHttp, Dart |
| `ClientTypeHttpClient` | Native | curl, wget, python-requests |
| `ClientTypeBot` | Bot | Googlebot, GPTBot |
| `ClientTypeUnknown` | *(none)* | Empty UA string |

## Data Files

The `data/` directory contains JSON databases embedded into the binary at compile time:

| File | Entries | Description |
|---|---|---|
| `bots.json` | ~100 | Known bot patterns with name, owner, and category |
| `apple.json` | ~180 | Apple device identifiers (e.g. `iPhone14,2` -> `iPhone 13 Pro`) |
| `android.json` | ~300 | Android models: Samsung, Google, Xiaomi, OPPO, vivo, OnePlus, Realme, Motorola, Sony, Nothing, Infinix, TECNO, Huawei |
| `darwin.json` | ~13 | Darwin kernel version -> iOS/macOS version mapping |

To add your own entries, use `WithOverrides` to merge a directory of JSON files on top of the built-in data.

## Architecture

**Pipeline order:** empty check -> bot detection (trie, then scoring) -> client type classification -> browser -> OS -> device -> IAB -> Client Hints enrichment.

Bot detection runs first — if a bot is found, all other stages are skipped.

## Performance

Benchmarks on Apple M3 Max, Go 1.26:

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| Parse (Chrome Desktop) | 5,600 | 3,856 | 14 |
| Parse (Safari iOS) | 6,700 | 3,856 | 14 |
| Parse (Googlebot) | 770 | 1,184 | 4 |
| Parse (CFNetwork) | 2,100 | 3,512 | 7 |
| Parse (empty) | 74 | 384 | 1 |
| DetectType (Chrome) | 2,070 | 944 | 6 |
| DetectType (Googlebot) | 159 | 0 | 0 |

Run benchmarks yourself:

```bash
go test -bench=. -benchmem ./...
```

## License

TODO
