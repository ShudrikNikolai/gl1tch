import type { Command } from '../types'

const GLYPHS = '01<>/\\|[]{}+=*#$%&'

export const matrix: Command = {
  name: 'matrix',
  description: 'ASCII rain',
  hidden: true,
  async run(ctx) {
    for (let frame = 0; frame < 35; frame++) {
      let row = ''
      for (let c = 0; c < 64; c++) {
        row += Math.random() < 0.35 ? GLYPHS[Math.floor(Math.random() * GLYPHS.length)] : ' '
      }
      ctx.out.line(`\x1b[32m${row}\x1b[0m`)
      await ctx.ui.sleep(60)
    }
    ctx.out.info('Wake up...')
  },
}

export const redpill: Command = {
  name: 'redpill',
  description: 'take the red pill',
  hidden: true,
  run(ctx) {
    ctx.ui.setTheme('redpill')
    ctx.out.line('You take the red pill. The theme stays red.')
  },
}

export const exit: Command = {
  name: 'exit',
  description: 'there is no exit',
  bare: true,
  hidden: true,
  run: (ctx) => ctx.out.info('There is no exit. Only logout.'),
}

export const cookie: Command = {
  name: 'cookie',
  description: 'cookie?',
  bare: true,
  hidden: true,
  run: (ctx) => ctx.out.info('Would you like a cookie?'),
}
