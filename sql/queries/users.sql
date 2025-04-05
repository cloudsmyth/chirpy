-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
returning *;

-- name: GetUserByEmail :one
select * from users where email = $1;

-- name: GetUserById :one
select * from users where id = $1;

-- name: UpdateUserById :one
update users
set (email, hashed_password, updated_at) = ($1, $2, NOW())
where id = $3
returning *;

-- name: UpgradeUserById :one
update users
set (is_chirpy_red, updated_at) = ($1, NOW())
where id = $2
returning *;
