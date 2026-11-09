import { defineStore } from 'pinia'
import { boardsApi, notificationsApi } from '~/api'
import type {
  Author, Board, OrderByThread, Post, Thread, ThreadCategory, ThreadStatus, Visibility,
} from '~/api/boards-types'

const PAGE = 20

export interface ThreadFilter {
  category?: ThreadCategory
  status?: ThreadStatus
  search?: string
  authorUserId?: string
  orderBy?: OrderByThread
}

type Pager =
  | { kind: 'threads'; token: string; filter: ThreadFilter }
  | { kind: 'posts'; token: string }

/** Состояние навигации по доскам: где мы (доска/тема) и что показывали последним. */
export const useBoardsStore = defineStore('boards', () => {
  const board = ref<Board | null>(null)
  const thread = ref<Thread | null>(null)
  /** Последний пост, пришедший по SSE; терминал печатает его и сбрасывает. */
  const livePost = ref<Post | null>(null)

  // нереактивное состояние: в шаблонах не нужно
  let boards: Board[] = []
  let threadList: Thread[] = []
  const postsByNumber = new Map<number, Post>()
  let pager: Pager | null = null
  const authors = new Map<string, string>() // username (lower) -> userId
  let liveAbort: AbortController | null = null
  let liveThreadId: string | null = null

  const path = computed(() => {
    if (!board.value) return '~'
    const t = thread.value
    if (!t) return `~/${board.value.slug}`
    return `~/${board.value.slug}/${(t.slug || t.threadId).slice(0, 18)}`
  })

  function remember(a?: Author) {
    if (a?.username && a.userId) authors.set(a.username.toLowerCase(), a.userId)
  }

  function storePost(p: Post) {
    remember(p.author)
    if (p.number !== undefined) postsByNumber.set(p.number, p)
  }

  // ---- доски

  async function loadBoards(force = false): Promise<Board[]> {
    if (force || boards.length === 0) {
      const list = await boardsApi.boards.list()
      boards = [...list].sort(
        (a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0) || a.title.localeCompare(b.title),
      )
      boards.forEach((b) => remember(b.lastPostAuthor))
    }
    return boards
  }

  const boardSlugs = () => boards.map((b) => b.slug)

  async function findBoard(arg: string): Promise<Board | null> {
    const list = await loadBoards()
    if (/^\d+$/.test(arg)) return list[Number(arg) - 1] ?? list.find((b) => b.boardId === arg) ?? null
    return list.find((b) => b.slug === arg || b.boardId === arg) ?? null
  }

  function stopLive() {
    liveAbort?.abort()
    liveAbort = null
    liveThreadId = null
  }

  function startLive(threadID: string) {
    if (!threadID || (liveThreadId === threadID && liveAbort)) return
    stopLive()

    const ctrl = new AbortController()
    liveAbort = ctrl
    liveThreadId = threadID

    void notificationsApi
      .streamThread(
        threadID,
        async (event) => {
          try {
            const p = await boardsApi.posts.get(event.postId)
            if (p.threadId !== threadID) return
            storePost(p)
            livePost.value = p
          } catch {
            // пост мог быть удалён между событием и запросом
          }
        },
        ctrl.signal,
      )
      .catch(() => undefined)
      .finally(() => {
        if (liveAbort === ctrl) {
          liveAbort = null
          liveThreadId = null
        }
      })
  }

  function enterBoard(b: Board) {
    stopLive()
    board.value = b
    thread.value = null
    threadList = []
    postsByNumber.clear()
    pager = null
  }

  async function createBoard(input: Parameters<typeof boardsApi.boards.create>[0]) {
    const b = await boardsApi.boards.create(input)
    await loadBoards(true)
    return b
  }

  // ---- темы

  async function fetchThreads(filter: ThreadFilter | null, more = false) {
    if (!board.value) throw new Error('no board')
    const f = more && pager?.kind === 'threads' ? pager.filter : (filter ?? {})
    const res = await boardsApi.threads.list({
      ...f,
      boardId: board.value.boardId,
      pageSize: PAGE,
      pageToken: more && pager?.kind === 'threads' ? pager.token : undefined,
    })
    const items = res.threads ?? []
    items.forEach((t) => {
      remember(t.author)
      remember(t.lastPostAuthor)
    })
    const offset = more ? threadList.length : 0
    threadList = more ? [...threadList, ...items] : items
    const next = res.page?.nextPageToken
    pager = res.page?.hasMore && next ? { kind: 'threads', token: next, filter: f } : null
    return { items, offset, hasMore: pager !== null }
  }

  /** «3» берёт тему из последнего списка, всё остальное считается id. */
  function resolveThreadId(arg: string): string {
    if (/^\d{1,3}$/.test(arg) && threadList[Number(arg) - 1]) return threadList[Number(arg) - 1]!.threadId
    return arg
  }

  async function fetchPosts(more = false) {
    const t = thread.value
    if (!t) throw new Error('no thread')
    const res = await boardsApi.posts.list(t.threadId, {
      pageSize: PAGE,
      pageToken: more && pager?.kind === 'posts' ? pager.token : undefined,
    })
    const items = res.posts ?? []
    items.forEach(storePost)
    const next = res.page?.nextPageToken
    pager = res.page?.hasMore && next ? { kind: 'posts', token: next } : null
    return { items, hasMore: pager !== null }
  }

  async function openThread(id: string) {
    const t = await boardsApi.threads.get(id)
    remember(t.author)
    remember(t.lastPostAuthor)
    if (board.value?.boardId !== t.boardId) {
      const owner = (await loadBoards()).find((b) => b.boardId === t.boardId)
      if (owner) board.value = owner
    }
    thread.value = t
    postsByNumber.clear()
    pager = null
    livePost.value = null
    const page = await fetchPosts(false)
    startLive(t.threadId)
    return { thread: t, ...page }
  }

  async function createThread(input: {
    title: string
    body: string
    category?: ThreadCategory
    visibility?: Visibility
  }) {
    if (!board.value) throw new Error('no board')
    const res = await boardsApi.threads.create({ boardId: board.value.boardId, ...input })
    return res.thread
  }

  async function updateTitle(title: string) {
    thread.value = await boardsApi.threads.update(thread.value!.threadId, { title })
    return thread.value
  }

  async function moderate(action: 'lock' | 'unlock' | 'pin' | 'unpin', reason?: string) {
    const id = thread.value!.threadId
    thread.value =
      action === 'lock' || action === 'unlock'
        ? await boardsApi.threads[action](id, reason)
        : await boardsApi.threads[action](id)
    return thread.value
  }

  async function deleteThread() {
    await boardsApi.threads.remove(thread.value!.threadId)
    stopLive()
    thread.value = null
    postsByNumber.clear()
    pager = null
  }

  // ---- посты

  const postByNumber = (n: number) => postsByNumber.get(n)
  const numberOfPost = (id: string) =>
    [...postsByNumber.values()].find((p) => p.postId === id)?.number

  async function createPost(body: string, parent?: Post) {
    const threadID = thread.value!.threadId
    const p = await boardsApi.posts.create({
      threadId: threadID,
      parentPostId: parent?.postId,
      body,
    })
    storePost(p)
    // До первого поста сервер мог ответить 403; после своего поста поток пора перезапустить.
    stopLive()
    startLive(threadID)
    return p
  }

  async function editPost(p: Post, body: string) {
    const np = await boardsApi.posts.update(p.postId, body)
    storePost(np)
    return np
  }

  async function deletePost(p: Post) {
    await boardsApi.posts.remove(p.postId)
    storePost({ ...p, deleted: true, body: '' })
  }

  // ---- прочее

  const pagerKind = () => pager?.kind ?? null
  const threadOffset = () => threadList.length

  /** Ник → userId для тех, кого мы видели на досках (поиска по нику на бэкенде нет). */
  const knownUserId = (nick: string) => authors.get(nick.toLowerCase())

  function back(): 'thread' | 'board' | 'root' {
    pager = null
    if (thread.value) {
      stopLive()
      thread.value = null
      livePost.value = null
      postsByNumber.clear()
      return 'board'
    }
    if (board.value) {
      board.value = null
      threadList = []
      return 'root'
    }
    return 'root'
  }

  function reset() {
    stopLive()
    board.value = null
    thread.value = null
    livePost.value = null
    boards = []
    threadList = []
    postsByNumber.clear()
    pager = null
    authors.clear()
  }

  return {
    board, thread, livePost, path,
    loadBoards, boardSlugs, findBoard, enterBoard, createBoard,
    fetchThreads, resolveThreadId, fetchPosts, openThread, createThread,
    updateTitle, moderate, deleteThread,
    postByNumber, numberOfPost, createPost, editPost, deletePost,
    pagerKind, threadOffset, knownUserId, back, reset,
    startLive, stopLive,
  }
})
