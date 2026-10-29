export const ERROR_INFO_TYPE = 'type.googleapis.com/google.rpc.ErrorInfo'

/** Значения ErrorInfo.reason, которые клиент понимает. Остальные добавим по мере появления. */
export const Reason = {
  USER_BANNED: 'USER_BANNED',
} as const

interface ErrorInfo {
  reason?: string
  domain?: string
  metadata?: Record<string, string>
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly reason?: string,
    readonly metadata: Record<string, string> = {},
  ) {
    super(message)
    this.name = 'ApiError'
  }

  get isBan(): boolean {
    return this.reason === Reason.USER_BANNED
  }

  /** Срок окончания бана из metadata.banned_until (RFC3339), если сервер его прислал. */
  get bannedUntil(): Date | null {
    const raw = this.metadata.banned_until
    if (!raw) return null
    const date = new Date(raw)
    return Number.isNaN(date.getTime()) ? null : date
  }
}

/** Тело ошибки grpc-gateway: { code, message, details: [{ '@type', reason, metadata }] }. */
export function parseApiError(status: number, body: unknown): ApiError {
  const b = (body && typeof body === 'object' ? body : {}) as {
    message?: string
    details?: Array<Record<string, unknown>>
  }
  const info = b.details?.find((d) => d['@type'] === ERROR_INFO_TYPE) as unknown as
    | ErrorInfo
    | undefined

  return new ApiError(status, b.message ?? `HTTP ${status}`, info?.reason, info?.metadata)
}
