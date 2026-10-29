import { describeError } from './errors'
import { parseArgs, splitCommand } from './parser'
import { all, canUse, find, register, suggest } from './registry'
import * as authCmds from './handlers/auth'
import * as boardCmds from './handlers/boards'
import * as easterCmds from './handlers/easter'
import * as systemCmds from './handlers/system'
import * as userCmds from './handlers/users'
import type { Command, Ctx } from './types'

register(
  ...(Object.values(authCmds) as Command[]),
  ...(Object.values(userCmds) as Command[]),
  ...(Object.values(boardCmds) as Command[]),
  ...(Object.values(systemCmds) as Command[]),
  ...(Object.values(easterCmds) as Command[]),
)

export { find, visible } from './registry'
export { tokenize } from './parser'

export async function execute(line: string, ctx: Ctx) {
  const parsed = splitCommand(line)
  if (!parsed) return
  const { slash, name, argsRaw } = parsed
  const cmd = find(name)

  // строка без «/»: bare-команда либо сообщение в комнату
  if (!slash && !cmd?.bare) {
    if (!ctx.session.authed) return ctx.out.error('ACCESS DENIED: /login or /register first')
    return ctx.out.system('chat rooms are not online yet')
  }

  if (!cmd) {
    const hint = suggest(name, all().filter((c) => !c.hidden).map((c) => c.name))
    return ctx.out.error(`unknown command: /${name}${hint ? `, did you mean /${hint}?` : ''}`)
  }
  if (cmd.guestOnly && ctx.session.authed) return ctx.out.error('already logged in')
  if (!canUse(cmd, ctx.session.rank, ctx.session.authed)) return ctx.out.error('ACCESS DENIED')

  try {
    await cmd.run(ctx, parseArgs(argsRaw, cmd))
  } catch (e) {
    ctx.out.error(describeError(e))
    if (!(e instanceof Error) || e.name !== 'ApiError') console.error(e)
  }
}
