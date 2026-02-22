-- name: CreateFeed :one
insert into feeds (id, created_at, updated_at, name, url, user_id)
values (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
)
returning *;

-- name: GetFeeds :many
select feeds.name as feed, url, users.name as username
from feeds
inner join users
on feeds.user_id = users.id;

-- name: GetFeed :one
select *
from feeds
where url = $1;

-- name: MarkFeedFetched :exec
update feeds
set last_fetched_at = $1, updated_at = $2
where id = $3;

-- name: GetNextFeedToFetch :one
select *
from feeds
order by last_fetched_at asc nulls first;

