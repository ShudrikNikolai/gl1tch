import type { TokenPair } from './types'

/**
 * Где живут токены. Access только в памяти. Refresh теперь в httpOnly-куке
 * (JS её не видит), поэтому в localStorage лежит лишь флаг-подсказка
 * «возможно, есть сессия», чтобы гость не дёргал refresh при каждой загрузке.
 */
export interface TokenStorage {
  getAccess(): string | null
  hasSession(): boolean
  save(tokens: TokenPair): void
  clear(): void
}

const HINT_KEY = 'gl1tch.session'

export function createLocalTokenStorage(): TokenStorage {
  let access: string | null = null

  return {
    getAccess: () => access,

    hasSession() {
      try {
        return localStorage.getItem(HINT_KEY) === '1'
      } catch {
        return false
      }
    },

    save(tokens) {
      access = tokens.accessToken
      try {
        localStorage.setItem(HINT_KEY, '1')
      } catch {
        // ignore
      }
    },

    clear() {
      access = null
      try {
        localStorage.removeItem(HINT_KEY)
      } catch {
        // ignore
      }
    },
  }
}
