# Cadenza

独立于博客的音乐平台。登录后搜索歌曲、试听，并把选中的歌转存到对象存储，再写入博客已有的 `blog_music`。博客后台继续只管理已经入库的音乐。

页面标题是 Cadenza，角落留一行 Crossroads。

## 技术栈

- 服务：Go，HTTP 用标准库 `net/http`。
- 页面：Vue 3 + Vite + TypeScript + `vue-router`。样式写在单文件组件里，图标用内联 SVG。
- 对象存储：腾讯云 COS（`github.com/tencentyun/cos-go-sdk-v5`）。
- MCP：官方 Go SDK，只用于本机 stdio，子命令 `cadenza mcp`。

Node 只在开发机跑 Vite 和打包。上线时 Vue 产物用 `go:embed` 嵌进 Go 程序，服务器只跑这一个进程。

## 目录

```text
cadenza/
  cadenza.md          实现说明，技术选型以本文为准
  cmd/cadenza/        程序入口（网页、REST、音频转发、mcp 子命令）
  internal/music/     搜索、试听、转存、写入博客
  web/                Vue 3 页面
  design/             页面设计稿
```

`cmd` 是入口，`internal/music` 是核心，`web` 是页面。搜索和入库的逻辑只写在核心里。网页、REST、MCP 调用同一套核心。

## 当前进度

`web/` 已有登录页和主页：搜索、结果列表、底部播放条，以及桌面 / 手机两套排版。接口层在 `web/src/api/`，开发期可用 `VITE_API_MODE=mock` 离线演示。Go 服务尚未落地，真实搜索、Range 转发音频和写入博客还没接通。

## 本地开发

页面：

```bash
cd web
npm install
npm run dev
```

默认把 `/api` 代理到 `http://localhost:8080`。没有 Go 进程时：

```bash
# PowerShell
$env:VITE_API_MODE = "mock"
npm run dev
```

类型检查与打包：

```bash
npm run typecheck
npm run build
```

## 行为摘要

1. 搜索只返回曲目信息：歌名、歌手、封面、时长、来源、来源侧 id。播放地址不交给浏览器，也不写入博客。
2. 试听由 Go 向源站取音频再转给浏览器，转发支持 `Range`。
3. 添加时 Go 下载音频和封面，上传到现有 COS，再调用博客添加接口。按歌名加作者查重，已有则提示。
4. Cadenza 不连接博客数据库，只调用博客已有的添加接口。

凭证放在环境变量里，页面不出现 COS 密钥、博客令牌和音乐源密钥。网页登录后使用 Cookie；REST 和 MCP 共用访问令牌。

完整约束见 [cadenza.md](cadenza.md)。给后续实现用的约定见 [AGENTS.md](AGENTS.md)。
