package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var GinMode = gin.DebugMode

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	app := &cli.App{
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run application",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "daemon",
						Aliases: []string{"d"},
						Usage:   "run server in daemon mode",
					},
				},
				Action: func(cCtx *cli.Context) error {
					runGin(cCtx.Bool("daemon"))

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
						},
						Action: func(cCtx *cli.Context) error {
							addUser(cCtx.String("username"), cCtx.String("password"))
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
							fmt.Println(users)

							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runGin(daemonMode bool) {
	gin.SetMode(GinMode)

	router := gin.Default()
	router.LoadHTMLGlob("public/*.html")
	router.Static("/css", "public/css")

	router.GET("/login", login)
	router.POST("/login", processLoginForm)

	address := GetListenAddress() + ":" + strconv.Itoa(GetListenPort())

	err := router.Run(address)

	if err != nil {
		panic(fmt.Errorf("fatal error trying to launch the webserver: %w", err))
	}
}

func addUser(username string, password string) {
	if username == "" {
		log.Fatalln("invalid username")
	}

	if password == "" {
		generatedPassword := GeneratePassword(8, 0, 1, 1)

		fmt.Printf("no password given, generated password for user '%s': '%s'\n", username, generatedPassword)

		addUser(username, generatedPassword)
	} else if err := CheckPasswordRequirements(password); err != nil {
		log.Fatalf("password does not meet minimum requirements: %s", err)
	} else {
		hash, err := GenerateHash(password)

		if err != nil {
			log.Fatalf("could not salt and hash password: %s", err)
		}

		user := User{
			Username: username,
			Password: hash,
		}

		err = CreateUser(&user)

		if err != nil {
			log.Fatalf("could not save user to database: %s", err)
		}
	}
}

func removeUser(username string) {
	err := RemoveUser(username)

	if err != nil {
		log.Fatalf("could not remove user from database: %s", err)
	} else {
		fmt.Printf("user with username '%s' has been removed", username)
	}
}

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func processLoginForm(c *gin.Context) {
	username := c.PostForm("inputUsername")
	password := c.PostForm("inputPassword")

	user := GetUserByUsername(username)

	if user == nil {
		c.AbortWithStatus(401)
	} else {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			c.AbortWithStatus(401)
		} else {
			cookie := GenerateAuthCookie(user.Username)
			err := CreateCookie(&cookie)

			if err != nil {
				log.Println(err)
				c.AbortWithStatus(500)
			} else {
				http.SetCookie(c.Writer, &http.Cookie{
					Name:    cookie.Name,
					Value:   cookie.Value,
					Expires: cookie.Expires,
					Domain:  cookie.Domain,
				})

				// todo: redirect
			}
		}
	}
}
