package users

import "context"

func (s Service) Create(ctx context.Context, userID, username string) error {
	return s.users.Create(ctx, userID, username)
}
