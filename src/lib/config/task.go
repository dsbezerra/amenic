package config

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dsbezerra/amenic/src/lib/persistence/models"
	"github.com/dsbezerra/amenic/src/lib/util/fileutil"
	"github.com/dsbezerra/amenic/src/lib/util/stringutil"
)

// LoadTasks from .tasks file.
func LoadTasks() (map[string]models.Task, error) {
	result := make(map[string]models.Task)

	log.Println("Loading .tasks file...")
	path, err := filepath.Abs("./.tasks")
	if err != nil {
		return nil, err
	}

	handler := &fileutil.TextFileHandler{}
	err = handler.StartFile(path)
	if err != nil {
		return nil, err
	}

	defer handler.CloseFile()

	var task *models.Task
	for {
		line, found := handler.ConsumeNextLine()
		if !found {
			break
		}

		if line[0] != '#' {
			word := strings.TrimSpace(line)
			if word == "task" {
				if task != nil && task.IsValid() {
					ok, _ := task.GenerateID()
					if ok {
						result[task.ID] = *task
					} else {
						log.Printf("couldn't generate ID for task %s", task.Name)
					}
				}

				task = &models.Task{}
			} else {
				lhs, rhs := stringutil.BreakByToken(line, '=')
				if lhs == "" {
					log.Println("syntax is wrong. expected to find something to the left of = token")
				}

				if rhs == "" {
					log.Println("syntax is wrong. expected to find something to the right of = token")
				}

				if lhs != "" && rhs != "" {
					switch lhs {
					case "service":
						task.Service = rhs
					case "name":
						task.Name = rhs
					case "description":
						task.Description = rhs
					case "cron":
						task.Cron = strings.Split(rhs, ",")
					case "type":
						task.Type = rhs
					case "args":
						task.Args = strings.Split(rhs, ",")
					case "enabled":
						v, _ := strconv.ParseBool(rhs)
						task.Enabled = v
					case "run_at_start":
						v, _ := strconv.ParseBool(rhs)
						task.RunAtStart = v
					}
				}
			}
		}
	}

	// Insert last one.
	if task != nil && task.IsValid() {
		ok, _ := task.GenerateID()
		if ok {
			result[task.ID] = *task
		} else {
			log.Printf("couldn't generate ID for task %s", task.Name)
		}
	}

	return result, nil
}
