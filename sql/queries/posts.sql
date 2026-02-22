-- name: CreatePost :exec
insert into posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
values ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
select posts.*
from posts
inner join feeds
on posts.feed_id = feeds.id
inner join feed_follows
on feed_follows.feed_id = feeds.id
where feed_follows.user_id = $1
order by published_at desc
limit $2;
