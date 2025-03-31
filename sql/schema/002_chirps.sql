-- +goose Up
create table chirps (
	id uuid primary key,
	body text not null,
	created_at timestamp not null,
	updated_at timestamp not null,
	user_id uuid not null references users(id) on delete cascade
);

-- +goose Down
drop table chirps;
