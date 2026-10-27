package auth

import "context"

// Logout завершает текущую сессию. Повторный logout и неизвестный токен не ошибка.
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	return uc.refresh.Revoke(ctx, refreshToken)
}
