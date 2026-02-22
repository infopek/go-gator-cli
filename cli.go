package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go-blog-aggregator/internal/config"
	"go-blog-aggregator/internal/database"
	"go-blog-aggregator/internal/rss"
	"time"
)

const DEFAULT_LIMIT = 2

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if command, exists := c.cmds[cmd.name]; exists {
		return command(s, cmd)
	}

	return errors.New("unknown command")
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("a username is required")
	}

	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return errors.New("user does not exist")
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}

	fmt.Printf("you logged in as %s\n", name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("a username is required")
	}

	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return errors.New("user already exists")
	}
	params := database.CreateUserParams{
		uuid.New(),
		time.Now(),
		time.Now(),
		name,
	}

	newUser, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return err
	}
	s.cfg.SetUser(name)
	fmt.Printf("user created with data: %s\n", newUser)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetTable(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func handlerList(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		fmt.Printf(" * %s", user.Name)
		if s.cfg.CurrentUserName == user.Name {
			fmt.Print(" (current)")
		}
		fmt.Print("\n")
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("a time between requests is required (e.g. '1s', '1m', '1h')")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	const url = "https://blog.boot.dev/index.xml"
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <- ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("a name and an url is required")
	}

	name := cmd.args[0]
	url := cmd.args[1]
	params := database.CreateFeedParams{
		uuid.New(),
		time.Now(),
		time.Now(),
		name,
		url,
		user.ID,
	}
	newFeed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	feedFollowParams := database.CreateFeedFollowParams{
		uuid.New(),
		time.Now(),
		time.Now(),
		user.ID,
		newFeed.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", newFeed)

	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feedsInfo, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, info := range feedsInfo {
		fmt.Printf("Feed: %s\n", info.Feed)
		fmt.Printf("URL: %s\n", info.Url)
		fmt.Printf("Created By: %s\n", info.Username)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("a url is required")
	}

	url := cmd.args[0]
	feedInfo, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		uuid.New(),
		time.Now(),
		time.Now(),
		user.ID,
		feedInfo.ID,
	}
	feeds, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", feeds[0].FeedName)
	fmt.Printf("%s\n", user.Name)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	res, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	for _, row := range res {
		fmt.Printf("%s\n", row.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("a url is required")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	err = s.db.UnfollowFeed(context.Background(), database.UnfollowFeedParams {
		user.ID,
		feed.ID,
	})	

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := DEFAULT_LIMIT
	if len(cmd.args) != 0 {
		if val, err := strconv.Atoi(cmd.args[0]); err == nil {
			limit = val
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams {
		UserID: user.ID,
		Limit: int32(limit),
	})
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("title: %s\n", post.Title)
	}

	return nil
}

func scrapeFeeds(s *state) error {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams {
		pq.NullTime {
			time.Now(),
			true,
		},
		time.Now(),
		feedToFetch.ID,
	})
	if err != nil {
		return err
	}

	feed, err := rss.FetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return err
	}

	for _, feedItem := range feed.Channel.Item {
		parsedDate, err := time.Parse(time.RFC1123Z, feedItem.PubDate)
		if err != nil {
			return err
		}

		err = s.db.CreatePost(context.Background(), database.CreatePostParams {
			uuid.New(),
			time.Now(),
			time.Now(),
			feedItem.Title,
			feedItem.Link,
			feedItem.Description,
			parsedDate,
			feedToFetch.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue	// ignore duplicate key error
			}
			return err
		}
	}

	return nil
}
