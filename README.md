# Gator CLI
A guided project on [Boot.Dev](https://www.boot.dev/) with which you can aggregate blogs, their posts, add RSS feeds, and so on.

## Requirements
- Go (v1.25.4)
- PostgreSQL (v15 or later)

## Installing Gator
First, clone the repository:
```bash
git clone https://github.com/infopek/go-gator-cli.git
cd go-gator-cli
```

Gator reads a `json` file from the user's home which contains information about the database. You need a connection
string to an existing database, or just create a new one locally. 
Set up the config file:
```bash
echo '{"db_url":"postgres://<USERNAME>:<PASSWORD>@<IP>:<PORT>/<DB_NAME>?sslmode=disable"}' > ~/.gatorconfig.json
```
You'll need to fill in the blanks with you actual credentials.

Install:
```bash
go install .
```

## Usage
Here are some examples on how to use gator:
```bash
# Register a user
./go-blog-aggregator register user1

# Add a blog to our follow list
./go-blog-aggregator addfeed https://feed.example.com

# See followed blogs
./go-blog-aggregator following

# Aggregate posts from followed content every 3 seconds
./go-blog-aggregator agg 3s

# See the four most recent posts
./go-blog-aggregator browse 4
```

Possible commands:
- `login`
- `register`
- `reset`
- `users`
- `agg`
- `addfeed`
- `feeds`
- `follow`
- `following`
- `unfollow`
- `browse`

## Future Ideas
- Sorting and filtering options to the `browser` command
- Pagination to the `browse` command
- Concurreny to the `agg` command for more frequent fetches
- `search` command that allows for fuzzy searching of posts
- Bookmarking or liking posts
- TUI that allows users to select a post in the terminal and view it in a more readable format
- HTTP API (and authc/authz) that allows other users to interact with the service remotely
- Service manager that keeps the `agg` command running in the background and restarts if it crashes
