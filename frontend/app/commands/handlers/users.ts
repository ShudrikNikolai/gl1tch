import { ApiError } from '~/api/errors'
import { clean, fmtTime } from '../render'
import type { Command } from '../types'

export const finger: Command = {
  name: 'finger',
  description: 'show info about a user (nick seen on boards, or user id)',
  usage: '/finger <nick|id>',
  minRole: 'user',
  argc: 1,
  async run(ctx, [arg]) {
    if (!arg) return ctx.out.error('usage: /finger <nick|id>')
    // поиска по нику на бэкенде нет: ник работает только для тех, кого мы видели на досках
    const known = ctx.boards.knownUserId(arg)
    let u
    try {
      u = await ctx.users.get(known ?? arg)
    } catch (e) {
      if (!known && e instanceof ApiError && (e.status === 404 || e.status === 400))
        return ctx.out.error('unknown user. Nicks resolve only for users seen on boards; use a user id')
      throw e
    }
    ctx.out.info(`login:   ${clean(u.username)}`)
    ctx.out.info(`role:    ${u.role.replace('ROLE_', '').toLowerCase()}`)
    ctx.out.info(`status:  ${u.status.replace('USER_STATUS_', '').toLowerCase()}`)
    if (u.bio) ctx.out.info(`bio:     ${clean(u.bio)}`)
    if (u.bannedUntil) ctx.out.info(`banned until: ${fmtTime(u.bannedUntil)}`)
    if (u.createdAt) ctx.out.info(`created: ${fmtTime(u.createdAt)}`)
    ctx.out.info(`id:      ${u.userId}`)
  },
}
