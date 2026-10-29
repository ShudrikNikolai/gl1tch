import { ApiError } from '~/api/errors'
import { clean } from './render'

// Подгони под реальные ErrorInfo.reason бэкенда. Известен только USER_BANNED.
const REASONS: Record<string, string> = {
  USER_BANNED: 'ACCESS REVOKED',
  INVALID_CREDENTIALS: 'ACCESS DENIED: wrong username or password',
  USERNAME_TAKEN: 'username is already taken',
}

export function describeError(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.reason && REASONS[e.reason]) return REASONS[e.reason]!
    if (e.status === 401) return 'ACCESS DENIED'
    if (e.status === 403) return 'FORBIDDEN'
    if (e.status === 404) return 'NOT FOUND'
    return clean(e.message) // сообщение пришло с сервера
  }
  if (e instanceof TypeError) return 'CONNECTION LOST' // сеть недоступна
  return 'UNEXPECTED ERROR'
}
