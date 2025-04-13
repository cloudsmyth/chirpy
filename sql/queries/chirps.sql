-- name: CreateChirp :one
insert into chirps (id, body, created_at, updated_at, user_id)
values (
	gen_random_uuid(),
	$1,
	NOW(),
	NOW(),
	$2
)
returning *;

-- name: GetChirps :many
select * from chirps order by created_at;

-- name: GetChirpById :one
select * from chirps where id = $1;

-- name: DeleteChirpById :exec
delete from chirps 
where id = $1 and user_id = $2;

-- name: GetChirpsByAuthor :many 
select * from chirps where user_id = $1 order by created_at;
