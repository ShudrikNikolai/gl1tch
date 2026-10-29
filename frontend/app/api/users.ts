import type { ApiClient } from './client'
import type { ListUsersResponse, User } from './types'

export interface ListUsersParams {
  pageSize?: number
  pageToken?: string
}

export function createUsersApi(api: ApiClient) {
  return {
    me: async () => (await api.get<{ user: User }>('/users/me')).user,
    get: async (userId: string) =>
      (await api.get<{ user: User }>(`/users/${encodeURIComponent(userId)}`)).user,
    // поиска по нику на бэкенде нет: ListUsers умеет только пагинацию
    list: (p: ListUsersParams = {}) =>
      api.get<ListUsersResponse>('/users', {
        'cursor.pageSize': p.pageSize,
        'cursor.pageToken': p.pageToken,
      }),
  }
}
