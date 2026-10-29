import { createApiClient, type SessionLossCause } from './client'
import { createAuthApi } from './auth'
import { createBoardsApi } from './boards'
import { createLocalTokenStorage } from './token-storage'
import { createUsersApi } from './users'

export const API_BASE = '/api'

type LostHandler = (cause: SessionLossCause, bannedUntil?: Date | null) => void
let lostHandler: LostHandler | null = null

/** Стор session подписывается здесь, чтобы api не импортировал сторы (нет циклов). */
export function setSessionLostHandler(fn: LostHandler) {
  lostHandler = fn
}

export const api = createApiClient({
  baseUrl: API_BASE,
  storage: createLocalTokenStorage(),
  onSessionLost: (cause, until) => lostHandler?.(cause, until),
})

export const authApi = createAuthApi(api)
export const usersApi = createUsersApi(api)
export const boardsApi = createBoardsApi(api)
