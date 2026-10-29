import { FX_LEVELS, type FxLevel } from '~/terminal/fx'
import { THEME_NAMES, isThemeName } from '~/terminal/theme'
import { visible } from '../registry'
import type { Command } from '../types'

export const help: Command = {
  name: 'help',
  aliases: ['?'],
  description: 'list available commands',
  run(ctx) {
    const list = visible(ctx.session.rank, ctx.session.authed)
    for (const c of list) {
      ctx.out.line(`  ${(c.usage ?? `/${c.name}`).padEnd(22)} ${c.description}`)
    }
  },
}

export const clear: Command = {
  name: 'clear',
  description: 'clear the screen',
  run: (ctx) => ctx.ui.clear(),
}

export const theme: Command = {
  name: 'theme',
  description: `switch theme (${THEME_NAMES.join(' / ')})`,
  usage: '/theme <name>',
  argc: 1,
  complete: () => THEME_NAMES,
  run(ctx, [name]) {
    if (!name) {
      const cur = ctx.ui.getTheme()
      return ctx.out.info('themes: ' + THEME_NAMES.map((n) => (n === cur ? `[${n}]` : n)).join('  '))
    }
    if (!isThemeName(name)) return ctx.out.error(`unknown theme: ${name}`)
    ctx.ui.setTheme(name)
    ctx.out.ok(`theme: ${name}`)
  },
}

export const fx: Command = {
  name: 'fx',
  description: 'visual effects: off | low | full',
  usage: '/fx <off|low|full>',
  argc: 1,
  complete: () => FX_LEVELS,
  run(ctx, [level]) {
    if (!level) return ctx.out.info(`fx: ${ctx.ui.getFx()}`)
    if (!(FX_LEVELS as string[]).includes(level)) return ctx.out.error('usage: /fx <off|low|full>')
    ctx.ui.setFx(level as FxLevel)
    ctx.out.ok(`fx: ${level}`)
  },
}

export const sound: Command = {
  name: 'sound',
  description: 'sound on | off',
  usage: '/sound <on|off>',
  argc: 1,
  complete: () => ['on', 'off'],
  run(ctx, [mode]) {
    if (mode !== 'on' && mode !== 'off') return ctx.out.info(`sound: ${ctx.ui.isSound() ? 'on' : 'off'}`)
    ctx.ui.setSound(mode === 'on')
    ctx.out.ok(`sound: ${mode}`)
  },
}
