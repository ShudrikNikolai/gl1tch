import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'
import { execute, find, tokenize, visible } from '~/commands'
import { postLines } from '~/commands/render'
import type { Ctx } from '~/commands/types'
import { useBoardsStore } from '~/stores/boards'
import { useSessionStore } from '~/stores/session'
import { useUsersStore } from '~/stores/users'
import { createSkipper, runBoot, wakeUp } from '~/terminal/boot'
import { createFx, loadFxLevel, saveFxLevel } from '~/terminal/fx'
import { createLineEditor, type Completion } from '~/terminal/line-editor'
import { C, createOutput } from '~/terminal/output'
import { isSoundEnabled, loadSound, setSound } from '~/terminal/sound'
import {
  THEMES, applyCssVars, loadThemeName, saveThemeName, toXtermTheme,
} from '~/terminal/theme'

const sleep = (ms: number) => new Promise<void>((r) => setTimeout(r, ms))

export function useTerminal(host: Ref<HTMLElement | null>) {
  const session = useSessionStore()
  const users = useUsersStore()
  const boards = useBoardsStore()
  let disposed = false
  const cleanup: Array<() => void> = []

  async function start() {
    const el = host.value
    if (!el) return
    const root = (el.closest('.crt') as HTMLElement | null) ?? el

    loadSound()
    let themeName = loadThemeName()
    let fxLevel = loadFxLevel()

    // метрики символов считаются по шрифту, поэтому ждём его загрузки до open()
    await document.fonts.load('15px "JetBrains Mono"').catch(() => undefined)
    if (disposed) return

    const term = new Terminal({
      fontFamily: '"JetBrains Mono", monospace',
      fontSize: 15,
      cursorBlink: true,
      scrollback: 2000,
      allowProposedApi: true,
      theme: toXtermTheme(THEMES[themeName]),
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(el)
    try {
      const gl = new WebglAddon()
      gl.onContextLoss(() => gl.dispose())
      term.loadAddon(gl)
    } catch {
      // WebGL недоступен: остаётся DOM-рендерер по умолчанию
    }
    fit.fit()
    const ro = new ResizeObserver(() => fit.fit())
    ro.observe(el)
    const focus = () => term.focus()
    el.addEventListener('click', focus)
    term.focus()

    const fx = createFx(root)
    const out = createOutput(term, { onError: () => fx.burst() })
    cleanup.push(() => {
      ro.disconnect()
      el.removeEventListener('click', focus)
      fx.stop()
      editor.dispose()
      term.dispose()
    })

    const ui: Ctx['ui'] = {
      clear: () => out.clear(),
      sleep,
      getTheme: () => themeName,
      setTheme(name) {
        themeName = name
        saveThemeName(name)
        const t = THEMES[name]
        term.options.theme = toXtermTheme(t)
        applyCssVars(root, t)
        fx.apply(t.fx, fxLevel)
      },
      getFx: () => fxLevel,
      setFx(level) {
        fxLevel = level
        saveFxLevel(level)
        fx.apply(THEMES[themeName].fx, level)
      },
      isSound: isSoundEnabled,
      setSound,
      wakeUp: (nick) => wakeUp(out, nick, sleep),
    }
    ui.setTheme(themeName)

    function complete(line: string): Completion {
      if (!line.startsWith('/')) return { start: 0, items: [] }
      if (!/\s/.test(line)) {
        const part = line.slice(1).toLowerCase()
        const names = visible(session.rank, session.authed).map((c) => c.name)
        return { start: 1, items: names.filter((n) => n.startsWith(part)) }
      }
      const m = line.match(/^\/(\S+)\s+(.*)$/s)
      const cmd = m ? find(m[1]!.toLowerCase()) : undefined
      if (!m || !cmd?.complete) return { start: 0, items: [] }
      const args = tokenize(m[2]!)
      const partial = line.endsWith(' ') ? '' : (args.at(-1)?.value ?? '')
      const items = cmd.complete(ctx, args.map((a) => a.value)).filter((i) => i.startsWith(partial))
      return { start: line.length - partial.length, items }
    }

    const editor = createLineEditor(term, { complete, onClear: () => out.clear() })

    const ctx: Ctx = {
      out,
      ask: (prompt, opts) => editor.readLine(prompt, opts),
      session,
      users,
      boards,
      ui,
    }

    const prompt = () => {
      const name = session.user?.username ?? 'guest'
      return `${C.green}${name}${C.reset}@${C.cyan}gl1tch${C.reset}:${C.yellow}${boards.path}${C.reset}$ `
    }

    // бан: терминал блокируется, оверлей рисует компонент BanOverlay
    const unlockWaiters: Array<() => void> = []
    cleanup.push(
      watch(
        () => boards.livePost,
        (p) => {
          if (!p) return
          boards.livePost = null
          if (p.author?.userId === session.user?.userId) return
          const replyTo = p.parentPostId ? boards.numberOfPost(p.parentPostId) : undefined
          out.system(`new post in ${boards.path}`)
          postLines(p, replyTo).forEach((line) => out.line(line))
        },
      ),
      watch(
        () => session.status,
        (status, prev) => {
          if (status === 'banned') {
            editor.lock()
            term.options.disableStdin = true
          } else if (prev === 'banned') {
            editor.unlock()
            term.options.disableStdin = false
            out.system('ban lifted. please /login')
            unlockWaiters.splice(0).forEach((resolve) => resolve())
          }
        },
      ),
      watch(
        () => session.expiredAt,
        (at) => {
          if (at) out.error('SESSION EXPIRED. please /login')
        },
      ),
    )

    const skipper = createSkipper(term)
    await runBoot(out, skipper.sleep)
    out.line(`${C.gray}[ .. ]${C.reset} contacting server`)
    const state = await session.boot()
    skipper.done()

    if (state === 'authed') out.ok(`session restored: ${session.user?.username}`)
    else if (state === 'offline') out.warn('server unreachable')
    else if (state === 'guest') out.info('type /help. /login or /register to enter')

    while (!disposed) {
      if (session.status === 'banned') {
        await new Promise<void>((resolve) => unlockWaiters.push(resolve))
        continue
      }
      const line = await editor.readLine(prompt())
      if (line === null || disposed) continue
      await execute(line, ctx)
    }
  }

  function dispose() {
    disposed = true
    boards.stopLive()
    cleanup.forEach((fn) => fn())
    cleanup.length = 0
  }

  return { start, dispose }
}
