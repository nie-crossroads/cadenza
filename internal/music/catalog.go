package music

import (
	"context"
	"log"
	"sync"
)

// catalogSource 是目录里的一个来源。Label 与页面上的渠道名一致。
type catalogSource interface {
	Source
	Label() string
	SearchPage(ctx context.Context, query string, offset, limit int) ([]Track, error)
}

// Catalog 并行搜索各个来源，并按 sourceId 前缀把试听和添加分回去。
type Catalog struct {
	sources []catalogSource
}

func NewCatalog(cookie string) *Catalog {
	client := sourceClient()
	return &Catalog{sources: []catalogSource{
		newNetease(neteaseBase, cookie, client),
		newQQ(client),
		newKugou(client),
		newKuwo(client),
		newQianqian(client),
		newYiting(client),
		newMigu(client),
	}}
}

func (c *Catalog) Search(ctx context.Context, query string) ([]Track, error) {
	return c.SearchFiltered(ctx, query, "", 0, 30)
}

func (c *Catalog) SearchFiltered(ctx context.Context, query, source string, offset, limit int) ([]Track, error) {
	offset, limit = clampPage(offset, limit)
	selected := make([]catalogSource, 0, len(c.sources))
	for _, src := range c.sources {
		if source != "" && src.Label() != source {
			continue
		}
		selected = append(selected, src)
	}
	if len(selected) == 0 {
		return []Track{}, nil
	}
	found := make([][]Track, len(selected))
	var wg sync.WaitGroup
	var mu sync.Mutex
	okN := 0
	failN := 0
	for i, src := range selected {
		wg.Add(1)
		go func(i int, src catalogSource) {
			defer wg.Done()
			items, err := src.SearchPage(ctx, query, offset, limit)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if ctx.Err() == nil {
					failN++
					log.Printf("来源搜索失败")
				}
				return
			}
			okN++
			found[i] = items
		}(i, src)
	}
	wg.Wait()
	if okN == 0 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if failN > 0 {
			return nil, errSource
		}
		return []Track{}, nil
	}
	out := make([]Track, 0)
	for _, items := range found {
		out = append(out, items...)
	}
	return out, nil
}

func (c *Catalog) Resolve(ctx context.Context, sourceID string) (Resolved, error) {
	src := c.match(sourceID)
	if src == nil {
		return Resolved{}, ErrNotFound
	}
	return src.Resolve(ctx, sourceID)
}

func (c *Catalog) AudioURL(ctx context.Context, sourceID, quality string) (string, error) {
	src := c.match(sourceID)
	if src == nil {
		return "", ErrNotFound
	}
	if loc, ok := src.(AudioLocator); ok {
		return loc.AudioURL(ctx, sourceID, quality)
	}
	resolved, err := src.Resolve(ctx, sourceID)
	if err != nil {
		return "", err
	}
	return resolved.Audio.URL, nil
}

func (c *Catalog) match(sourceID string) catalogSource {
	for _, src := range c.sources {
		if owns(src, sourceID) {
			return src
		}
	}
	return nil
}

func owns(src catalogSource, sourceID string) bool {
	switch src.Label() {
	case sourceNetease:
		return len(sourceID) > len(neteasePrefix) && sourceID[:len(neteasePrefix)] == neteasePrefix
	case sourceQQ:
		return len(sourceID) > len(qqPrefix) && sourceID[:len(qqPrefix)] == qqPrefix
	case sourceKugou:
		return len(sourceID) > len(kugouPrefix) && sourceID[:len(kugouPrefix)] == kugouPrefix
	case sourceKuwo:
		return len(sourceID) > len(kuwoPrefix) && sourceID[:len(kuwoPrefix)] == kuwoPrefix
	case sourceQianqian:
		return len(sourceID) > len(qianqianPrefix) && sourceID[:len(qianqianPrefix)] == qianqianPrefix
	case sourceYiting:
		return len(sourceID) > len(yitingPrefix) && sourceID[:len(yitingPrefix)] == yitingPrefix
	case sourceMigu:
		return len(sourceID) > len(miguPrefix) && sourceID[:len(miguPrefix)] == miguPrefix
	default:
		return false
	}
}
