import type { Post } from '~/api/boards-types'
import { C } from '~/terminal/output'
import type { Command, Ctx } from '../types'
import {
  CATEGORY_NAMES, VISIBILITY_NAMES, boardLine, clean, extractFlags, parseCategory, parseOrder,
  parseStatus, parseVisibility, postLines, threadHeader, threadLine, unescapeBody,
} from '../render'

// ---- общие помощники (не экспортируются, иначе реестр примет их за команды)

const NEED_BOARD = 'enter a board first: /board <n|slug>'
const NEED_THREAD = 'open a thread first: /thread <n|id>'
const hint = (ctx: Ctx, s: string) => ctx.out.line(`${C.gray}${s}${C.reset}`)

function printPost(ctx: Ctx, p: Post) {
  const replyTo = p.parentPostId ? ctx.boards.numberOfPost(p.parentPostId) : undefined
  postLines(p, replyTo).forEach((l) => ctx.out.line(l))
}

function printThreads(ctx: Ctx, items: Parameters<typeof threadLine>[1][], offset: number, hasMore: boolean) {
  if (items.length === 0 && offset === 0) return ctx.out.info('no threads')
  items.forEach((t, i) => ctx.out.line(threadLine(offset + i + 1, t)))
  hint(ctx, hasMore ? '-- /more for next page, /thread <n> to open --' : '-- /thread <n> to open --')
}

async function openAndPrint(ctx: Ctx, id: string) {
  const { thread, items, hasMore } = await ctx.boards.openThread(id)
  ctx.out.line('')
  threadHeader(thread).forEach((l) => ctx.out.line(l))
  items.forEach((p) => printPost(ctx, p))
  hint(ctx, hasMore ? '-- /more for next page --' : '-- end. /post <text>, /reply <#n> <text> --')
}

const postNumber = (arg: string) => Number(arg.replace(/^#/, ''))

function requirePost(ctx: Ctx, arg: string): Post | null {
  const n = postNumber(arg)
  const p = Number.isInteger(n) ? ctx.boards.postByNumber(n) : undefined
  if (!p) ctx.out.error(`post ${arg} is not loaded (open the thread, /more for further pages)`)
  return p ?? null
}

// ---- навигация

export const boards: Command = {
  name: 'boards',
  description: 'list boards',
  minRole: 'user',
  async run(ctx) {
    const list = await ctx.boards.loadBoards(true)
    if (list.length === 0) return ctx.out.info('no boards yet')
    list.forEach((b, i) => ctx.out.line(boardLine(i + 1, b)))
    hint(ctx, '-- /board <n|slug> to enter --')
  },
}

export const board: Command = {
  name: 'board',
  aliases: ['b'],
  description: 'enter a board and list its threads',
  usage: '/board <n|slug>',
  minRole: 'user',
  argc: 1,
  complete: (ctx) => ctx.boards.boardSlugs(),
  async run(ctx, [arg]) {
    if (!arg) return ctx.out.error('usage: /board <n|slug>')
    const b = await ctx.boards.findBoard(arg)
    if (!b) return ctx.out.error(`no such board: ${arg}`)
    ctx.boards.enterBoard(b)
    ctx.out.ok(`~/${clean(b.slug)}  ${clean(b.title)}`)
    if (b.description) ctx.out.info(clean(b.description))
    const { items, offset, hasMore } = await ctx.boards.fetchThreads({})
    printThreads(ctx, items, offset, hasMore)
  },
}

export const threads: Command = {
  name: 'threads',
  aliases: ['ls'],
  description: 'list threads: [search] --cat=faq|general|flood --status=open|locked|archived --by=created|bump --mine',
  usage: '/threads [search]',
  minRole: 'user',
  argc: 1,
  rest: true,
  async run(ctx, [raw = '']) {
    if (!ctx.boards.board) return ctx.out.error(NEED_BOARD)
    const { flags, text } = extractFlags(raw)
    const category = parseCategory(flags.cat)
    const status = parseStatus(flags.status)
    const orderBy = parseOrder(flags.by)
    if (flags.cat && !category) return ctx.out.error(`--cat: ${CATEGORY_NAMES.join(' | ')}`)
    if (flags.status && !status) return ctx.out.error('--status: open | locked | archived')
    if (flags.by && !orderBy) return ctx.out.error('--by: created | bump')

    const { items, offset, hasMore } = await ctx.boards.fetchThreads({
      category,
      status,
      orderBy,
      search: text || undefined,
      authorUserId: flags.mine ? ctx.session.user?.userId : undefined,
    })
    printThreads(ctx, items, offset, hasMore)
  },
}

export const thread: Command = {
  name: 'thread',
  aliases: ['t', 'open'],
  description: 'open a thread (number from the last list, or id)',
  usage: '/thread <n|id>',
  minRole: 'user',
  argc: 1,
  async run(ctx, [arg]) {
    if (!arg) return ctx.out.error('usage: /thread <n|id>')
    await openAndPrint(ctx, ctx.boards.resolveThreadId(arg))
  },
}

export const more: Command = {
  name: 'more',
  description: 'next page of threads or posts',
  minRole: 'user',
  async run(ctx) {
    const kind = ctx.boards.pagerKind()
    if (!kind) return ctx.out.info('nothing more')
    if (kind === 'threads') {
      const offset = ctx.boards.threadOffset()
      const { items, hasMore } = await ctx.boards.fetchThreads(null, true)
      return printThreads(ctx, items, offset, hasMore)
    }
    const { items, hasMore } = await ctx.boards.fetchPosts(true)
    items.forEach((p) => printPost(ctx, p))
    hint(ctx, hasMore ? '-- /more for next page --' : '-- end --')
  },
}

export const back: Command = {
  name: 'back',
  aliases: ['up'],
  description: 'go one level up (thread, board, root)',
  minRole: 'user',
  run(ctx) {
    ctx.boards.back()
    ctx.out.info(ctx.boards.path)
  },
}

// ---- посты

export const post: Command = {
  name: 'post',
  description: 'write a post in the open thread (\\n for a line break)',
  usage: '/post <text>',
  minRole: 'user',
  argc: 1,
  rest: true,
  async run(ctx, [raw]) {
    const t = ctx.boards.thread
    if (!t) return ctx.out.error(NEED_THREAD)
    if (t.status !== 'THREAD_STATUS_OPEN') return ctx.out.error('thread is closed')
    const body = unescapeBody(raw ?? '').trim()
    if (!body) return ctx.out.error('usage: /post <text>')
    const p = await ctx.boards.createPost(body)
    ctx.out.ok(`posted #${p.number ?? '?'}`)
  },
}

export const reply: Command = {
  name: 'reply',
  description: 'reply to a top-level post',
  usage: '/reply <#n> <text>',
  minRole: 'user',
  argc: 2,
  rest: true,
  async run(ctx, [num, raw]) {
    const t = ctx.boards.thread
    if (!t) return ctx.out.error(NEED_THREAD)
    if (t.status !== 'THREAD_STATUS_OPEN') return ctx.out.error('thread is closed')
    const body = unescapeBody(raw ?? '').trim()
    if (!num || !body) return ctx.out.error('usage: /reply <#n> <text>')
    const parent = requirePost(ctx, num)
    if (!parent) return
    if (parent.parentPostId) return ctx.out.error('you can only reply to top-level posts')
    const p = await ctx.boards.createPost(body, parent)
    ctx.out.ok(`posted #${p.number ?? '?'} (re #${parent.number})`)
  },
}

export const edit: Command = {
  name: 'edit',
  description: 'edit your post',
  usage: '/edit <#n> <text>',
  minRole: 'user',
  argc: 2,
  rest: true,
  async run(ctx, [num, raw]) {
    const body = unescapeBody(raw ?? '').trim()
    if (!num || !body) return ctx.out.error('usage: /edit <#n> <text>')
    const p = requirePost(ctx, num)
    if (!p) return
    await ctx.boards.editPost(p, body)
    ctx.out.ok(`edited #${p.number}`)
  },
}

export const rm: Command = {
  name: 'rm',
  description: 'delete a post (soft: replies stay)',
  usage: '/rm <#n>',
  minRole: 'user',
  argc: 1,
  async run(ctx, [num]) {
    if (!num) return ctx.out.error('usage: /rm <#n>')
    const p = requirePost(ctx, num)
    if (!p) return
    await ctx.boards.deletePost(p)
    ctx.out.ok(`deleted #${p.number}`)
  },
}

// ---- темы

export const newthread: Command = {
  name: 'newthread',
  aliases: ['nt'],
  description: 'create a thread: <title> --cat=general|faq|flood --vis=public|author|moderator',
  usage: '/newthread <title>',
  minRole: 'user',
  argc: 1,
  rest: true,
  async run(ctx, [raw = '']) {
    if (!ctx.boards.board) return ctx.out.error(NEED_BOARD)
    const { flags, text: title } = extractFlags(raw)
    const category = parseCategory(flags.cat)
    const visibility = parseVisibility(flags.vis)
    if (!title) return ctx.out.error('usage: /newthread <title> [--cat=..] [--vis=..]')
    if (flags.cat && !category) return ctx.out.error(`--cat: ${CATEGORY_NAMES.join(' | ')}`)
    if (flags.vis && !visibility) return ctx.out.error(`--vis: ${VISIBILITY_NAMES.join(' | ')}`)

    const body = await ctx.ask('first post (\\n = line break): ')
    if (body === null || !body.trim()) return ctx.out.info('cancelled')

    const created = await ctx.boards.createThread({
      title,
      body: unescapeBody(body).trim(),
      category,
      visibility,
    })
    ctx.out.ok('thread created')
    await openAndPrint(ctx, created.threadId)
  },
}

export const title: Command = {
  name: 'title',
  description: 'rename the open thread',
  usage: '/title <text>',
  minRole: 'user',
  argc: 1,
  rest: true,
  async run(ctx, [raw]) {
    if (!ctx.boards.thread) return ctx.out.error(NEED_THREAD)
    const text = (raw ?? '').trim()
    if (!text) return ctx.out.error('usage: /title <text>')
    const t = await ctx.boards.updateTitle(text)
    ctx.out.ok(`renamed: ${clean(t.title)}`)
  },
}

function moderation(
  name: 'lock' | 'unlock' | 'pin' | 'unpin',
  description: string,
  withReason: boolean,
): Command {
  return {
    name,
    description,
    usage: withReason ? `/${name} [reason]` : `/${name}`,
    minRole: 'moderator',
    argc: 1,
    rest: true,
    async run(ctx, [reason]) {
      if (!ctx.boards.thread) return ctx.out.error(NEED_THREAD)
      await ctx.boards.moderate(name, withReason ? (reason ?? '').trim() : undefined)
      ctx.out.ok(`${name}: done`)
    },
  }
}
export const lock = moderation('lock', 'lock the open thread (read-only)', true)
export const unlock = moderation('unlock', 'unlock the open thread', true)
export const pin = moderation('pin', 'pin the open thread to the top', false)
export const unpin = moderation('unpin', 'unpin the open thread', false)

export const rmthread: Command = {
  name: 'rmthread',
  description: 'delete the open thread',
  minRole: 'user',
  async run(ctx) {
    const t = ctx.boards.thread
    if (!t) return ctx.out.error(NEED_THREAD)
    const answer = await ctx.ask(`delete "${clean(t.title)}"? type yes: `)
    if (answer?.trim().toLowerCase() !== 'yes') return ctx.out.info('cancelled')
    await ctx.boards.deleteThread()
    ctx.out.ok('thread deleted')
  },
}

// ---- доски

export const mkboard: Command = {
  name: 'mkboard',
  description: 'create a board: <slug> <title> --vis=public|moderator --order=10',
  usage: '/mkboard <slug> <title>',
  minRole: 'moderator',
  argc: 2,
  rest: true,
  async run(ctx, [slug, raw = '']) {
    const { flags, text: title } = extractFlags(raw)
    if (!slug || !title) return ctx.out.error('usage: /mkboard <slug> <title> [--vis=..] [--order=n]')
    const visibility = parseVisibility(flags.vis ?? 'public')
    if (visibility !== 'VISIBILITY_PUBLIC' && visibility !== 'VISIBILITY_MODERATOR')
      return ctx.out.error('--vis: public | moderator')
    const description = await ctx.ask('description (empty = none): ')
    if (description === null) return ctx.out.info('cancelled')

    const b = await ctx.boards.createBoard({
      slug,
      title,
      description: description.trim(),
      visibility,
      sortOrder: Number.parseInt(flags.order ?? '0', 10) || 0,
    })
    ctx.out.ok(`board created: ${clean(b.slug)}`)
  },
}
