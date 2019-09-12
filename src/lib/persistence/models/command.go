package models

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

const (
	// CommandCreateStatic creates all static JSON files or the one(s) specified in args.
	CommandCreateStatic = "create_static"

	// CommandClearStatic clears all static JSON files or the one(s) specified in args.
	CommandClearStatic = "clear_static"

	// CommandClearShowtimes clears all showtimes or showtimes for the specified theaters in args.
	CommandClearShowtimes = "clear_showtimes"

	// CommandStartScraper starts all scrapers or the one(s) specified in args.
	CommandStartScraper = "start_scraper"

	// CommandCheckOpeningMovies checks which movies are new to the movies or the one(s) specified in args.
	CommandCheckOpeningMovies = "check_opening_movies"

	// CommandSyncScores syncs scores of now playing movies
	CommandSyncScores = "sync_scores"
)

// Command ...
type Command struct {
	Name string   `json:"command_name" binding:"required"`
	Args []string `json:"command_args" binding:"required"`
}

// CommandInfo ...
type CommandInfo struct {
	Name         string       `json:"command_name"`
	Description  string       `json:"command_description"`
	PossibleArgs []CommandArg `json:"command_args"`
}

// CommandArg ...
type CommandArg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Hash ...
func (cmd *Command) Hash() string {
	h := md5.New()

	b, err := json.Marshal(cmd)
	if err != nil {
		return ""
	}

	return string(h.Sum(b))
}

// IsValid ...
func (cmd *Command) IsValid() bool {
	result := false

	result = (cmd.Name == CommandCreateStatic ||
		cmd.Name == CommandClearStatic ||
		cmd.Name == CommandClearShowtimes ||
		cmd.Name == CommandStartScraper ||
		cmd.Name == CommandCheckOpeningMovies ||
		cmd.Name == CommandSyncScores)

	return result
}

// ParseArgs ...
func ParseArgs(args []string) map[string]string {
	result := make(map[string]string)

	size := len(args)
	if size%2 != 0 {
		fmt.Println("args size must be even")
	} else {
		for i, arg := range args {
			if arg[0] == '-' {
				result[arg[1:]] = args[i+1]
			}
		}
	}

	return result
}
