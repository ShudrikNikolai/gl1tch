// Временные ручные типы под boards.v1 (proto3 JSON: camelCase, enum строками).
// Замени сгенерированными, когда появятся.
import type { PageInfo } from './types'

export type Visibility =
  | 'VISIBILITY_UNSPECIFIED'
  | 'VISIBILITY_PUBLIC'
  | 'VISIBILITY_AUTHOR'
  | 'VISIBILITY_MODERATOR'

export type ThreadStatus =
  | 'THREAD_STATUS_UNSPECIFIED'
  | 'THREAD_STATUS_OPEN'
  | 'THREAD_STATUS_LOCKED'
  | 'THREAD_STATUS_ARCHIVED'

export type ThreadCategory =
  | 'THREAD_CATEGORY_UNSPECIFIED'
  | 'THREAD_CATEGORY_GENERAL'
  | 'THREAD_CATEGORY_FAQ'
  | 'THREAD_CATEGORY_FLOOD'

export type OrderByThread =
  | 'ORDER_BY_THREAD_UNSPECIFIED'
  | 'ORDER_BY_THREAD_CREATED_AT'
  | 'ORDER_BY_THREAD_LAST_POST_AT'

export interface Author {
  userId: string
  username: string // снимок на момент записи
}

export interface Board {
  boardId: string
  slug: string
  title: string
  description?: string
  visibility?: Visibility
  threadCount?: number
  postCount?: number
  lastPostAuthor?: Author
  lastPostAt?: string
  sortOrder?: number
}

export interface Thread {
  threadId: string
  boardId: string
  title: string
  slug?: string
  author?: Author
  status?: ThreadStatus
  category?: ThreadCategory
  visibility?: Visibility
  postCount?: number
  pinned?: boolean
  lastPostAuthor?: Author
  lastPostAt?: string
  createdAt?: string
  updatedAt?: string
}

export interface Post {
  postId: string
  threadId: string
  parentPostId?: string
  author?: Author
  body?: string
  number?: number
  deleted?: boolean
  editedAt?: string
  createdAt?: string
  updatedAt?: string
}

export type { PageInfo }
