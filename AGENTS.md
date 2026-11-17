# Cadenza

独立于博客的音乐平台，仓库名 `cadenza`，和 `agent-blog` 分开。选型已定，按 [cadenza.md](cadenza.md) 做。

## 技术栈

- Go：标准库 `net/http`。不上 Gin、Echo，不上 gRPC，不自带数据库。
- 页面：Vue 3 + Vite + TypeScript + `vue-router`。不上 Element Plus 和其他 UI 组件库，不上 Pinia。图标用内联 SVG，样式写在单文件组件里。
- COS：`github.com/tencentyun/cos-go-sdk-v5`。
- MCP：`github.com/modelcontextprotocol/go-sdk`，只用于本地 stdio，子命令 `cadenza mcp`。
- Node 只在开发机跑 Vite 和打包。上线把 `web/dist` 用 `go:embed` 嵌进同一个 Go 进程。

第一版不接歌词，只做一个音乐来源。

## 目录

```text
cmd/cadenza/      入口：网页、REST、音频转发、mcp
internal/music/   搜索、试听、转存、写入博客。三扇门都调用这里
web/              Vue 3 页面
design/           设计稿，实现时以 cadenza.md 和现有 web/ 为准
```

搜索和入库只写在 `internal/music`。`web/src/api/` 里的契约是草案，和 Go 冻结前保持字段一致，不要在页面里另写一套请求。

## 三扇门

| 门 | 谁用 | 能做什么 |
|---|---|---|
| 网页 | 人登录后 | 搜索、列表、试听、添加 |
| REST | 其他程序 | 搜索、添加 |
| MCP | 本机 Agent | 搜索、添加 |

REST 和 MCP 不带播放器。

## 页面

路由只有登录和主页。桌面端列表在上、播放条在底；手机端列表占满宽度，播放条固定在底部。同一套页面，不做 App。

播放用浏览器原生音频。`audio.src` 只指向 Go 的转发接口，浏览器不连接音乐源站。

文案和错误码白名单集中在 `web/src/copy/messages.ts`。页面只渲染这里的中文，不渲染后端 `message` 原文。错误码本身不进 DOM、属性、Toast、标题或日志。

`Track` 有 `name`、`author`、`cover`、`duration`、`source`、`sourceId`、`quality`。`quality` 只取 `hi` / `std` / `hr`，页面映射成高品、标准、Hi-Res。`sourceId` 只留在内存和 `data-source-id`，不显示成文本，不进 URL。

添加按钮按行维护状态。`409` 表示已存在，只提示这一行。只有会话失效才退出登录；博客侧鉴权失败算这一行添加失败。

页面标题写 Cadenza，角落留一行 Crossroads。

## 核心行为

1. 搜索只返回曲目信息，不把源站播放地址交给浏览器，也不写入博客。
2. 试听由 Go 向源站取音频再转发，必须支持 `Range`。
3. 添加时 Go 下载音频和封面，上传到现有 COS，拿到短地址后调用博客添加接口。限制文件大小和超时。失败把原因返回给这一次操作。
4. 按歌名加作者查重，已有则提示，不重复插入。

源站地址会过期，不能存进 `blog_music`。COS 地址必须放进 `url` varchar(80) 和 `cover` varchar(100)。歌名、作者各不超过 20 个字符。

## 写入博客

Cadenza 不连接博客数据库，不改 `blog_music` 表。只调用：

`POST /agent-blog/server/api/cadenza/music`

请求头 `Authorization: Bearer <博客令牌>`。请求体：`name`、`author`、`url`、`cover`、`status`。不传 `lrc`。成功 HTTP 200，body 是新记录 id。同名同作者返回 409。

## 凭证

三套凭证分开，页面、DOM 和日志里不出现 COS 密钥、博客令牌、音乐源密钥。

- 网页：环境变量里的口令，登录后 HttpOnly Cookie。前端不读写 Cookie，不设置认证请求头。
- REST 和 MCP：共用一个访问令牌。添加必须带令牌。网页上的添加必须先登录。
- 音乐源、COS、博客接口的密钥只在 Go 进程的环境变量里。

## 本地页面

```bash
cd web
npm install
npm run dev          # /api 代理到 http://localhost:8080
npm run typecheck
npm run build
```

没有 Go 进程时设置 `VITE_API_MODE=mock` 再 `npm run dev`。改完页面后跑 `npm run build`。

## 提交

用户明确要求提交时才 `git commit`。信息只写这次改动涉及的类型，一行一个，类型和说明之间用全角冒号 `：`。没涉及的类型不写，不加空行占位。说明用中文，写做了什么。

```text
refactor：重构
docs：优化文档
chore：升级 SDK、移除废弃代码
style：样式调整
fix：bug 修复
feat：增加一个功能
perf：优化代码
```

只改了文档时：

```text
docs：补充提交说明
```

同一次提交里有多项时，只保留对应行：

```text
feat：增加歌曲搜索
fix：修正空关键词报错
```

author 和 committer 都使用本机 Git 已配置的作者，当前是 `打酱油的大白菜 <1762546812@qq.com>`。不要传 `--author`，不要设置 `GIT_AUTHOR_NAME`、`GIT_AUTHOR_EMAIL`、`GIT_COMMITTER_NAME`、`GIT_COMMITTER_EMAIL`。

信息末尾不要加 `Co-authored-by`、`Made-with`、`Generated-by`。不要出现 Cursor、cursoragent、Workbuddy、workbuddyagent 或其他 agent 的名字和邮箱。GitHub Contributors 里只保留这位作者。

`scripts/git-hooks/commit-msg` 会在提交时去掉这些署名行，并在 author 或 committer 不是上述作者时拒绝提交。不要用 `--no-verify` 跳过它。

不修改 git config。

## 不要做

- 不要改技术栈，不要拆仓库，不要为这个平台再引入 UI 组件库或第二个线上进程。
- 不要把聚合站嵌进页面，不要让浏览器直接请求源站。
- 不要把临时播放链接写进 `blog_music`。
- 不要让 Cadenza 直连博客的 MySQL，不要改表结构。
- 不要在第一版做 gRPC、歌词、多来源、App。
- 不要提交 `.env`、本地 yml、密钥，也不要主动 `git commit` / `git push`，除非用户明确要求。
- 不要把 Cursor、Workbuddy 或其他 agent 写进提交作者、提交者或 `Co-authored-by`。
