import { ApiError, parseApiError } from './errors'
import type { TokenStorage } from './token-storage'
import type { TokenPair } from './types'

export type SessionLossCause = 'expired' | 'banned'

export interface ApiClientOptions {
  /** Например '/api': reverse proxy шлёт /api/* на gateway. */
  baseUrl: string
  storage: TokenStorage
  /**
   * Сессия потеряна: refresh не удался или пользователь забанен.
   * Токены к этому моменту уже стёрты. Обработчик (стор session) должен быть
   * идемпотентным: при бане его могут вызвать несколько параллельных запросов.
   */
  onSessionLost: (cause: SessionLossCause, bannedUntil?: Date | null) => void
}

interface RequestOptions {
  body?: unknown
  query?: Record<string, string | number | undefined>
  /** false для login/register/refresh: без Bearer и без попытки refresh при 401. */
  auth?: boolean
}

function safeJson(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return undefined
  }
}

export function createApiClient(opts: ApiClientOptions) {
  const { baseUrl, storage } = opts
    console.log(' ???? ', baseUrl)
  let refreshing: Promise<boolean> | null = null

  function lose(cause: SessionLossCause, bannedUntil?: Date | null) {
    storage.clear()
    opts.onSessionLost(cause, bannedUntil)
  }

  function buildUrl(path: string, query?: RequestOptions['query']): string {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(query ?? {})) {
      if (value !== undefined) params.set(key, String(value))
    }
    const qs = params.toString()
    return `${baseUrl}${path}${qs ? `?${qs}` : ''}`
  }

  async function send<T>(method: string, path: string, o: RequestOptions = {}): Promise<T> {
    const headers: Record<string, string> = {}
    if (o.body !== undefined) headers['Content-Type'] = 'application/json'
    const access = o.auth === false ? null : storage.getAccess()
    if (access) headers.Authorization = `Bearer ${access}`

    const res = await fetch(buildUrl(path, o.query), {
      method,
      headers,
      credentials: 'include', // нужны для Set-Cookie и отправки refresh-куки
      body: o.body === undefined ? undefined : JSON.stringify(o.body),
    })

    const text = await res.text()
    const data = text ? safeJson(text) : undefined
    if (res.ok) return data as T

    const err = parseApiError(res.status, data)
    if (err.isBan) lose('banned', err.bannedUntil)
    throw err
  }

  async function doRefresh(): Promise<boolean> {
    if (!storage.hasSession()) {
      lose('expired')
      return false
    }
    try {
      // refresh-токен едет в httpOnly-куке, тело пустое
      const res = await send<{ tokens: TokenPair }>('POST', '/auth/refresh', {
        body: {},
        auth: false,
      })
      storage.save(res.tokens)
      return true
    } catch (e) {
      // сеть или 5xx: сессию не трогаем, пусть ошибка уйдёт вызывающему
      if (!(e instanceof ApiError) || e.status >= 500) throw e
      if (!e.isBan) lose('expired') // бан уже обработан в send()
      return false
    }
  }

  /** Все параллельные 401 ждут один и тот же refresh. */
  function refreshOnce(): Promise<boolean> {
    refreshing ??= doRefresh().finally(() => {
      refreshing = null
    })
    return refreshing
  }

  async function request<T>(method: string, path: string, o: RequestOptions = {}): Promise<T> {
    const usedAccess = storage.getAccess()
    try {
      return await send<T>(method, path, o)
    } catch (e) {
      if (!(e instanceof ApiError) || e.status !== 401 || o.auth === false) throw e

      // токен уже успели обновить, пока запрос был в полёте: просто повторяем
      const refreshedMeanwhile = storage.getAccess() !== null && storage.getAccess() !== usedAccess
      if (!refreshedMeanwhile && !(await refreshOnce())) throw e

      return send<T>(method, path, o) // ровно один повтор
    }
  }

  return {
    get: <T>(path: string, query?: RequestOptions['query']) =>
      request<T>('GET', path, { query }),
    post: <T>(path: string, body?: unknown, auth = true) =>
      request<T>('POST', path, { body, auth }),
    patch: <T>(path: string, body: unknown = {}) => request<T>('PATCH', path, { body }),
    del: <T = unknown>(path: string) => request<T>('DELETE', path),
    setTokens: (tokens: TokenPair) => storage.save(tokens),
    clearTokens: () => storage.clear(),
    hasSession: () => storage.hasSession(),
  }
}

export type ApiClient = ReturnType<typeof createApiClient>
