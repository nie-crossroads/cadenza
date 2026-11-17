/**
 * ============================================================
 * Cadenza 全部中文文案 + 错误码白名单映射
 * ------------------------------------------------------------
 * 审计约定（务必遵守）：
 *  1. 任何渲染到页面 / aria-label / aria-valuetext / title / Toast /
 *     document.title / 控制台日志的文本，必须来自本文件。
 *  2. 绝不渲染后端 message 原文：api 层只回传白名单错误码，本文件负责映射。
 *  3. 错误码本身绝不写进 DOM、属性、Toast、title 或日志。
 *  4. 任何涉及后端凭证或服务端内部细节的措辞都不允许出现在用户可见文案里。
 * ============================================================
 */

export const BRAND = {
  name: 'Cadenza',
  corner: 'Crossroads',
  fermataAlt: '延长记号装饰',
  crosshairAlt: '十字准星装饰',
} as const

/** 页面 / 表单 / 状态文案 */
export const COPY = {
  /* 登录 */
  loginPrompt: '请输入口令',
  loginAccess: '访问口令',
  loginEnter: '进入',
  loginSub: '输入访问口令，进入你的谱架。',
  loginHint: '登录后可以把歌曲收进博客。',
  coverLead: '搜索、试听、转存。把散落在各个平台的一首曲子，收进自己的谱架。',
  loginSubmit: '登录',
  loginPending: '登录中…',
  loginInvalid: '口令不正确，请重试',
  loginUnavailable: '暂时无法登录，请稍后重试',
  loginExpired: '登录已失效，请重新登录',

  /* 搜索 */
  searchPlaceholder: '输入歌名或歌手',
  clearSearch: '清空搜索',
  emptyTitle: '想听什么？',
  emptySub: '搜歌名或歌手。结果按来源分组，点一行即可试听。',
  searchGo: '搜索',
  themeToLight: '浅色模式',
  themeToDark: '深色模式',
  themeSettings: '显示设置',
  themeColor: '主题色',
  recentLabel: '最近搜索',
  allSources: '全部',
  pickSources: '选择来源',
  sourcePickOk: '确定',
  sourcePickCancel: '取消',
  sourceNames: ['网易', 'ＱＱ', '酷狗', '酷我', '千千', '一听', '咪咕'],
  defaultSources: ['网易', 'ＱＱ', '酷狗'],
  loggedIn: '已登录',
  logout: '退出',
  groupFold: '收起',
  groupExpand: '展开',
  pickSource: '选源',
  mergedGroup: '同曲合并',
  mergedTag: '多来源',
  nowPlaying: '正在试听',
  addToBlog: '添加到博客',
  saveToBlog: '转存到博客',
  closePlayer: '关闭播放页',
  proxyNote: '音频经本地服务转发，浏览器不直接连接音乐平台。',
  qualityHi: '高品',
  qualityStd: '标准',
  qualityHr: 'Hi-Res',
  sourcesHere: (n: number): string => `此曲在 ${n} 个来源`,
  emptyFocus: '聚焦搜索框',
  noResultsTitle: '没有找到「{kw}」相关的歌曲',
  noResultsSub: '换个关键词，或检查拼写再试试',
  searching: '正在搜索',
  found: (n: number): string => `找到 ${n} 首歌曲`,
  searchFailedTitle: '搜索失败',
  searchFailedSub: '网络连接失败，请检查网络后重试',
  loadMore: '加载更多',
  loadMorePending: '加载中…',
  sourceEmpty: '这个来源没有歌曲',
  retry: '重试',

  /* 试听 / 播放条 */
  previewEmpty: '还没有试听的歌曲',
  previewFailed: '试听失败，点击重试',
  previewToast: (name: string): string => `「${name}」暂时不能试听`,
  previewToastDetail: '这个来源没有提供可播放的地址',
  rangeNotice: '当前不支持拖动进度',
  statusIdle: '试听',
  auditioning: '试听中',
  /* 播放状态静态文本统一三值「播放中 / 已暂停 / 缓冲中」，
     reduced-motion 静态文本、波形 sr-only、播放条状态区全部复用这一套。 */
  statusPlaying: '播放中',
  statusPaused: '已暂停',
  statusBuffering: '缓冲中',
  statusError: '试听失败',
  statusEnded: '播放结束',

  /* 添加五态 */
  addIdle: '添加',
  addQualityTitle: '选择音质',
  addQualityHint: (name: string): string => `为「${name}」选择要保存的音质`,
  addQualityNext: '下一步',
  addPlayableTitle: '检查音质',
  addPlayableChecking: '正在确认所选音质能否播放',
  addUnplayableTitle: '无法添加',
  addUnplayableBody: '该歌曲无法播放，不能添加到网站中',
  addUnplayableOk: '知道了',
  addConfirmTitle: '确认添加',
  addConfirmBody: (name: string, tone: string): string => `把「${name}」以${tone}添加到博客？`,
  addConfirmOk: '确认添加',
  addConfirmCancel: '取消',
  addPending: '添加中…',
  addSuccess: '已添加',
  addDuplicate: '已存在',
  addError: '重试',
  addDuplicateNote: '这首歌已在博客里，无需重复添加',
  toastAdded: '已添加到博客',
  toastAddFailed: '添加失败',

  /* 列表标签 */
  durationLabel: '时长',
  sourceLabel: '来源',
  resultsLabel: '搜索结果',
  skeletonLabel: '正在加载搜索结果',
} as const

/** aria / 播报文案（函数名表示需要拼接动态值） */
export const A11Y = {
  homeMain: 'Cadenza 主页',
  homeAnnouncement: 'Cadenza 主页',
  loginMain: 'Cadenza 登录',
  searchLabel: '搜索歌曲',
  searchInputDesc: '按回车立即搜索，按 Esc 清空',
  showPassword: '显示口令',
  hidePassword: '隐藏口令',
  passwordLabel: '口令',
  play: '播放',
  pause: '暂停',
  prev: '上一首',
  next: '下一首',
  openPlayer: '展开播放页',
  switchSource: (name: string): string => `换到 ${name}`,
  mute: '静音',
  unmute: '取消静音',
  volume: '音量',
  progress: '播放进度',
  resultsCount: (n: number): string => `共 ${n} 条结果`,
  audition: (name: string): string => `试听 ${name}`,
  pauseAudition: (name: string): string => `暂停 ${name}`,
  resumeAudition: (name: string): string => `继续试听 ${name}`,
  retryAudition: (name: string): string => `重试试听 ${name}`,
  addTrack: (name: string): string => `添加 ${name}`,
  retryAdd: (name: string): string => `重试添加 ${name}`,
  timecode: (text: string): string => `播放进度 ${text}`,
  volumeValue: (v: number): string => `音量 ${v}%`,
  playingRow: (name: string): string => `正在试听 ${name}`,
  eqLamp: (text: string): string => `播放状态指示：${text || '未在播放'}`,
  closeToast: '关闭提示',
  emptyVinyl: '黑胶唱片装饰',
} as const

/** 音质码 → 页面标签。未知码落到标准，不渲染原文。 */
export function qualityLabel(code: string): string {
  if (code === 'hi') return COPY.qualityHi
  if (code === 'hr') return COPY.qualityHr
  return COPY.qualityStd
}

/** 添加时可选的档位：不超过这首歌标出的最高音质。 */
export function qualityChoices(code: string): Array<'std' | 'hi' | 'hr'> {
  if (code === 'hr') return ['std', 'hi', 'hr']
  if (code === 'hi') return ['std', 'hi']
  return ['std']
}

/** 模板占位符替换（只用白名单模板，不拼接后端字符串） */
export function fill(template: string, values: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (whole, key: string) =>
    Object.prototype.hasOwnProperty.call(values, key) ? values[key] : whole,
  )
}

/**
 * 错误码白名单 → 用户可见文案。
 * 未知码 / 后端原样字符串一律落到通用兜底，绝不透传。
 */
export const ERROR_MESSAGES: Readonly<Record<string, string>> = Object.freeze({
  /* 添加失败原因（D 节白名单，逐条固定） */
  ADD_DOWNLOAD_FAILED: '音频下载失败，请稍后重试',
  ADD_COS_UPLOAD_FAILED: '文件转存失败，请稍后重试',
  ADD_FILE_TOO_LARGE: '文件过大，无法添加',
  ADD_TIMEOUT: '添加超时，请重试',
  ADD_BLOG_TOKEN_INVALID: '暂时无法写入博客，请稍后重试',
  NETWORK_ERROR: '网络连接失败，请检查网络后重试',
  BLOG_VALIDATION_FAILED: '歌曲信息不符合博客要求，无法添加',
  ADD_UNKNOWN: '添加失败，请重试',

  /* 登录 / 会话 / 搜索 / 试听 */
  LOGIN_INVALID: COPY.loginInvalid,
  LOGIN_UNAVAILABLE: COPY.loginUnavailable,
  SESSION_INVALID: COPY.loginExpired,
  UNAUTHENTICATED: COPY.loginExpired,
  SEARCH_FAILED: COPY.searchFailedSub,
  PREVIEW_FAILED: COPY.previewFailed,
})

/** 添加类错误的兜底文案 */
export const ADD_FALLBACK = '添加失败，请重试'
/** 通用兜底文案 */
export const GENERIC_FALLBACK = '暂时无法完成操作，请稍后重试'

/** 白名单映射，未知码走兜底 */
export function mapError(code: string, fallback: string = GENERIC_FALLBACK): string {
  return ERROR_MESSAGES[code] ?? fallback
}

/** 添加失败专用映射：未知码统一为 ADD_UNKNOWN 的文案 */
export function mapAddError(code: string): string {
  return ERROR_MESSAGES[code] ?? ERROR_MESSAGES.ADD_UNKNOWN ?? ADD_FALLBACK
}

/** 409 已存在：幂等提示，不走错误色、不弹 Toast */
export const DUPLICATE_CODE = 'ADD_DUPLICATE'

/** 会话类 401/419 白名单：只有这两个码才允许清态 + 停播 + 跳登录 */
export const SESSION_INVALID_CODES: readonly string[] = Object.freeze([
  'SESSION_INVALID',
  'UNAUTHENTICATED',
])

/** 绝不按裸 HTTP 401 判断登出 */
export function isSessionInvalid(code: string): boolean {
  return SESSION_INVALID_CODES.includes(code)
}

/** 试听状态文本（reduced-motion 下同样保留，供 sr-only 与 aria 使用） */
export function playerStatusText(status: string): string {
  switch (status) {
    case 'playing': return COPY.statusPlaying
    case 'paused': return COPY.statusPaused
    case 'buffering': return COPY.statusBuffering
    case 'error': return COPY.statusError
    case 'ended': return COPY.statusEnded
    default: return COPY.statusIdle
  }
}

/** 播放条可见静态状态文本：复用统一三值，静态渲染、不进 aria-live，
 *  避免读屏被高频播放状态刷屏；空态返回空串（不显示占位文字）。 */
export function staticStatusText(status: string): string {
  switch (status) {
    case 'playing': return COPY.statusPlaying
    case 'paused': return COPY.statusPaused
    case 'buffering': return COPY.statusBuffering
    case 'error': return COPY.statusError
    case 'ended': return COPY.statusEnded
    default: return ''
  }
}
