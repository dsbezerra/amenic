package env

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/dsbezerra/amenic/src/lib/util/fileutil"
	"github.com/dsbezerra/amenic/src/lib/util/stringutil"
	"github.com/sirupsen/logrus"
)

// LoadEnv uses a .env file in the the same directory of this file
// to setup any defined environment variables.
//
// Doesn't allow redefinitions of variables.
func LoadEnv() (map[string]string, error) {

	log := logrus.New()

	log.Info("Loading .env file...")
	path, err := filepath.Abs("./.env")
	if err != nil {
		return nil, err
	}

	handler := &fileutil.TextFileHandler{}
	err = handler.StartFile(path)
	if err != nil {
		return nil, err
	}
	defer handler.CloseFile()

	table := map[string]string{}
	for {
		line, found := handler.ConsumeNextLine()
		if !found {
			break
		}

		if line[0] != '#' {
			name, remainder := stringutil.BreakByToken(line, '=')
			if name != "" {
				if remainder != "" {
					if table[name] != "" {
						log.Errorf("Attempt to redefine variable '%s' to '%s'", name, remainder)
					} else {
						log.Infof("Defining env variable '%s' to '%s'", name, remainder)
						table[name] = remainder
						os.Setenv(name, remainder)
					}
				} else {
					log.Errorf("Missing value for variable '%s'.", name)
				}
			}
		}
	}

	return table, err
}

// IsEnvVariableTrue Checks if a given variable is defined to true.
func IsEnvVariableTrue(variable string) bool {
	result := false

	if variable == "" {
		return result
	}

	if v := os.Getenv(variable); v != "" {
		value, err := strconv.ParseBool(v)
		if err != nil {
			panic(err)
		}

		result = value
	}

	return result
}
