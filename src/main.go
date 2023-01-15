package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
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
		fmt.Printf(err.Error())
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

	serverAddress := GetListenAddress() + ":" + strconv.Itoa(GetListenPort())
	tlsEnabled := GetTlsEnabled()
	tlsCertPath := GetTlsCertPath()
	tlsKeyPath := GetTlsKeyPath()

	if tlsEnabled {
		serverAddress = GetListenAddress() + ":" + strconv.Itoa(GetTlsListenPort())

		if err := CheckFileReadable(tlsCertPath); err != nil {
			appLog.Fatalf("fatal error: TLS certificate at '%s' does not exist or is not readable. Check the configuration and/or file permissions.", GetTlsCertPath())
		}

		if err := CheckFileReadable(tlsKeyPath); err != nil {
			appLog.Fatalf("fatal error: TLS key at '%s' does not exist or is not readable. Check the configuration and/or file permissions.", GetTlsKeyPath())
		}
	}

	server := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		var err error = nil

		if tlsEnabled {
			appLog.Printf("listening and serving HTTPS request on %s\n", serverAddress)
			err = server.ListenAndServeTLS(tlsCertPath, tlsKeyPath)
		} else {
			appLog.Printf("listening and serving HTTP request on %s\n", serverAddress)
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			appLog.Fatalf("fatal error trying to launch the webserver: %s\n", err)
		}
	}()

	// gracefully quit Gin server (https://gin-gonic.com/docs/examples/graceful-restart-or-stop/)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLog.Println("Shutting down webserver...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLog.Fatalf("fatal error: could not shutdown server gracefully. %s\n", err)
	}
}

func addUser(username string, password string, otp bool) {
	if username == "" {
		appLog.Fatalf("invalid username")
	}

	if GetUserByUsernameCaseInsensitive(username) != nil {
		appLog.Fatalf("user with username %s already exists", username)
	}

	if password == "" {
		generatedPassword := GeneratePassword(8, 1, 1)

		fmt.Printf("no password given, generated password for user '%s': '%s'\n", username, generatedPassword)

		addUser(username, generatedPassword, otp)
	} else if err := CheckPasswordRequirements(password); err != nil {
		fmt.Printf("password does not meet minimum requirements: %s\n", err)
		return
	} else {
		encodedPasswordHash := GenerateHash(password)

		var encryptedOtpSecret []byte

		if otp {
			otpKey, err := totp.Generate(totp.GenerateOpts{
				Issuer:      GetDomain(),
				AccountName: username,
			})

			if err != nil {
				appLog.Fatalf("could not create TOTP: %s", err)
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
			appLog.Fatalf("fatal error: could not save user to database: %s", err)
		} else {
			appLog.Printf("user with username '%s' successfully created\n", username)
		}
	}
}

func removeUser(username string) {
	err := RemoveUser(username)

	if err != nil {
		appLog.Fatalf("fatal error: could not remove user from database: %s", err)
	} else {
		appLog.Printf("user with username '%s' has been removed\n", username)
	}

	err = DeleteCookiesByUsername(username)

	if err != nil {
		appLog.Fatalf("fatal error: could not remove user associated cookies from database for username '%s': %s\n", username, err)
	} else {
		appLog.Printf("user associated cookies for username '%s' have been removed\n", username)
	}
}

func authenticate(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	if _, err = VerifyCookie(token); err != nil {
		c.AbortWithStatus(401)
		return
	} else {
		c.Status(200)
		return
	}

}

func login(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil {
		if _, err = VerifyCookie(token); err == nil {
			// user already authorized
			c.Status(200)
			return
		}
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"recaptchaEnabled": GetRecaptchaEnabled(),
		"recaptchaSiteKey": GetRecaptchaSiteKey(),
	})
}

func logout(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	if cookie, err := VerifyCookie(token); err == nil {
		DeleteCookieFromCache(cookie)
		err := DeleteCookie(cookie)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("could not delete user cookie from database"))
			authLog.Printf("error: user with username '%s' and client IP '%s' tried logging out,"+
				"but the associated authentication cookie could not be deleted from the database. %s\n", cookie.Username, c.ClientIP(), err)
			return
		} else {
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     cookie.Name,
				Value:    fmt.Sprintf("$username=%s,$value=%s", cookie.Username, cookie.Value),
				Expires:  time.Now(),
				Domain:   cookie.Domain,
				HttpOnly: cookie.HttpOnly,
				Secure:   cookie.Secure,
			})

			c.String(http.StatusOK, "user successfully logged out")
			authLog.Printf("user with username '%s' and client IP '%s' successfully logged out\n", cookie.Username, c.ClientIP())
			return
		}
	} else {
		c.AbortWithError(http.StatusUnauthorized, errors.New("user not logged in or invalid/expired cookie provided"))
		return
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
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil {
		if _, err = VerifyCookie(token); err == nil {
			// user already authorized
			c.Status(200)
			return
		}
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
			c.AbortWithStatusJSON(500, gin.H{"error": "could not deserialize reCAPTCHA verification response from Google"})
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
			authLog.Printf("LDAP user with username '%s' and client IP '%s' logged in successfully\n", data.Username, c.ClientIP())
		} else {
			c.AbortWithStatus(401)
			return
		}
	} else {
		if CompareHashAndPassword(user.Password, data.Password) != nil {
			c.AbortWithStatus(401)
			authLog.Printf("invalid password for user with username '%s' and client IP '%s'\n", data.Username, c.ClientIP())
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
			authLog.Printf("user with username '%s' and client IP '%s' logged in successfully\n", data.Username, c.ClientIP())
		}
	}
}

func createAndSetAuthCookie(c *gin.Context, username string) Cookie {
	plainCookieValue := GeneratePassword(96, 25, 35)

	cookie := Cookie{
		Name:     "Nginx-Auth-Server-Token",
		Value:    GenerateHash(plainCookieValue),
		Expires:  time.Now().AddDate(0, 0, GetCookieLifetime()),
		Domain:   GetDomain(),
		Username: username,
		HttpOnly: true,
		Secure:   GetCookieSecure(),
	}

	err := SaveCookie(cookie)

	if err != nil {
		appLog.Fatalf("fatal error: could not save the cookie to the database: %s", err)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookie.Name,
		Value:    fmt.Sprintf("$username=%s,$value=%s", username, plainCookieValue),
		Expires:  cookie.Expires,
		Domain:   cookie.Domain,
		HttpOnly: cookie.HttpOnly,
		Secure:   cookie.Secure,
	})

	SaveCookieToCache(&cookie, plainCookieValue)

	return cookie
}

func whoami(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	cookie, err := VerifyCookie(token)

	if err != nil {
		c.AbortWithStatus(401)
		return
	} else {
		c.JSON(200, gin.H{"username": cookie.Username})
		return
	}
}
