import type { ApiClient } from './client'
import type { TokenPair } from './types'

interface AuthResult {
  userId: string
  tokens: TokenPair
}

export function createAuthApi(api: ApiClient) {
  async function authenticate(path: string, username: string, password: string) {
      console.log('INFO >>>', {
          path,
          username
      })
    const res = await api.post<AuthResult>(path, { username, password }, false)
    api.setTokens(res.tokens)
    return res.userId
  }
  

  return {
    register: (username: string, password: string) =>
      authenticate('/auth/register', username, password),
    login: (username: string, password: string) =>
      authenticate('/auth/login', username, password),

    async logout() {
      try {
        // кука уходит сама; сервер гасит её через Set-Cookie
        await api.post('/auth/logout', {}, false)
      } finally {
        api.clearTokens() // локально выходим при любом исходе
      }
    },
  }
}
