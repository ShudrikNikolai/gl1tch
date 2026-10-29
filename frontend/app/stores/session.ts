import { defineStore } from 'pinia'
import { api, authApi, usersApi, setSessionLostHandler } from '~/api'
import type { SessionLossCause } from '~/api/client'
import { ApiError } from '~/api/errors'
import type { Role, User } from '~/api/types'
import { useBoardsStore } from './boards'
import { useUsersStore } from './users'

export type SessionStatus = 'booting' | 'guest' | 'authed' | 'banned'
export type BootResult = 'authed' | 'guest' | 'banned' | 'offline'

const BAN_KEY = 'gl1tch.ban'
const DAY_MS = 24 * 60 * 60 * 1000
const MAX_TIMER = 2 ** 31 - 1

const ROLE_RANK: Record<Role, number> = {
  ROLE_UNSPECIFIED: 0,
  ROLE_USER: 1,
  ROLE_MODERATOR: 2,
  ROLE_ADMIN: 3,
}

function readBan(): Date | null {
  try {
    const raw = localStorage.getItem(BAN_KEY)
    if (!raw) return null
    const date = new Date(raw)
    return Number.isNaN(date.getTime()) ? null : date
  } catch {
    return null
  }
}
function writeBan(date: Date) {
  try {
    localStorage.setItem(BAN_KEY, date.toISOString())
  } catch {
    // ignore
  }
}
function clearBan() {
  try {
    localStorage.removeItem(BAN_KEY)
  } catch {
    // ignore
  }
}

export const useSessionStore = defineStore('session', () => {
  const status = ref<SessionStatus>('booting')
  const user = ref<User | null>(null)
  const banUntil = ref<Date | null>(null)
  /** Растёт при потере сессии из-за истёкшего refresh: терминал печатает уведомление. */
  const expiredAt = ref<number | null>(null)
  let banTimer: ReturnType<typeof setTimeout> | null = null

  const authed = computed(() => status.value === 'authed')
  /** 0 = гость (и ROLE_UNSPECIFIED), 1 = user, 2 = moderator, 3 = admin. */
  const rank = computed(() =>
    status.value === 'authed' && user.value ? (ROLE_RANK[user.value.role] ?? 0) : 0,
  )

  function resetAll() {
    user.value = null
    useUsersStore().reset()
    useBoardsStore().reset()
  }

  function liftBan() {
    if (banTimer) clearTimeout(banTimer)
    banTimer = null
    clearBan()
    banUntil.value = null
    status.value = 'guest'
  }

  function scheduleLift(end: Date) {
    if (banTimer) clearTimeout(banTimer)
    const ms = end.getTime() - Date.now()
    if (ms <= 0) return liftBan()
    banTimer = setTimeout(() => scheduleLift(end), Math.min(ms, MAX_TIMER))
  }

  function enterBan(until: Date | null) {
    const end = until ?? new Date(Date.now() + DAY_MS)
    writeBan(end)
    banUntil.value = end
    status.value = 'banned'
    scheduleLift(end)
  }

  // идемпотентно: при бане сюда могут прийти несколько параллельных запросов
  function onSessionLost(cause: SessionLossCause, until?: Date | null) {
    resetAll()
    if (cause === 'banned') {
      enterBan(until ?? null)
    } else {
      status.value = 'guest'
      expiredAt.value = Date.now()
    }
  }
  setSessionLostHandler(onSessionLost)

  async function boot(): Promise<BootResult> {
    const until = readBan()
    if (until && until.getTime() > Date.now()) {
      enterBan(until)
      return 'banned'
    }
    clearBan()
    if (!api.hasSession()) {
      status.value = 'guest'
      return 'guest'
    }
    try {
      user.value = await usersApi.me() // при 401 клиент сам сделает refresh
      status.value = 'authed'
      return 'authed'
    } catch (e) {
      if (status.value === 'banned') return 'banned'
      status.value = 'guest'
      return e instanceof ApiError ? 'guest' : 'offline'
    }
  }

  async function login(username: string, password: string) {
    await authApi.login(username, password)
    user.value = await usersApi.me()
    status.value = 'authed'
  }

  async function register(username: string, password: string) {
    await authApi.register(username, password)
    user.value = await usersApi.me()
    status.value = 'authed'
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // токены в любом случае стёрты локально
    }
    resetAll()
    status.value = 'guest'
  }

  return { status, user, banUntil, expiredAt, authed, rank, boot, login, register, logout, resetAll }
})
