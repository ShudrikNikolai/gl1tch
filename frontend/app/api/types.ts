// Временная ручная заглушка. Когда появятся сгенерированные из proto типы,
// положи их в app/api/gen/ и замени этот файл реэкспортом.

export interface TokenPair {
  accessToken: string
  /** Refresh теперь живёт в httpOnly-куке; в теле может отсутствовать, клиент его не читает. */
  refreshToken?: string
  accessExpiresAt: string // RFC3339
  refreshExpiresAt?: string // RFC3339
}

export type Role = 'ROLE_UNSPECIFIED' | 'ROLE_USER' | 'ROLE_MODERATOR' | 'ROLE_ADMIN'

export type UserStatus =
  | 'USER_STATUS_UNSPECIFIED'
  | 'USER_STATUS_ACTIVE'
  | 'USER_STATUS_BANNED'
  | 'USER_STATUS_DELETED'
  | 'USER_STATUS_PENDING'

export type Gender = 'GENDER_UNSPECIFIED' | 'GENDER_MALE' | 'GENDER_FEMALE'

export interface User {
  userId: string
  username: string
  bio?: string
  role: Role
  status: UserStatus
  gender?: Gender
  createdAt?: string
  updatedAt?: string
  deletedAt?: string
  bannedUntil?: string
}

export interface PageInfo {
  nextPageToken?: string
  hasMore?: boolean
}

export interface ListUsersResponse {
  users?: User[]
  cursor?: PageInfo
}
