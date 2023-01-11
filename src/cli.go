package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

var app = &cli.App{
	Name:                 "nginx-auth-server",
	Usage:                "simple authentication server designed to be used in conjunction with nginx 'http_auth_request_module'. Written in Go.",
	Version:              "0.0.7",
	EnableBashCompletion: true,
	Authors: []*cli.Author{
		{
			Name:  "Burak Kavak",
			Email: "burak@kavak.dev",
		},
	},
	Commands: []*cli.Command{
		{
			Name:  "run",
			Usage: "run application",
			Action: func(cCtx *cli.Context) error {
				runGin()
				return nil
			},
		},
		{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "options for user management",
			Subcommands: []*cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "add a new user",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "username",
							Aliases:  []string{"u"},
							Required: true,
						},
						&cli.BoolFlag{
							Name:    "otp",
							Aliases: []string{"o"},
						},
					},
					Action: func(cCtx *cli.Context) error {
						username := cCtx.String("username")
						existingUser := GetUserByUsernameCaseInsensitive(username)

						if existingUser != nil {
							return fmt.Errorf("error: user with username '%s' already exists\n", existingUser.Username)
						}

						if GetLDAPEnabled() && ldapCheckUserExists(username) {
							fmt.Printf("warning: LDAP user with the same username '%s' already exists\n", username)

							answer := promptYesNo("Do you want to continue creating a local user?")

							if !answer {
								return nil
							}

						}

						password, err := promptPasswordInput()

						if err != nil {
							return errors.New("error: password input failed")
						}

						addUser(username, password, cCtx.Bool("otp"))
						return nil
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove an existing user",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "username",
							Aliases:  []string{"u"},
							Required: true,
						},
					},
					Action: func(cCtx *cli.Context) error {
						removeUser(cCtx.String("username"))
						return nil
					},
				},
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "list all users",
					Action: func(cCtx *cli.Context) error {
						users := GetUsers()

						fmt.Printf("the database contains %d users\n", len(users))

						if len(users) != 0 {
							usersJson, _ := json.MarshalIndent(users, "", "  ")

							fmt.Println(string(usersJson))
						}

						return nil
					},
				},
			},
		},
		{
			Name:    "cookie",
			Aliases: []string{"c"},
			Usage:   "options for cookie management",
			Subcommands: []*cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "list all cookies",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "username",
							Aliases: []string{"u"},
							Usage:   "filter cookies by username",
						},
					},
					Action: func(cCtx *cli.Context) error {
						username := cCtx.String("username")
						var cookies []Cookie

						if username != "" {
							cookies = GetCookiesByUsername(username)
						} else {
							cookies = GetCookies()
						}

						fmt.Printf("the database contains %d cookies\n", len(cookies))

						if len(cookies) != 0 {
							cookiesJson, _ := json.MarshalIndent(cookies, "", "  ")

							fmt.Println(string(cookiesJson))
						}

						return nil
					},
				},
				{
					Name:    "purge",
					Aliases: []string{"p"},
					Usage:   "remove all cookies",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "username",
							Aliases: []string{"u"},
							Usage:   "remove all cookies for user with given username",
						},
					},
					Action: func(cCtx *cli.Context) error {
						username := cCtx.String("username")

						var err error = nil

						if username != "" {
							err = DeleteCookiesByUsername(username)
						} else {
							err = PurgeCookies()
						}

						if err != nil {
							log.Fatalf("error: could not delete cookies: %s", err)
						} else {
							fmt.Printf("deleted all cookies from database\n")
						}

						return nil
					},
				},
			},
		},
	},
}

func promptYesNo(message string) bool {
	answer := "initial"
	var err error = nil

	for !strings.EqualFold(answer, "n") && !strings.EqualFold(answer, "y") &&
		!strings.EqualFold(answer, "no") && !strings.EqualFold(answer, "yes") &&
		answer != "" {
		fmt.Printf("%s [Y/n] ", message)
		reader := bufio.NewReader(os.Stdin)

		answer, err = reader.ReadString('\n')
		answer = strings.TrimSpace(answer)

		if err != nil {
			log.Fatalf("fatal error: could not read user input: %s", err)
		}
	}

	if strings.EqualFold(answer, "n") || strings.EqualFold(answer, "no") {
		return false
	} else {
		return true
	}
}

func promptPasswordInput() (string, error) {
	fmt.Print("Enter password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", err
	}

	fmt.Print("\nRepeat password: ")
	byteRepeatPassword, err := term.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	if bytes.Compare(bytePassword, byteRepeatPassword) != 0 {
		return "", errors.New("error: password mismatch")
	}

	password := string(bytePassword)

	if err = CheckPasswordRequirements(password); err != nil {
		fmt.Printf("%s\n", err)
		return promptPasswordInput()
	} else {
		return strings.TrimSpace(password), nil
	}
}
