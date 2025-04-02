-- name: CreateRefreshToken :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at)
values (
	$1,
	NOW(),
	NOW(),
	$2,
	$3
)
returning *;

-- name: GetRefreshByToken :one
select * from refresh_tokens where token = $1;

-- name: RevokeRefreshByToken :one
update refresh_tokens
set (revoked_at, updated_at) = (now(), now())
where token = $1
returning *;
