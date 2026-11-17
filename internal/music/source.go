package music

import "context"

// Remote 是只在服务端使用的临时地址，不会写进搜索结果或博客。
type Remote struct {
	URL string
}

// Resolved 是按 sourceId 重新取到的曲目。添加和试听都走这里，不信任调用方提交的歌名。
type Resolved struct {
	Track Track
	Audio Remote
	Cover string
}

// Source 是音乐来源。搜索和试听都从这里走，页面不直接连接源站。
type Source interface {
	Search(ctx context.Context, query string) ([]Track, error)
	Resolve(ctx context.Context, sourceID string) (Resolved, error)
}

// AudioLocator 只解析试听地址，不再请求曲目详情。
type AudioLocator interface {
	AudioURL(ctx context.Context, sourceID, quality string) (string, error)
}
