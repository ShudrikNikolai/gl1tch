import type { Terminal } from '@xterm/xterm'
import { blip } from './sound'

/** Только индексы палитры ANSI: цвета берутся из активной темы, смена темы перекрашивает всё. */
export const C = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  cyan: '\x1b[36m',
  gray: '\x1b[90m',
}

export function createOutput(term: Terminal, hooks: { onError?: () => void } = {}) {
  const write = (s: string) => term.write(s)
  const line = (s = '') => write(`${s}\r\n`)

  return {
    write,
    line,
    info: (s: string) => line(`${C.cyan}${s}${C.reset}`),
    system: (s: string) => line(`${C.gray}[sys]${C.reset} ${s}`),
    warn: (s: string) => line(`${C.yellow}WARN${C.reset} ${s}`),
    ok: (s: string) => {
      line(`${C.green}${s}${C.reset}`)
      blip('ok')
    },
    error: (s: string) => {
      line(`${C.red}ERR${C.reset} ${s}`)
      blip('error')
      hooks.onError?.()
    },
    /** Очищает экран и scrollback. */
    clear: () => write('\x1b[2J\x1b[3J\x1b[H'),
  }
}

export type Output = ReturnType<typeof createOutput>
