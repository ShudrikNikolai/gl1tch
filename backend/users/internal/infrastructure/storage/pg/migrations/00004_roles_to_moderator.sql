-- +goose Up
UPDATE users
SET roles = array_replace(roles, 'moder'::user_role, 'moderator'::user_role)
WHERE 'moder'::user_role = ANY(roles);

-- +goose Down
UPDATE users
SET roles = array_replace(roles, 'moderator'::user_role, 'moder'::user_role)
WHERE 'moderator'::user_role = ANY(roles);
