import type { Terminal } from '@xterm/xterm'
import { blip } from './sound'

export interface Completion {
  /** Индекс в строке, с которого заменяется набранная часть. */
  start: number
  items: string[]
}

export interface EditorOptions {
  complete?: (beforeCursor: string) => Completion
  onClear?: () => void
}

export interface ReadOptions {
  /** Ничего не печатать, не сохранять в историю (пароли). */
  secret?: boolean
  history?: boolean
}

interface ReadState {
  resolve: (value: string | null) => void
  prompt: string
  promptLen: number
  secret: boolean
  useHistory: boolean
  buf: string
  cur: number
  row: number // строка экрана, на которой стоит курсор, относительно начала ввода
  histIdx: number
  draft: string
}

const HISTORY_KEY = 'gl1tch.history'
const HISTORY_MAX = 100
const LINE_MAX = 2000
const ANSI_RE = /\x1b\[[0-9;]*m/g

function loadHistory(): string[] {
  try {
    const raw = JSON.parse(localStorage.getItem(HISTORY_KEY) ?? '[]')
    return Array.isArray(raw) ? raw.filter((x): x is string => typeof x === 'string') : []
  } catch {
    return []
  }
}

function saveHistory(history: string[]) {
  try {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history))
  } catch {
    // ignore
  }
}

export function createLineEditor(term: Terminal, opts: EditorOptions = {}) {
  const history = loadHistory()
  let locked = false
  let s: ReadState | null = null

  function redraw() {
    if (!s) return
    const cols = term.cols
    const shown = s.secret ? '' : s.buf
    const endPos = s.promptLen + shown.length
    const curPos = s.promptLen + (s.secret ? 0 : s.cur)

    let out = '\r' + (s.row > 0 ? `\x1b[${s.row}A` : '') + '\x1b[J' + s.prompt + shown
    // ровно заполненная строка: курсор ещё «висит» в конце, переносим вручную
    if (endPos > 0 && endPos % cols === 0) out += '\r\n'

    const endRow = Math.floor(endPos / cols)
    const curRow = Math.floor(curPos / cols)
    const curCol = curPos % cols
    if (endRow > curRow) out += `\x1b[${endRow - curRow}A`
    out += '\r' + (curCol > 0 ? `\x1b[${curCol}C` : '')

    s.row = curRow
    term.write(out)
  }

  function finish(value: string | null) {
    if (!s) return
    const cur = s
    s = null
    cur.resolve(value)
  }

  function submit() {
    if (!s) return
    s.cur = s.buf.length
    redraw()
    term.write('\r\n')
    const { buf, secret, useHistory } = s
    if (useHistory && !secret && buf.trim() && history[history.length - 1] !== buf) {
      history.push(buf)
      if (history.length > HISTORY_MAX) history.shift()
      saveHistory(history)
    }
    finish(buf)
  }

  function cancel() {
    if (!s) return
    s.cur = s.buf.length
    redraw()
    term.write('^C\r\n')
    finish(null)
  }

  function insert(text: string) {
    if (!s || !text) return
    const room = LINE_MAX - s.buf.length
    if (room <= 0) return
    const chunk = text.slice(0, room)
    s.buf = s.buf.slice(0, s.cur) + chunk + s.buf.slice(s.cur)
    s.cur += chunk.length
    redraw()
    blip('key')
  }

  function setBuf(buf: string, cur = buf.length) {
    if (!s) return
    s.buf = buf
    s.cur = cur
    redraw()
  }

  function historyMove(dir: -1 | 1) {
    if (!s || !s.useHistory || s.secret) return
    if (dir === -1 && s.histIdx > 0) {
      if (s.histIdx === history.length) s.draft = s.buf
      s.histIdx--
      setBuf(history[s.histIdx] ?? '')
    } else if (dir === 1 && s.histIdx < history.length) {
      s.histIdx++
      setBuf(s.histIdx === history.length ? s.draft : (history[s.histIdx] ?? ''))
    }
  }

  function tab() {
    if (!s || s.secret || !opts.complete) return
    const before = s.buf.slice(0, s.cur)
    const { start, items } = opts.complete(before)
    if (items.length === 0) return
    const typed = before.slice(start)
    const replaceWith = (text: string) =>
      setBuf(before.slice(0, start) + text + s!.buf.slice(s!.cur), start + text.length)

    if (items.length === 1) return replaceWith(`${items[0]} `)

    let prefix = items[0] ?? ''
    for (const it of items) while (!it.startsWith(prefix)) prefix = prefix.slice(0, -1)
    if (prefix.length > typed.length) return replaceWith(prefix)

    s.cur = s.buf.length
    redraw()
    term.write(`\r\n${items.join('  ')}\r\n`)
    s.row = 0
    redraw()
  }

  function handleData(data: string) {
    if (!s || locked) return
    switch (data) {
      case '\r': return submit()
      case '\x7f':
      case '\b':
        if (s.cur > 0) setBuf(s.buf.slice(0, s.cur - 1) + s.buf.slice(s.cur), s.cur - 1)
        return
      case '\x1b[3~':
        if (s.cur < s.buf.length) setBuf(s.buf.slice(0, s.cur) + s.buf.slice(s.cur + 1), s.cur)
        return
      case '\x1b[A': return historyMove(-1)
      case '\x1b[B': return historyMove(1)
      case '\x1b[C': return setBuf(s.buf, Math.min(s.buf.length, s.cur + 1))
      case '\x1b[D': return setBuf(s.buf, Math.max(0, s.cur - 1))
      case '\x1b[H':
      case '\x1b[1~':
      case '\x01': return setBuf(s.buf, 0)
      case '\x1b[F':
      case '\x1b[4~':
      case '\x05': return setBuf(s.buf, s.buf.length)
      case '\x15': return setBuf(s.buf.slice(s.cur), 0)
      case '\x17': {
        const before = s.buf.slice(0, s.cur).replace(/\s*\S*$/, '')
        return setBuf(before + s.buf.slice(s.cur), before.length)
      }
      case '\x0c':
        opts.onClear?.()
        s.row = 0
        return redraw()
      case '\x03': return cancel()
      case '\t': return tab()
    }
    if (data.startsWith('\x1b')) return // прочие escape-последовательности игнорируем

    // вставка: переводы строк превращаются в литерал \n (в секретном вводе отбрасываются)
    const text = s.secret
      ? data.replace(/[\x00-\x1f\x7f]/g, '')
      : data.replace(/\r\n|\r|\n/g, '\\n').replace(/[\x00-\x1f\x7f]/g, '')
    insert(text)
  }

  const sub = term.onData(handleData)

  function readLine(prompt: string, o: ReadOptions = {}): Promise<string | null> {
    if (locked) return Promise.resolve(null)
    return new Promise((resolve) => {
      s = {
        resolve,
        prompt,
        promptLen: prompt.replace(ANSI_RE, '').length,
        secret: !!o.secret,
        useHistory: o.history ?? !o.secret,
        buf: '',
        cur: 0,
        row: 0,
        histIdx: history.length,
        draft: '',
      }
      redraw()
    })
  }

  return {
    readLine,
    lock() {
      locked = true
      finish(null)
    },
    unlock() {
      locked = false
    },
    dispose() {
      sub.dispose()
      finish(null)
    },
  }
}

export type LineEditor = ReturnType<typeof createLineEditor>
