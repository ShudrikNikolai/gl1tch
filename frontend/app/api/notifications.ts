import type { ApiClient } from './client'

export interface PostCreatedEvent {
  eventId?: string
  postId: string
  threadId: string
  boardId?: string
  threadTitle?: string
  authorUserId?: string
  authorName?: string
  postNumber?: number
  createdAt?: string
}

export function createNotificationsApi(api: ApiClient) {
  return {
    streamThread(
      threadId: string,
      onPost: (event: PostCreatedEvent) => void,
      signal: AbortSignal,
    ) {
      return api.stream(
        `/notifications/v1/threads/${encodeURIComponent(threadId)}/stream`,
        (event) => {
          if (event.event !== 'post.created' || !event.data) return
          try {
            onPost(JSON.parse(event.data) as PostCreatedEvent)
          } catch {
            // битое событие пропускаем, поток продолжает работать
          }
        },
        signal,
      )
    },
  }
}
