import type { Command } from '../types'

export const register: Command = {
  name: 'register',
  description: 'create an account',
  usage: '/register <nick>',
  guestOnly: true,
  argc: 1,
  async run(ctx, [nick]) {
    if (!nick) return ctx.out.error('usage: /register <nick>')
    const pass = await ctx.ask('password: ', { secret: true })
    if (!pass) return
    const again = await ctx.ask('repeat password: ', { secret: true })
    if (again === null) return
    if (pass !== again) return ctx.out.error('passwords do not match')

    await ctx.session.register(nick, pass)
    await ctx.ui.wakeUp(nick) // один раз после регистрации
    ctx.out.ok(`welcome, ${nick}`)
  },
}

export const login: Command = {
  name: 'login',
  description: 'sign in',
  usage: '/login <nick>',
  guestOnly: true,
  argc: 1,
  async run(ctx, [nick]) {
    if (!nick) return ctx.out.error('usage: /login <nick>')
    const pass = await ctx.ask('password: ', { secret: true })
    if (!pass) return
    await ctx.session.login(nick, pass)
    ctx.out.ok(`access granted: ${ctx.session.user?.username ?? nick}`)
  },
}

export const logout: Command = {
  name: 'logout',
  description: 'sign out',
  async run(ctx) {
    if (!ctx.session.authed) return ctx.out.error('not logged in')
    await ctx.session.logout()
    ctx.out.info('logged out')
  },
}

export const whoami: Command = {
  name: 'whoami',
  aliases: ['me'],
  description: 'show current user',
  run(ctx) {
    const u = ctx.session.user
    if (!ctx.session.authed || !u) return ctx.out.info('guest')
    ctx.out.info(`${u.username}  role=${u.role.replace('ROLE_', '').toLowerCase()}  id=${u.userId}`)
  },
}
