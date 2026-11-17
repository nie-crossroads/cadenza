package music

import (
	"strings"
	"unicode/utf16"
)

const (
	maxNameRunes   = 20
	sourceNetease  = "网易"
	sourceQQ       = "ＱＱ"
	sourceKugou    = "酷狗"
	sourceKuwo     = "酷我"
	sourceQianqian = "千千"
	sourceYiting   = "一听"
	sourceMigu     = "咪咕"
)

// Track 是搜索结果。不含播放地址。
type Track struct {
	Name     string `json:"name"`
	Author   string `json:"author"`
	Cover    string `json:"cover"`
	Duration int    `json:"duration"`
	Source   string `json:"source"`
	SourceID string `json:"sourceId"`
	Quality  string `json:"quality"`
}

// UTF16Len 与博客侧 Java String.length 一致。
func UTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func Seconds(ms int64) int {
	if ms <= 0 {
		return 0
	}
	return int((ms + 500) / 1000)
}

// ChooseAuthor 优先保留能放进 varchar(20) 的作者。超长时只用第一位歌手。
func ChooseAuthor(names []string) string {
	clean := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			clean = append(clean, name)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	joined := strings.Join(clean, "、")
	if UTF16Len(joined) <= maxNameRunes {
		return joined
	}
	return clean[0]
}

func QualityOf(playLevel, maxLevel string, maxBR int, preferMax bool) string {
	level := strings.ToLower(strings.TrimSpace(playLevel))
	maxLevel = strings.ToLower(strings.TrimSpace(maxLevel))
	if preferMax || level == "" || level == "none" {
		if maxLevel != "" && maxLevel != "none" {
			level = maxLevel
		}
	}
	switch level {
	case "hires", "jymaster", "sky":
		return "hr"
	case "lossless", "exhigh", "higher", "jyeffect":
		return "hi"
	case "standard":
		return "std"
	}
	switch {
	case maxBR >= 1_500_000:
		return "hr"
	case maxBR >= 192_000:
		return "hi"
	default:
		return "std"
	}
}

func normalizeQuality(q string) string {
	switch q {
	case "hr", "hi":
		return q
	default:
		return "std"
	}
}

// CanonicalQuality 只接受空、std、hi、hr。空按标准。
func CanonicalQuality(raw string) (string, bool) {
	switch strings.TrimSpace(raw) {
	case "", "std":
		return "std", true
	case "hi", "hr":
		return strings.TrimSpace(raw), true
	default:
		return "", false
	}
}
