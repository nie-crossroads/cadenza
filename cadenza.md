# Cadenza

给后续 agent 的实现说明。选型已经定过，按本文做，不要再改技术栈或拆仓库。

Cadenza 是独立于博客的音乐平台：搜索歌曲、列出结果、试听，并把选中的歌转存后写入博客 `blog_music`。博客后台继续只管理已经入库的音乐。仓库名 `cadenza`，和 `agent-blog` 分开。

页面标题写 Cadenza，角落留一行 Crossroads。

## 技术栈

- 服务：Go 1.27（一个安装包同时是编译器和依赖管理，没有 JDK / Maven）。HTTP 用标准库 `net/http`，不用 Gin、Echo。
- 页面：Vue 3 + Vite + TypeScript + `vue-router`。不上 Element Plus，不上其他 UI 组件库，不上 Pinia。图标用少量内联 SVG。样式写在单文件组件里。
- 对象存储：腾讯云 COS 官方 Go SDK `github.com/tencentyun/cos-go-sdk-v5`。
- MCP：官方 Go SDK `github.com/modelcontextprotocol/go-sdk`，只用于本地 stdio。
- Node 只在开发机用来跑 Vite 和打包。服务器不装 Node。

第一版不做 gRPC，不自带数据库，不接歌词。

## 进程

开发时两个进程：Vite 出页面，Go 出接口。Vite 把接口代理到 Go。

上线时把 Vue 打成静态文件，用 `go:embed` 嵌进 Go 程序。服务器只跑这一个进程，页面和接口都从它来。

## 目录

一个 Git 仓库里的目录，不是多个仓库。

```text
cadenza/
  cmd/cadenza/      程序入口。网页、REST、音频转发，以及 mcp 子命令
  internal/music/   搜索、试听、转存、写入博客。三扇门都调用这里
  web/              Vue 3 页面
```

`cmd` 是门，`internal/music` 是核心，`web` 是页面。搜索和入库的逻辑只写在核心里。

## 三扇门

三扇门调用同一套核心。

| 门 | 谁用 | 能做什么 |
|---|---|---|
| 网页 | 人登录后操作 | 搜索、列表、试听、添加 |
| REST | 其他程序 | 搜索、添加 |
| MCP | Cursor 里的 Agent，本机 stdio | 搜索、添加 |

MCP 用子命令 `cadenza mcp`。REST 和 MCP 不带播放器。

## 页面

两个路由：登录、主页。

主页有搜索框、结果列表、底部播放条。桌面端列表在上方、播放条在底部。手机端列表占满宽度，播放条固定在底部。同一套页面的两种排版，不做 App。

播放用浏览器原生音频。音频地址指向 Go 的转发接口，浏览器不直接连接音乐源站。

## 核心行为

音乐来源做成接口，第一版只实现一个。密钥放环境变量。

1. 搜索只返回曲目信息：歌名、歌手、封面、时长、来源、来源侧 id。不把源站播放地址交给浏览器，也不写入博客。
2. 试听由 Go 向源站取音频，再转给浏览器。转发必须支持 `Range`，进度条才能拖动。
3. 添加时 Go 下载音频和封面，上传到现有 COS，拿到短地址后调用博客的添加接口。限制文件大小和超时。下载或上传失败时，把原因返回给这一次操作。
4. 按歌名加作者查重，已有则提示，不重复插入。

源站地址又长又会过期，不能存进 `blog_music`。

## 写入博客

Cadenza 不连接博客数据库。

博客侧已经提供添加接口，内部调用 `MusicService.save`。Cadenza 只调这个接口，不要再改表。

`POST /agent-blog/server/api/cadenza/music`

请求头 `Authorization: Bearer <博客令牌>`。令牌配置在博客的 `blog.cadenza.token`（本地写在 `application-local.yml`，不要提交）。请求体 JSON：

```json
{ "name": "歌名", "author": "作者", "url": "播放地址", "cover": "封面地址", "status": 1 }
```

`author`、`cover`、`status` 可以不传。`status` 只接受 `0` 或 `1`，不传则前台显示。不接受 `lrc`，写入时歌词为空。成功 HTTP 200，body 是新记录的 id。失败 HTTP 4xx，body 为 `{ code, message }`。

| 情况 | HTTP | message |
|---|---|---|
| 令牌缺失、错误，或服务端未配置令牌 | 401 | 音乐平台令牌无效 |
| 没填歌名 | 400 | 请填写歌名 |
| 歌名超过 20 | 400 | 歌名不能超过20个字符 |
| 作者超过 20 | 400 | 作者不能超过20个字符 |
| 没填播放地址 | 400 | 请填写播放地址 |
| 播放地址超过 80 | 400 | 播放地址不能超过80个字符 |
| 封面地址超过 100 | 400 | 封面地址不能超过100个字符 |
| 同名同作者已存在 | 409 | 这首歌已添加 |

列长度：`name` varchar(20)、`author` varchar(20)、`url` varchar(80)、`cover` varchar(100)、`lrc` varchar(50)。COS 地址必须放得进 `url` 和 `cover`。

人在博客后台手动填写地址的表单保留。

## 凭证

三套凭证分开，页面里不出现 COS 密钥、博客令牌和音乐源密钥。

- 网页登录：环境变量里的口令，登录后使用 Cookie。
- REST 和 MCP：共用一个访问令牌。添加必须带令牌。网页上的添加必须先登录。
- 音乐源、COS、博客接口各自的密钥只在 Go 进程的环境变量里。

环境变量至少包括：网页口令、访问令牌、COS 的密钥 / 桶 / 地域 / 访问域名、博客接口地址、博客令牌、音乐源所需的密钥。

## 不要做的事

- 不要把 `mp3.ltyuanfang.cn` 或任何聚合站嵌进页面，也不要让浏览器直接请求源站。
- 不要把临时播放链接写进 `blog_music`。
- 不要让 Cadenza 直连博客的 MySQL。
- 不要改 `blog_music` 的表结构。
- 不要在第一版做 gRPC、歌词、多来源、App。
- 不要为了这个平台引入 UI 组件库或第二个线上进程。
