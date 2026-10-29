import type { Command, MinRole } from './types'

const byName = new Map<string, Command>()
const RANK: Record<MinRole, number> = { guest: 0, user: 1, moderator: 2, admin: 3 }

export function register(...commands: Command[]) {
  for (const c of commands) {
    byName.set(c.name, c)
    c.aliases?.forEach((a) => byName.set(a, c))
  }
}

export const find = (name: string) => byName.get(name)
export const all = () => [...new Set(byName.values())]

export function canUse(cmd: Command, rank: number, authed: boolean): boolean {
  if (cmd.guestOnly && authed) return false
  return rank >= RANK[cmd.minRole ?? 'guest']
}

export const visible = (rank: number, authed: boolean) =>
  all().filter((c) => !c.hidden && canUse(c, rank, authed))

function distance(a: string, b: string): number {
  const dp = Array.from({ length: a.length + 1 }, (_, i) => [i, ...Array(b.length).fill(0)] as number[])
  for (let j = 1; j <= b.length; j++) dp[0]![j] = j
  for (let i = 1; i <= a.length; i++)
    for (let j = 1; j <= b.length; j++)
      dp[i]![j] = Math.min(
        dp[i - 1]![j]! + 1,
        dp[i]![j - 1]! + 1,
        dp[i - 1]![j - 1]! + (a[i - 1] === b[j - 1] ? 0 : 1),
      )
  return dp[a.length]![b.length]!
}

/** Ближайшая команда для «возможно, вы имели в виду...». */
export function suggest(name: string, pool: string[]): string | null {
  let best: string | null = null
  let bestDist = 3
  for (const p of pool) {
    const d = distance(name, p)
    if (d < bestDist) {
      best = p
      bestDist = d
    }
  }
  return best
}
