import type { Terminal } from '@xterm/xterm'
import { C, type Output } from './output'

type Sleep = (ms: number) => Promise<void>

/** Любая клавиша во время загрузки: все паузы схлопываются, текст выводится сразу. */
export function createSkipper(term: Terminal) {
  let skipped = false
  const sub = term.onKey(() => {
    skipped = true
  })
  return {
    sleep: ((ms) =>
      skipped ? Promise.resolve() : new Promise<void>((r) => setTimeout(r, ms))) as Sleep,
    done: () => sub.dispose(),
  }
}

export async function runBoot(out: Output, sleep: Sleep) {
  const lines = ['link up', 'handshake complete', 'signal acquired']
  out.line(`${C.bold}GL1TCH${C.reset} ${C.gray}// glitch in the signal${C.reset}`)
  for (const l of lines) {
    await sleep(220)
    out.line(`${C.gray}[${C.reset}${C.green} OK ${C.reset}${C.gray}]${C.reset} ${l}`)
  }
  await sleep(150)
}

/** Показывается один раз после регистрации. */
export async function wakeUp(out: Output, nick: string, sleep: Sleep) {
  out.clear()
  for (const ch of `Wake up, ${nick}...`) {
    out.write(`${C.green}${ch}${C.reset}`)
    await sleep(90)
  }
  await sleep(1200)
  out.line('')
  out.line(`${C.green}The Matrix has you...${C.reset}`)
  await sleep(1000)
  out.line(`${C.green}Follow the white rabbit.${C.reset}`)
  await sleep(800)
  out.line('')
}
