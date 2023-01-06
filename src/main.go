package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

//go:embed templates
var templateFiles embed.FS

//go:embed css/main.css js/app.bundle.js
var staticFiles embed.FS

var GinMode = gin.DebugMode

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

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
	router.GET("/whoami", whoami)

	address := GetListenAddress() + ":" + strconv.Itoa(GetListenPort())

	fmt.Printf("listening and serving HTTP request on %s\n", address)

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
		generatedPassword := GeneratePassword(8, 1, 1)

		fmt.Printf("no password given, generated password for user '%s': '%s'\n", username, generatedPassword)

		addUser(username, generatedPassword, otp)
	} else if err := CheckPasswordRequirements(password); err != nil {
		log.Fatalf("password does not meet minimum requirements: %s", err)
	} else {
		encodedPasswordHash := GenerateHash(password)

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
			Password:  encodedPasswordHash,
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

	c.HTML(http.StatusOK, "login.html", gin.H{
		"recaptchaEnabled": GetRecaptchaEnabled(),
		"recaptchaSiteKey": GetRecaptchaSiteKey(),
	})
}

func logout(c *gin.Context) {
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil && VerifyCookie(cookieValue) == nil {
		cookie := GetCookieByValue(cookieValue)
		err := DeleteCookieByValue(cookieValue)

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

type LoginFormData struct {
	Username       string `json:"inputUsername"`
	Password       string `json:"inputPassword"`
	TOTP           string `json:"inputTotp"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type RecaptchaResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	Hostname           string    `json:"hostname"`
}

func processLoginForm(c *gin.Context) {
	requestCookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil && VerifyCookie(requestCookieValue) == nil {
		// user already authorized
		c.Status(200)
		return
	}

	var data LoginFormData
	_ = c.Bind(&data)

	if GetRecaptchaEnabled() {
		if data.RecaptchaToken == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad input for reCAPTCHA token"})
			return
		}

		postBody := bytes.NewBuffer([]byte(fmt.Sprintf("secret=%s&response=%s&remoteip=%s", GetRecaptchaSecretKey(), data.RecaptchaToken, c.ClientIP())))

		httpResponse, err := http.Post("https://www.google.com/recaptcha/api/siteverify", "application/x-www-form-urlencoded", postBody)

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "could not verify reCAPTCHA token with Google servers"})
			return
		}

		defer httpResponse.Body.Close()

		responseBody, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "could not read reCAPTCHA verification response from Google"})
			return
		}

		var response RecaptchaResponse
		err = json.Unmarshal(responseBody, &response)

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "could not unserialize reCAPTCHA verification response from Google"})
			return
		}

		if !response.Success {
			c.AbortWithStatusJSON(500, gin.H{"error": "reCAPTCHA verification unsuccessful"})
			return
		}
	}

	user := GetUserByUsername(data.Username)

	if user == nil {
		if ldapAuthenticate(data.Username, data.Password) {
			createAndSetAuthCookie(c, data.Username)
			c.Status(200)
		} else {
			c.AbortWithStatus(401)
			return
		}
	} else {
		if CompareHashAndPassword(user.Password, data.Password) != nil {
			c.AbortWithStatus(401)
			return
		} else {
			if len(user.OtpSecret) != 0 {
				token := c.PostForm("inputTotp")

				secret := Decrypt(user.OtpSecret, data.Password)

				tokenIsValid := totp.Validate(token, string(secret))

				if !tokenIsValid {
					c.AbortWithStatusJSON(401, gin.H{"error": "invalid TOTP"})
					return
				}
			}

			cookie := createAndSetAuthCookie(c, user.Username)
			c.JSON(200, gin.H{"expires": cookie.Expires.UnixMilli()})
		}
	}
}

func createAndSetAuthCookie(c *gin.Context, username string) Cookie {
	cookie := Cookie{
		Name:     "Nginx-Auth-Server-Token",
		Value:    GeneratePassword(192, 45, 90),
		Expires:  time.Now().AddDate(0, 0, GetCookieLifetime()),
		Domain:   GetDomain(),
		Username: username,
		HttpOnly: true,
		Secure:   GetCookieSecure(),
	}

	err := SaveCookie(cookie)

	if err != nil {
		log.Fatalf("error while trying to save the cookie to the database: %s", err)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Expires:  cookie.Expires,
		Domain:   cookie.Domain,
		HttpOnly: cookie.HttpOnly,
		Secure:   cookie.Secure,
	})

	return cookie
}

func whoami(c *gin.Context) {
	cookieValue, err := c.Cookie("Nginx-Auth-Server-Token")

	if err != nil || VerifyCookie(cookieValue) != nil {
		c.AbortWithStatus(401)
		return
	} else {
		cookie := GetCookieByValue(cookieValue)

		c.JSON(200, gin.H{"username": cookie.Username})
		return
	}

}
