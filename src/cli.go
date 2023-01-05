package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:                 "nginx-auth-server",
	Usage:                "simple authentication server designed to be used in conjunction with nginx 'http_auth_request_module'. Written in Go.",
	Version:              "0.0.5",
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
						&cli.StringFlag{
							Name:    "password",
							Aliases: []string{"p"},
						},
						&cli.BoolFlag{
							Name:    "otp",
							Aliases: []string{"o"},
						},
					},
					Action: func(cCtx *cli.Context) error {
						// TODO: check case-insensitive for existing users and prevent account creation
						// TODO: check for existing LDAP users and show warning
						// TODO: let user input password from terminal instead of using a parameter to avoid plain-text passwords in bash/zsh history
						addUser(cCtx.String("username"), cCtx.String("password"), cCtx.Bool("otp"))
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
