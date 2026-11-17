package music

// clampPage 把分页收进单次请求能承受的范围。
func clampPage(offset, limit int) (int, int) {
	if limit < 1 {
		limit = 5
	}
	if limit > 30 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 300 {
		offset = 300
	}
	return offset, limit
}

// upstreamWindow 把偏移量换成「第几页、页大小、页内跳过几条」。
// 页号从 1 开始。源站页大小不超过 30。
func upstreamWindow(offset, limit int) (page, size, skip int) {
	offset, limit = clampPage(offset, limit)
	if offset == 0 {
		return 1, limit, 0
	}
	if offset+limit <= 30 {
		return 1, offset + limit, offset
	}
	return offset/30 + 1, 30, offset % 30
}

func takeWindow(items []Track, skip, limit int) []Track {
	if skip < 0 {
		skip = 0
	}
	if skip >= len(items) {
		return []Track{}
	}
	items = items[skip:]
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func sliceTracks(items []Track, source string, offset, limit int) []Track {
	if source != "" {
		filtered := make([]Track, 0, len(items))
		for _, item := range items {
			if item.Source == source {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	if offset > len(items) {
		return []Track{}
	}
	if offset > 0 {
		items = items[offset:]
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}
