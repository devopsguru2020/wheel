package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/unity26org/wheel/commons/notify"
	"github.com/unity26org/wheel/generator"
	"github.com/unity26org/wheel/help"
	"github.com/unity26org/wheel/version"
)

func IsGoInstalled() bool {
	var out bytes.Buffer

	cmd := exec.Command("go", "version")
	cmd.Stdout = &out
	err := cmd.Run()

	return err == nil
}

func CheckDependences() error {
	var out bytes.Buffer
	var hasDependence bool

	notify.Simpleln("Checking dependences...")

	cmd := exec.Command("go", "list", "...")
	cmd.Stdout = &out

	cmd.Run()
	// if err := cmd.Run(); err != nil {
	//   fmt.Println("error:", err)
	// }

	installedDependences := strings.Split(out.String(), "\n")

	requiredDependences := []string{"github.com/jinzhu/gorm", "gopkg.in/yaml.v2", "github.com/gorilla/mux", "github.com/dgrijalva/jwt-go", "github.com/satori/go.uuid", "github.com/lib/pq", "golang.org/x/crypto/bcrypt"}
	for _, requiredDependence := range requiredDependences {
		hasDependence = false
		for _, installedDependence := range installedDependences {
			hasDependence = (requiredDependence == installedDependence)
			if hasDependence {
				break
			}
		}

		if !hasDependence {
			notify.Simple(fmt.Sprintf("         package %s was not found, installing...", requiredDependence))
			cmd := exec.Command("go", "get", requiredDependence)
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil {
				return err
			}

			notify.Simpleln(fmt.Sprintf("         package %s was successfully installed", requiredDependence))
		} else {
			notify.Simpleln(fmt.Sprintf("         package %s was found", requiredDependence))
		}
	}

	return nil
}

func optionsAreValid(args []string) bool {
	b := true

	for index, value := range args {
		if (index > 2) && (value != "-G" && value != "--skip-git") {
			b = false
			break
		}
	}

	return b
}

func checkGitIgnore(args []string) bool {
	b := true

	for index, value := range args {
		if index > 2 && value == "-G" || value == "--skip-git" {
			b = false
			break
		}
	}

	return b
}

func handleNewApp(args []string) error {
	var options = make(map[string]interface{})

	err := checkIsGoInstalled()
	if err != nil {
		return err
	}

	if !optionsAreValid(args) {
		return errors.New("invalid option. Run \"wheel --help\" for details")
	} else {
		preOptions := strings.Split(os.Args[2], "/")

		options["app_name"] = preOptions[len(preOptions)-1]
		options["app_repository"] = os.Args[2]
		options["git_ignore"] = checkGitIgnore(os.Args)

		notify.Simpleln("Generating new app...")
		return generator.NewApp(options)
	}
}

func isResourceNameValid(name string) bool {
	var regexpInvalidChar = regexp.MustCompile(`[^\w]`)
	return !regexpInvalidChar.MatchString(name)
}

func buildGenerateOptions(args []string) (map[string]bool, error) {
	var options = make(map[string]bool)
	var subject string
	var err error

	if len(args) > 2 {
		subject = args[2]
	} else {
		subject = "invalid_subject"
	}

	options["model"] = false
	options["entity"] = false
	options["view"] = false
	options["handler"] = false
	options["routes"] = false
	options["migrate"] = false
	options["authorize"] = false

	switch subject {
	case "scaffold":
		options["model"] = true
		options["entity"] = true
		options["view"] = true
		options["handler"] = true
		options["routes"] = true
		options["migrate"] = true
		options["authorize"] = true
		if len(args) < 4 || !isResourceNameValid(args[3]) {
			err = errors.New("invalid scaffold name")
		}
	case "model":
		options["model"] = true
		options["entity"] = true
		options["migrate"] = true
		if len(args) < 4 || !isResourceNameValid(args[3]) {
			err = errors.New("invalid model name")
		}
	case "handler":
		options["handler"] = true
		options["routes"] = true
		options["authorize"] = true
		if len(args) < 4 || !isResourceNameValid(args[3]) {
			err = errors.New("invalid handler name")
		}
	case "entity":
		options["entity"] = true
		options["migrate"] = true
		if len(args) < 4 || !isResourceNameValid(args[3]) {
			err = errors.New("invalid entity name")
		}
	case "migration":
		// wheel g migration add_total_to_users total:integer
		// wheel g migration remove_total_from_users
		options["migrate"] = true
		if len(args) < 4 || !isResourceNameValid(args[3]) {
			err = errors.New("invalid migration name")
		}
	default:
		err = errors.New("invalid generate subject. Run \"wheel --help\" for details")
	}

	return options, err
}

func handleGenerateNewCrud(args []string, options map[string]bool) error {
	var columns []string

	for index, value := range args {
		if index <= 3 {
			continue
		} else {
			columns = append(columns, value)
		}
	}

	notify.Simpleln("Generating new CRUD...")
	return generator.NewCrud(args[3], columns, options)
}

func handleGenerate(args []string) error {
	var options map[string]bool
	var err error

	err = checkIsGoInstalled()
	if err != nil {
		return err
	}

	options, err = buildGenerateOptions(args)
	if err != nil {
		return err
	}

	return handleGenerateNewCrud(args, options)
}

func handleHelp() {
	notify.Simpleln(help.Content)
}

func handleVersion() {
	notify.Simpleln(version.Content)
}

func checkIsGoInstalled() error {
	if !IsGoInstalled() {
		return errors.New("\"Go\" seems not installed")
	} else {
		notify.Simpleln("\"Go\" seems installed")
		return CheckDependences()
	}
}

func main() {
	command := ""

	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	if command == "new" || command == "n" {
		if err := handleNewApp(os.Args); err != nil {
			notify.Error(err)
		}
	} else if command == "generate" || command == "g" {
		if err := handleGenerate(os.Args); err != nil {
			notify.Error(err)
		}
	} else if command == "--help" || command == "-h" {
		handleHelp()
	} else if command == "--version" || command == "-v" {
		handleVersion()
	} else {
		notify.ErrorJustified("invalid argument. Use \"new\" or \"generate\". Run \"wheel --help\" for details", 0)
		handleHelp()
	}
}
