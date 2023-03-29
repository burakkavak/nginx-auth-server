package main

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// This file handles any logic regarding the configuration (config.ini).
// Refer to the go-ini/ini docs (https://ini.unknwon.io/docs).

// Server :: [Server]-Section of .ini
type Server struct {
	ListenAddress string `ini:"listen_address"`
	ListenPort    int    `ini:"listen_port"`
	Domain        string `ini:"domain"`
}

// TLS :: [TLS]-Section of .ini
type TLS struct {
	Enabled    bool   `ini:"enabled"`
	ListenPort int    `ini:"listen_port"`
	CertPath   string `ini:"cert_path"`
	KeyPath    string `ini:"key_path"`
}

// Cookies :: [Cookies]-Section of .ini
type Cookies struct {
	Lifetime int    `ini:"lifetime"`
	Secure   bool   `ini:"secure"`
	Name     string `ini:"name"`
}

// LDAP :: [LDAP]-Section of .ini
type LDAP struct {
	Enabled            bool   `ini:"enabled"`
	URL                string `ini:"url"`
	OrganizationalUnit string `ini:"organizational_unit"`
	DomainComponents   string `ini:"domain_components"`
}

// Recaptcha :: [Recaptcha]-Section of .ini
type Recaptcha struct {
	Enabled   bool   `ini:"enabled"`
	SiteKey   string `ini:"site_key"`
	SecretKey string `ini:"secret_key"`
}

// GeoIP :: [GeoIP]-Section of .ini
type GeoIP struct {
	Path string `ini:"path"`
}

type Config struct {
	Server
	TLS
	Cookies
	LDAP
	Recaptcha
	GeoIP
}

var (
	// parsed defines if the config.ini was already parsed
	parsed = false

	// config defines the default configuration state and
	// will be overridden with any changed values in the config.ini
	config = &Config{
		Server: Server{
			ListenAddress: "127.0.0.1",
			ListenPort:    17397,
			Domain:        "localhost",
		},
		TLS: TLS{
			Enabled:    false,
			ListenPort: 17760,
			CertPath:   "",
			KeyPath:    "",
		},
		Cookies: Cookies{
			Lifetime: 7,
			Secure:   true,
			Name:     "Nginx-Auth-Server-Token",
		},
		LDAP: LDAP{
			Enabled:            false,
			URL:                "",
			OrganizationalUnit: "users",
			DomainComponents:   "",
		},
		Recaptcha: Recaptcha{
			Enabled:   false,
			SiteKey:   "",
			SecretKey: "",
		},
		GeoIP: GeoIP{
			Path: "",
		},
	}
)

const (
	configFileName = "config.ini"
)

func parse() {
	// don't parse the configuration if it already was parsed
	if parsed {
		return
	}

	file, err := ini.Load(fmt.Sprintf("%s/%s", GetExecutableDirectory(), configFileName))

	if err != nil {
		appLog.Fatalf("fatal error while reading configuration from 'config.ini': %s", err)
	}

	// map the configuration in config.ini to the data structure
	err = file.MapTo(config)

	if err != nil {
		appLog.Fatalf("fatal error while pasing configuration to types: %s", err)
	}

	parsed = true
}

// Getter for all configuration keys.
// TODO: maybe there is a better way to do this

func GetListenAddress() string {
	parse()
	return config.Server.ListenAddress
}

func GetListenPort() int {
	parse()
	return config.Server.ListenPort
}

func GetDomain() string {
	parse()
	return config.Server.Domain
}

func GetTlsEnabled() bool {
	parse()
	return config.TLS.Enabled
}

func GetTlsListenPort() int {
	parse()
	return config.TLS.ListenPort
}

func GetTlsCertPath() string {
	parse()
	return config.TLS.CertPath
}

func GetTlsKeyPath() string {
	parse()
	return config.TLS.KeyPath
}

func GetCookieLifetime() int {
	parse()
	return config.Cookies.Lifetime
}

func GetCookieSecure() bool {
	parse()
	return config.Cookies.Secure
}

func GetCookieName() string {
	parse()
	return config.Cookies.Name
}

func GetLDAPEnabled() bool {
	parse()
	return config.LDAP.Enabled
}

func GetLDAPUrl() string {
	parse()
	return config.LDAP.URL
}

func GetLDAPOrganizationalUnit() string {
	parse()
	return config.LDAP.OrganizationalUnit
}

func GetLDAPDomainComponents() string {
	parse()
	return config.LDAP.DomainComponents
}

func GetRecaptchaEnabled() bool {
	parse()
	return config.Recaptcha.Enabled
}

func GetRecaptchaSiteKey() string {
	parse()
	return config.Recaptcha.SiteKey
}

func GetRecaptchaSecretKey() string {
	parse()
	return config.Recaptcha.SecretKey
}

func GetGeoIPDatabasePath() string {
	parse()
	return config.GeoIP.Path
}
