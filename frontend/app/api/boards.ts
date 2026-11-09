import type { ApiClient } from './client'
import type {
  Board, OrderByThread, Post, Thread, ThreadCategory, ThreadStatus, Visibility,
} from './boards-types'
import type { PageInfo } from './types'

const enc = encodeURIComponent

export interface ThreadsQuery {
  boardId: string
  category?: ThreadCategory
  status?: ThreadStatus
  search?: string
  authorUserId?: string
  orderBy?: OrderByThread
  pageSize?: number
  pageToken?: string
}

export interface PageQuery {
  pageSize?: number
  pageToken?: string
}

export function createBoardsApi(api: ApiClient) {
  const boards = {
    list: async () => (await api.get<{ boards?: Board[] }>('/v1/boards')).boards ?? [],
    get: async (id: string) => (await api.get<{ board: Board }>(`/v1/boards/${enc(id)}`)).board,
    create: async (b: {
      slug: string
      title: string
      description: string
      visibility: Visibility
      sortOrder: number
    }) => (await api.post<{ board: Board }>('/v1/boards', b)).board,
  }

  const threadAction = (id: string, action: string, body: unknown = {}) =>
    api.patch<{ thread: Thread }>(`/v1/threads/${enc(id)}/${action}`, body).then((r) => r.thread)

  const threads = {
    list: (q: ThreadsQuery) =>
      api.get<{ threads?: Thread[]; page?: PageInfo }>('/v1/threads', {
        boardId: q.boardId,
        category: q.category,
        status: q.status,
        search: q.search,
        authorUserId: q.authorUserId,
        orderBy: q.orderBy,
        'page.pageSize': q.pageSize,
        'page.pageToken': q.pageToken,
      }),
    get: async (id: string) => (await api.get<{ thread: Thread }>(`/v1/threads/${enc(id)}`)).thread,
    create: (t: {
      boardId: string
      title: string
      body: string
      category?: ThreadCategory
      visibility?: Visibility
    }) => api.post<{ thread: Thread; firstPost?: Post }>('/v1/threads', t),
    // пустые поля и UNSPECIFIED означают «не менять»
    update: async (
      id: string,
      patch: { title?: string; category?: ThreadCategory; visibility?: Visibility },
    ) => (await api.patch<{ thread: Thread }>(`/v1/threads/${enc(id)}`, patch)).thread,
    lock: (id: string, reason?: string) => threadAction(id, 'lock', { reason: reason ?? '' }),
    unlock: (id: string, reason?: string) => threadAction(id, 'unlock', { reason: reason ?? '' }),
    pin: (id: string) => threadAction(id, 'pin'),
    unpin: (id: string) => threadAction(id, 'unpin'),
    remove: (id: string) => api.del(`/v1/threads/${enc(id)}`),
  }

  const posts = {
    list: (threadId: string, p: PageQuery = {}) =>
      api.get<{ posts?: Post[]; page?: PageInfo }>('/v1/posts', {
        threadId,
        'page.pageSize': p.pageSize,
        'page.pageToken': p.pageToken,
      }),
    create: async (p: { threadId: string; parentPostId?: string; body: string }) =>
      (await api.post<{ post: Post }>('/v1/posts', p)).post,
    get: async (id: string) => (await api.get<{ post: Post }>(`/v1/posts/${enc(id)}`)).post,
    update: async (id: string, body: string) =>
      (await api.patch<{ post: Post }>(`/v1/posts/${enc(id)}`, { body })).post,
    remove: (id: string) => api.del(`/v1/posts/${enc(id)}`),
  }

  return { boards, threads, posts }
}
