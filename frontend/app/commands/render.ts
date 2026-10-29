import type { Board, Post, Thread } from '~/api/boards-types'
import { C } from '~/terminal/output'

/** Всё, что пришло с сервера, чистим от управляющих символов (защита от escape-инъекций в терминал). */
export const clean = (s?: string) => (s ?? '').replace(/[\x00-\x08\x0b-\x1f\x7f-\x9f]/g, '')

/** Литерал \n в введённом тексте превращается в перенос строки. */
export const unescapeBody = (s: string) => s.replace(/\\n/g, '\n')

/** Флаги вида --cat=faq или --mine вырезаются из строки. Пробелы в остатке схлопываются. */
export function extractFlags(text: string) {
  const flags: Record<string, string> = {}
  const rest = text
    .replace(/(?:^|\s)--(\w+)(?:=(\S+))?(?=\s|$)/g, (_m, k: string, v?: string) => {
      flags[k.toLowerCase()] = v ?? 'true'
      return ' '
    })
    .replace(/\s+/g, ' ')
    .trim()
  return { flags, text: rest }
}

export function fmtTime(iso?: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : d.toISOString().slice(0, 16).replace('T', ' ') // UTC
}

const lower = (v: string | undefined, prefix: string) => (v ?? '').replace(prefix, '').toLowerCase()

export function boardLine(i: number, b: Board): string {
  const mod = b.visibility === 'VISIBILITY_MODERATOR' ? `${C.yellow}[mod]${C.reset} ` : ''
  return (
    `${String(i).padStart(2)}  ${C.green}${clean(b.slug).padEnd(14)}${C.reset} ${mod}${clean(b.title)}` +
    `  ${C.gray}${b.threadCount ?? 0}t/${b.postCount ?? 0}p  ${fmtTime(b.lastPostAt)}${C.reset}`
  )
}

function threadFlags(t: Thread): string {
  const f: string[] = []
  if (t.pinned) f.push('pin')
  if (t.status === 'THREAD_STATUS_LOCKED') f.push('lock')
  if (t.status === 'THREAD_STATUS_ARCHIVED') f.push('arch')
  if (t.category && t.category !== 'THREAD_CATEGORY_GENERAL' && t.category !== 'THREAD_CATEGORY_UNSPECIFIED')
    f.push(lower(t.category, 'THREAD_CATEGORY_'))
  if (t.visibility && t.visibility !== 'VISIBILITY_PUBLIC' && t.visibility !== 'VISIBILITY_UNSPECIFIED')
    f.push(lower(t.visibility, 'VISIBILITY_'))
  return f.length ? `${C.yellow}[${f.join(',')}]${C.reset} ` : ''
}

export function threadLine(i: number, t: Thread): string {
  return (
    `${String(i).padStart(3)}  ${threadFlags(t)}${clean(t.title)}` +
    `  ${C.gray}by ${clean(t.author?.username)}  ${t.postCount ?? 0}p  ${fmtTime(t.lastPostAt ?? t.createdAt)}${C.reset}`
  )
}

export function threadHeader(t: Thread): string[] {
  return [
    `${C.bold}${clean(t.title)}${C.reset}`,
    `${C.gray}by ${clean(t.author?.username)}  ${fmtTime(t.createdAt)}  ${t.postCount ?? 0} posts  ` +
      `${lower(t.status, 'THREAD_STATUS_')}  ${lower(t.category, 'THREAD_CATEGORY_')}  id=${clean(t.threadId)}${C.reset}`,
    `${C.gray}${'-'.repeat(40)}${C.reset}`,
  ]
}

export function postLines(p: Post, replyTo?: number): string[] {
  const indent = p.parentPostId ? '    ' : ''
  const head =
    `${indent}${C.cyan}#${p.number ?? '?'}${C.reset} ${C.green}${clean(p.author?.username) || '?'}${C.reset} ` +
    `${C.gray}${fmtTime(p.createdAt)}${p.editedAt ? ' (edited)' : ''}${replyTo ? ` re #${replyTo}` : ''}${C.reset}`
  if (p.deleted) return [head, `${indent}  ${C.gray}[deleted]${C.reset}`]
  return [head, ...clean(p.body).split('\n').map((l) => `${indent}  ${l}`)]
}

// ---- парсинг значений флагов

const CATEGORY = {
  general: 'THREAD_CATEGORY_GENERAL',
  faq: 'THREAD_CATEGORY_FAQ',
  flood: 'THREAD_CATEGORY_FLOOD',
} as const
const STATUS = {
  open: 'THREAD_STATUS_OPEN',
  locked: 'THREAD_STATUS_LOCKED',
  archived: 'THREAD_STATUS_ARCHIVED',
} as const
const VISIBILITY = {
  public: 'VISIBILITY_PUBLIC',
  author: 'VISIBILITY_AUTHOR',
  moderator: 'VISIBILITY_MODERATOR',
} as const
const ORDER = {
  created: 'ORDER_BY_THREAD_CREATED_AT',
  bump: 'ORDER_BY_THREAD_LAST_POST_AT',
} as const

function pick<T extends Record<string, string>>(map: T, v: string | undefined) {
  return v ? map[v.toLowerCase() as keyof T] : undefined
}
export const parseCategory = (v?: string) => pick(CATEGORY, v)
export const parseStatus = (v?: string) => pick(STATUS, v)
export const parseVisibility = (v?: string) => pick(VISIBILITY, v)
export const parseOrder = (v?: string) => pick(ORDER, v)
export const CATEGORY_NAMES = Object.keys(CATEGORY)
export const VISIBILITY_NAMES = Object.keys(VISIBILITY)
