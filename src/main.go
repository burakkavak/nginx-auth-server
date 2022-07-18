package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
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
		Name:                 "nginx-auth-server",
		Usage:                "simple authentication server designed to be used in conjunction with nginx 'http_auth_request_module'. Written in Go.",
		Version:              "0.0.1",
		EnableBashCompletion: true,
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runGin() {
	gin.SetMode(GinMode)

	router := gin.Default()
	templates, _ := template.ParseFS(templateFiles, "templates/*.html")
	router.SetHTMLTemplate(templates)
	router.StaticFS("/nginx-auth-server-static", http.FS(staticFiles))

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

func addUser(username string, password string, otp bool) {
	if username == "" {
		log.Fatalf("invalid username")
	}

	if GetUserByUsername(username) != nil {
		log.Fatalf("user with username %s already exists", username)
	}

	if password == "" {
		generatedPassword := GeneratePassword(8, 0, 1, 1)

		fmt.Printf("no password given, generated password for user '%s': '%s'\n", username, generatedPassword)

		addUser(username, generatedPassword, otp)
	} else if err := CheckPasswordRequirements(password); err != nil {
		log.Fatalf("password does not meet minimum requirements: %s", err)
	} else {
		hash, err := GenerateHash(password)

		if err != nil {
			log.Fatalf("could not salt and hash password: %s", err)
		}

		var encryptedOtpSecret []byte

		if otp {
			otpKey, err := totp.Generate(totp.GenerateOpts{
				Issuer:      GetDomain(),
				AccountName: username,
			})

			if err != nil {
				log.Fatalf("could not create TOTP: %s", err)
			}

			fmt.Printf("TOTP secret key for user '%s': '%s'\n", username, otpKey.Secret())

			output, err := exec.Command("sh", "-c", fmt.Sprintf("qrencode -t UTF8 '%s'", otpKey.URL())).Output()

			fmt.Printf("TOTP URL for user '%s': '%s'\n", username, otpKey.URL())

			if err != nil {
				fmt.Println("install 'qrencode' library to display a QR code.")
			} else {
				fmt.Println(string(output))
			}

			encryptedOtpSecret = Encrypt([]byte(otpKey.Secret()), password)
		}

		user := User{
			Username:  username,
			Password:  hash,
			OtpSecret: encryptedOtpSecret,
		}

		err = CreateUser(&user)

		if err != nil {
			log.Fatalf("could not save user to database: %s", err)
		} else {
			fmt.Printf("user with username '%s' successfully created\n", username)
		}
	}
}

func removeUser(username string) {
	err := RemoveUser(username)

	if err != nil {
		log.Fatalf("could not remove user from database: %s", err)
	} else {
		fmt.Printf("user with username '%s' has been removed\n", username)
	}

	err = DeleteCookiesByUsername(username)

	if err != nil {
		log.Fatalf("could not remove user associated cookies from database for username '%s': %s\n", username, err)
	} else {
		fmt.Printf("user associated cookies for username '%s' have been removed\n", username)
	}
}

func authenticate(c *gin.Context) {
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

	if err != nil || VerifyCookie(cookieValue) != nil {
		c.AbortWithStatus(401)
		return
	} else {
		c.Status(200)
		return
	}

}

func login(c *gin.Context) {
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil && VerifyCookie(cookieValue) == nil {
		// user already authorized
		c.Status(200)
		return
	}

	c.HTML(http.StatusOK, "login.html", nil)
}

func logout(c *gin.Context) {
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

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
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

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
		return
	} else {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			c.AbortWithStatus(401)
			return
		} else {
			if len(user.OtpSecret) != 0 {
				token := c.PostForm("inputTotp")

				secret := Decrypt(user.OtpSecret, password)

				tokenIsValid := totp.Validate(token, string(secret))

				if !tokenIsValid {
					c.AbortWithStatusJSON(401, gin.H{"error": "invalid TOTP"})
					return
				}
			}

			cookie := GenerateAuthCookie(user.Username)
			err := CreateCookie(&cookie)

			if err != nil {
				log.Println(err)
				c.AbortWithStatus(500)
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:    cookie.Name,
				Value:   cookie.Value,
				Expires: cookie.Expires,
				Domain:  cookie.Domain,
			})

			c.Status(200)
		}
	}
}
