package main

import (
	"context"
	"errors"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nie-crossroads/cadenza/internal/config"
	"github.com/nie-crossroads/cadenza/internal/music"
)

func serveMCP(cfg config.Config, svc *music.Service) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "cadenza", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "按歌名或歌手搜索曲目。只返回歌名、歌手、封面、时长、来源、来源侧 id 和音质，不含播放地址。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Query string `json:"query" jsonschema:"歌名或歌手"`
	}) (*mcp.CallToolResult, struct {
		Items []music.Track `json:"items"`
	}, error) {
		items, err := svc.Search(ctx, in.Query)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, struct {
					Items []music.Track `json:"items"`
				}{}, err
			}
			return nil, struct {
				Items []music.Track `json:"items"`
			}{}, errors.New(music.CodeSearchFailed)
		}
		return nil, struct {
			Items []music.Track `json:"items"`
		}{Items: items}, nil
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add",
		Description: "按搜索结果里的 sourceId 下载歌曲和封面，转存后写入博客。quality 取 std、hi、hr，省略时按标准音质。已存在时返回 ADD_DUPLICATE。不要传入播放地址或密钥。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		SourceID string `json:"sourceId" jsonschema:"搜索结果中的 sourceId"`
		Quality  string `json:"quality,omitempty" jsonschema:"音质 std、hi 或 hr，省略为 std"`
	}) (*mcp.CallToolResult, struct {
		ID int64 `json:"id,omitempty"`
	}, error) {
		if cfg.AccessToken == "" {
			log.Printf("添加失败：访问令牌未配置")
			return nil, struct {
				ID int64 `json:"id,omitempty"`
			}{}, errors.New(music.CodeUnknown)
		}
		id, err := svc.Add(ctx, in.SourceID, in.Quality)
		if err != nil {
			if errors.Is(err, music.ErrNotFound) {
				return nil, struct {
					ID int64 `json:"id,omitempty"`
				}{}, errors.New(music.CodeDownloadFailed)
			}
			var f *music.Fail
			if errors.As(err, &f) && f.Code != "" {
				return nil, struct {
					ID int64 `json:"id,omitempty"`
				}{}, errors.New(f.Code)
			}
			return nil, struct {
				ID int64 `json:"id,omitempty"`
			}{}, errors.New(music.CodeUnknown)
		}
		return nil, struct {
			ID int64 `json:"id,omitempty"`
		}{ID: id}, nil
	})
	return server.Run(context.Background(), &mcp.StdioTransport{})
}
