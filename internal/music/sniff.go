package music

import "bytes"

// playablePrefix 判断这一小段是不是音频。网页和 JSON 不算能播放。
func playablePrefix(data []byte, contentType string) bool {
	if len(data) == 0 {
		return false
	}
	ct := bytes.ToLower([]byte(contentType))
	if bytes.Contains(ct, []byte("text/html")) || bytes.Contains(ct, []byte("application/json")) || bytes.Contains(ct, []byte("text/plain")) {
		return false
	}
	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) > 0 && (trimmed[0] == '<' || trimmed[0] == '{') {
		return false
	}
	if bytes.HasPrefix(data, []byte("fLaC")) || bytes.HasPrefix(data, []byte("ID3")) || bytes.HasPrefix(data, []byte("OggS")) {
		return true
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1]&0xE0 == 0xE0 {
		return true
	}
	if len(data) >= 8 && bytes.Equal(data[4:8], []byte("ftyp")) {
		return true
	}
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WAVE")) {
		return true
	}
	return bytes.HasPrefix(ct, []byte("audio/"))
}

func sniffAudio(data []byte, declared string) (ext, contentType string) {
	if bytes.HasPrefix(data, []byte("fLaC")) {
		return ".flac", "audio/flac"
	}
	if bytes.HasPrefix(data, []byte("ID3")) || (len(data) >= 2 && data[0] == 0xFF && data[1]&0xE0 == 0xE0) {
		return ".mp3", "audio/mpeg"
	}
	if stringsHas(declared, "flac") {
		return ".flac", "audio/flac"
	}
	return ".mp3", "audio/mpeg"
}

func sniffCover(data []byte, declared string) (ext, contentType string) {
	switch {
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}):
		return ".png", "image/png"
	case len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return ".webp", "image/webp"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return ".gif", "image/gif"
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return ".jpg", "image/jpeg"
	}
	switch {
	case stringsHas(declared, "png"):
		return ".png", "image/png"
	case stringsHas(declared, "webp"):
		return ".webp", "image/webp"
	case stringsHas(declared, "gif"):
		return ".gif", "image/gif"
	default:
		return ".jpg", "image/jpeg"
	}
}

func stringsHas(s, part string) bool {
	return len(s) > 0 && len(part) > 0 && (len(s) >= len(part)) && (bytes.Contains(bytes.ToLower([]byte(s)), []byte(part)))
}
