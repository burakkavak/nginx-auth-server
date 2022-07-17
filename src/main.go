package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

//go:embed templates
var templateFiles embed.FS

// TODO: only include compiled files (omit *.ts files)
//go:embed css js
var staticFiles embed.FS

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
			{
				Name:    "cookie",
				Aliases: []string{"c"},
				Usage:   "options for cookie management",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "list all cookies",
						Action: func(cCtx *cli.Context) error {
							cookies := GetCookies()

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
						Action: func(cCtx *cli.Context) error {
							PurgeCookies()

							fmt.Printf("deleted all cookies from database")

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
	templates, _ := template.ParseFS(templateFiles, "templates/*.html")
	router.SetHTMLTemplate(templates)
	router.StaticFS("/static", http.FS(staticFiles))

	router.GET("/auth", authenticate)
	router.GET("/login", login)
	router.POST("/login", processLoginForm)
	router.GET("/logout", logout)

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

func authenticate(c *gin.Context) {
	cookieValue, err := c.Cookie("Auth")

	if err != nil || VerifyCookie(cookieValue) != nil {
		c.AbortWithStatus(401)
		return
	} else {
		c.Status(200)
		return
	}

}

func login(c *gin.Context) {
	cookieValue, err := c.Cookie("Auth")

	if err == nil && VerifyCookie(cookieValue) == nil {
		// user already authorized
		c.Status(200)
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func logout(c *gin.Context) {
	cookieValue, err := c.Cookie("Auth")

	if err == nil && VerifyCookie(cookieValue) == nil {
		cookie := GetCookieByValue(cookieValue)
		err := DeleteCookie(cookieValue)

		if err != nil || cookie == nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("could not delete user cookie from database"))
			return
		} else {
			http.SetCookie(c.Writer, &http.Cookie{
				Name:    cookie.Name,
				Value:   cookie.Value,
				Expires: time.Now(),
				Domain:  cookie.Domain,
			})

			c.String(http.StatusOK, "user successfully logged out")
			return
		}
	} else {
		c.AbortWithError(http.StatusUnauthorized, errors.New("user not logged in"))
	}
}

func processLoginForm(c *gin.Context) {
	cookieValue, err := c.Cookie("Auth")

	if err == nil && VerifyCookie(cookieValue) == nil {
		// user already authorized
		c.Status(200)
		return
	}

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

				c.Redirect(303, c.GetHeader("Referer"))
			}
		}
	}
}
