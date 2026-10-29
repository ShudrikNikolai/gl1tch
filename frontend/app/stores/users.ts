import { defineStore } from 'pinia'
import { usersApi } from '~/api'
import type { ListUsersParams } from '~/api/users'

// Кэша нет: просто запросы. Стор оставлен как точка расширения.
export const useUsersStore = defineStore('users', () => {
  return {
    get: (userId: string) => usersApi.get(userId),
    list: (params?: ListUsersParams) => usersApi.list(params),
    reset() {
      // нечего сбрасывать, пока нет кэша
    },
  }
})
