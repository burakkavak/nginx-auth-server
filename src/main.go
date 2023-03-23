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

// templateFiles contains any files that are served prefixed with the relative URL /nginx-auth-server-static.
// These files are embedded in the final executable using go:embed.
//
//go:embed templates
var templateFiles embed.FS

// staticFiles contains any files that are served prefixed with the relative URL /nginx-auth-server-static.
// These files are embedded in the final executable using go:embed.
//
//go:embed css js
var staticFiles embed.FS

// GinMode describes the Gin web framework operating mode. This variable is overwritten in prod.go
// if the executable is build using the Go build tag 'prod'.
var GinMode = gin.DebugMode

const AppVersion = "0.0.8"

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := app.Run(os.Args); err != nil {
		fmt.Printf(err.Error())
	}
}

// runGin sets up the Gin router and starts the webserver.
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

	// overwrite serverAddress if TLS is configured and enabled
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

	// start the webserver in HTTP or HTTPS mode
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

// addUser receives the username and plaintext password and adds the new user to the database.
// If the password is empty, addUser will generate a password.
func addUser(username string, password string, otp bool) {
	if username == "" {
		appLog.Fatalf("invalid username")
	}

	if GetUserByUsernameCaseInsensitive(username) != nil {
		appLog.Fatalf("user with username %s already exists", username)
	}

	// generate password if empty password was given
	if password == "" {
		generatedPassword := GeneratePassword(8, 1, 1)

		fmt.Printf("no password given, generated password for user '%s': '%s'\n", username, generatedPassword)

		addUser(username, generatedPassword, otp)
	} else if err := CheckPasswordRequirements(password); err != nil {
		fmt.Printf("password does not meet minimum requirements: %s\n", err)
		return
	} else {
		// hash password using argon2 for database storage
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

			// output TOTP url as QR code to stdout using libqrencode
			output, err := exec.Command("sh", "-c", fmt.Sprintf("qrencode -t UTF8 '%s'", otpKey.URL())).Output()

			fmt.Printf("TOTP URL for user '%s': '%s'\n", username, otpKey.URL())

			if err != nil {
				fmt.Println("install 'qrencode' library to display a QR code.")
			} else {
				fmt.Println(string(output))
			}

			// encrypt TOTP secret using user password for database storage
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

// removeUser removes a user from the database.
// TODO: don't delete associated user cookies if there is an existing LDAP user with the same username
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

// authenticate handles the /auth route. If a valid cookie is found in the request header, the
// the response will be 200. If the cookie is invalid or expired, 401 is set as a response status.
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

// login handles the /login route. If a valid cookie is found in the request header, the
// the response will be 302 redirect to the given 'callback' query param. If no callback is given, the user
// will be redirected to the root page. If the user is not authenticated,
// the login form template will be displayed.
func login(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")

	if err == nil {
		if _, err = VerifyCookie(token); err == nil {
			// user already authorized
			// refer to: https://github.com/burakkavak/nginx-auth-server/issues/2
			c.Redirect(302, c.Query("callback"))
			return
		}
	}

	// attach all embedded CSS/JS files to the HTML template
	cssFiles := GetFilenamesFromFS(staticFiles, "css")
	jsFiles := GetFilenamesFromFS(staticFiles, "js")

	c.HTML(http.StatusOK, "login.html", gin.H{
		"cssFiles":         cssFiles,
		"jsFiles":          jsFiles,
		"recaptchaEnabled": GetRecaptchaEnabled(),
		"recaptchaSiteKey": GetRecaptchaSiteKey(),
	})
}

// logout handles the /logout route. If a valid cookie is found in the request header, the
// the response will be 200 and the cookie will be deleted.
func logout(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")
	clientIp := GetClientIpFromContext(c)

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
				"but the associated authentication cookie could not be deleted from the database. %s\n", cookie.Username, clientIp, err)
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
			authLog.Printf("user with username '%s' and client IP '%s' successfully logged out\n", cookie.Username, clientIp)
			return
		}
	} else {
		c.AbortWithError(http.StatusUnauthorized, errors.New("user not logged in or invalid/expired cookie provided"))
		return
	}
}

// LoginFormData represents the inputs defined in the login template as a struct.
type LoginFormData struct {
	Username       string `json:"inputUsername"`
	Password       string `json:"inputPassword"`
	TOTP           string `json:"inputTotp"`
	RecaptchaToken string `json:"recaptchaToken"`
}

// RecaptchaResponse defines the structure of the Google reCAPTCHA verification response as a struct.
type RecaptchaResponse struct {
	Success            bool      `json:"success"`
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	Hostname           string    `json:"hostname"`
}

// processLoginForm handles the POST /login route. If the request already contains a valid cookie, 200 is returned.
// If the user has provided valid credentials in the login form, the response will contain a new cookie (200).
// If the username, the password, the TOTP token or the reCAPTCHA token is invalid, the request is rejected.
func processLoginForm(c *gin.Context) {
	token, err := c.Cookie("Nginx-Auth-Server-Token")
	clientIp := GetClientIpFromContext(c)

	if err == nil {
		if _, err = VerifyCookie(token); err == nil {
			// user already authorized
			c.Status(200)
			return
		}
	}

	var data LoginFormData
	_ = c.Bind(&data)

	// verify reCAPTCHA token if reCAPTCHA is enabled
	if GetRecaptchaEnabled() {
		if data.RecaptchaToken == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad input for reCAPTCHA token"})
			return
		}

		postBody := bytes.NewBuffer([]byte(fmt.Sprintf("secret=%s&response=%s&remoteip=%s", GetRecaptchaSecretKey(), data.RecaptchaToken, clientIp)))

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
		// if a user with the given username does not exist, check if LDAP authenticates
		if ldapAuthenticate(data.Username, data.Password) {
			createAndSetAuthCookie(c, data.Username)
			c.Status(200)
			authLog.Printf("LDAP user with username '%s' and client IP '%s' logged in successfully\n", data.Username, clientIp)
		} else {
			c.AbortWithStatus(401)
			return
		}
	} else {
		// if a user with the given username was found in the database, check password validity
		if CompareHashAndPassword(user.Password, data.Password) != nil {
			c.AbortWithStatus(401)
			authLog.Printf("invalid password for user with username '%s' and client IP '%s'\n", data.Username, clientIp)
			return
		} else {
			// if TOTP is enabled for the user, check the validity of the TOTP token input from the user
			if len(user.OtpSecret) != 0 {
				secret := Decrypt(user.OtpSecret, data.Password)

				tokenIsValid := totp.Validate(data.TOTP, string(secret))

				if !tokenIsValid {
					c.AbortWithStatusJSON(401, gin.H{"error": "invalid TOTP"})
					return
				}
			}

			cookie := createAndSetAuthCookie(c, user.Username)
			c.JSON(200, gin.H{"expires": cookie.Expires.UnixMilli()})
			authLog.Printf("user with username '%s' and client IP '%s' logged in successfully\n", data.Username, clientIp)
		}
	}
}

// createAndSetAuthCookie sets a new cookie for the given gin.Context and username and saves it to the database.
// This function is called after the user credentials have been verified.
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

// whoami handles the /whoami route. If the request contains a valid cookie,
// 200 and the username (formatted as JSON) is returned.
// If the cookie in the request header is invalid, 401 Unauthorized is returned.
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
