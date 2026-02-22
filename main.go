package main

import (
	"go-blog-aggregator/internal/config"
	"go-blog-aggregator/internal/database"
	"fmt"
	"os"
	"database/sql"
	_ "github.com/lib/pq"
)


func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("an error occurred while reading config file: %s\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	dbQueries := database.New(db)
	s := state {
		cfg: &cfg,
		db:   dbQueries,
	}

	// Register commands
	commands := commands {
		cmds: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerList)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerListFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))

	// Get and run command 
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("too few arguments\n")
		os.Exit(1)
	}
	cmd := command {
		name: args[1],
		args: args[2:],
	}

	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Printf("error while running command %s: %s\n", cmd.name, err)
		os.Exit(1)
	}

	os.Exit(0)
}

