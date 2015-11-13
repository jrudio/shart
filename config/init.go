package config

// package main

import (
	"errors"
	"fmt"
	"github.com/jeffail/gabs"
	"io/ioutil"
	"log"
	"os"
)

// Uncomment to test package
// TODO: Learn how to test packages with out using main()
// func main() {
// 	// fields, initErr := Init()
// 	_, initErr := Init()

// 	if initErr != nil {
// 		log.Fatal(initErr)
// 	}

// fmt.Println(fields.Path("couchpotato.host").String())
// }

var (
	configFileName = "config.json"
	defaultConfig  = `{
	"couchpotato": {
		"host": "localhost",
		"apiKey": "abc123",
		"username": "bob123",
		"password": "password123"
	},
	"plex": {
		"host": "localhost",
		"apiKey": "abc123",
		"username": "bob123",
		"password": "password123"
	},
	"sonarr": {
		"host": "localhost",
		"apiKey": "abc123",
		"username": "bob123",
		"password": "password123"
	},
	"slack": {
		"token": "abc123",
		"incomingUrl": "http://hooks.slack.com/services",
		"botName": "MediaBot"
	},
	"shart": {
		"port": "3000"
	}
}`
	requiredObjects = [4]string{
		"couchpotato",
		"plex",
		"sonarr",
		"slack", // Special case: check for token & url
	}
	requiredFields = [2]string{
		"host",
		"apiKey",
	}
	slackRequiredFields = [2]string{
		"token",
		"incomingUrl",
	}
)

func Init() (*gabs.Container, error) {
	configErr := createConfig()

	if configErr != nil {
		return nil, configErr
	}

	fields, readConfigErr := readConfig(configFileName)

	if readConfigErr != nil {
		return nil, readConfigErr
	}

	// Return an error if required fields are not present
	// default config statisfies this
	if !hasRequiredFields(fields) {
		return nil, errors.New("Missing one or more required fields in config.json")
	}

	// return fields usable by Go
	// return nil, nil
	return fields, nil
}

func hasRequiredFields(fields *gabs.Container) bool {
	var hasFields bool = true

	// Outer object check
	for ii := 0; ii < len(requiredObjects); ii++ {
		if !hasFields {
			break
		}

		reqObj := requiredObjects[ii]

		children, childrenErr := fields.S(reqObj).ChildrenMap()

		if childrenErr != nil {
			log.Fatal(childrenErr)
		}

		// Handle non-slack fields differently
		if reqObj == "slack" {
			requiredFields = slackRequiredFields
		}

		// Determine whether the required fields are present
		for key, val := range children {
			if !hasFields {
				break
			}

			for jj := 0; jj < len(requiredFields); jj++ {
				// We don't care about this key, so go to next iteration
				if key != requiredFields[jj] {
					continue
				}

				// We are on a required field
				if val.Data().(string) == "" {
					hasFields = false

					break
				}
			}
		}
	}

	return hasFields
}

/**
*	Returns config file.
*	Creates the file if it doesn't exist
 */
func createConfig() error {
	file, err := os.Open(configFileName)

	if err != nil {
		// Create the file
		createdFile, createdErr := os.Create(configFileName)

		if createdErr != nil {
			return createdErr
		}

		defer createdFile.Close()

		// Write defaults to file
		writeDefaultConfig(createdFile)

		file = createdFile
	} else {
		fmt.Println("Config already created")
		defer file.Close()
	}

	return nil
}

/**
*	Returns file as json
 */
func readConfig(fileName string) (*gabs.Container, error) {
	// Read file
	fileBytes, readErr := ioutil.ReadFile(fileName)

	if readErr != nil {
		return nil, readErr
	}

	// Convert to usable map
	parsedJson, parsedJsonErr := gabs.ParseJSON(fileBytes)

	if parsedJsonErr != nil {
		return nil, parsedJsonErr
	}

	return parsedJson, nil
}

func writeDefaultConfig(file *os.File) error {
	bytesWritten, writeErr := file.Write([]byte(defaultConfig))

	if writeErr != nil {
		return writeErr
	}

	if bytesWritten < 1 {
		fmt.Println("There was a problem writing the default config to file")
	} else {
		fmt.Println("Successfully wrote the default config")
	}

	return nil
}
