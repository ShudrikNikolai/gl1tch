import type { Output } from '~/terminal/output'
import type { FxLevel } from '~/terminal/fx'
import type { ThemeName } from '~/terminal/theme'
import type { useSessionStore } from '~/stores/session'
import type { useBoardsStore } from '~/stores/boards'
import type { useUsersStore } from '~/stores/users'

/** guest = доступно всем, остальное сравнивается с рангом роли (user < moderator < admin). */
export type MinRole = 'guest' | 'user' | 'moderator' | 'admin'

export interface Ctx {
  out: Output
  /** Интерактивный запрос (например, пароль). null, если пользователь нажал Ctrl+C. */
  ask(prompt: string, opts?: { secret?: boolean }): Promise<string | null>
  session: ReturnType<typeof useSessionStore>
  users: ReturnType<typeof useUsersStore>
  boards: ReturnType<typeof useBoardsStore>
  ui: {
    clear(): void
    sleep(ms: number): Promise<void>
    getTheme(): ThemeName
    setTheme(name: ThemeName): void
    getFx(): FxLevel
    setFx(level: FxLevel): void
    isSound(): boolean
    setSound(on: boolean): void
    wakeUp(nick: string): Promise<void>
  }
}

export interface Command {
  name: string
  aliases?: string[]
  description: string
  usage?: string
  minRole?: MinRole
  /** Только для гостей: после входа недоступна (/login, /register). */
  guestOnly?: boolean
  /** Можно вызвать без «/» (exit, cookie). */
  bare?: boolean
  /** Не показывать в /help (пасхалки). */
  hidden?: boolean
  /** Число аргументов; вместе с rest последний аргумент забирает остаток строки. */
  argc?: number
  rest?: boolean
  /** Подсказки для Tab по текущему аргументу. */
  complete?: (ctx: Ctx, args: string[]) => string[]
  run(ctx: Ctx, args: string[]): Promise<void> | void
}
