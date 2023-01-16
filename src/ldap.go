package main

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
)

// This file handles any logic related to the LDAP interface.
// Refer to the go-ldap/ldap documentation (https://pkg.go.dev/github.com/go-ldap/ldap).

// ldapAuthenticate returns true if the user was successfully authenticated with the LDAP server.
func ldapAuthenticate(username string, password string) bool {
	if !GetLDAPEnabled() {
		return false
	}

	l := ldapConnect()
	defer l.Close()

	err := l.Bind(fmt.Sprintf("CN=%s,ou=%s,%s", username, GetLDAPOrganizationalUnit(), GetLDAPDomainComponents()), password)

	if err != nil {
		appLog.Printf("error validating credentials: %s\n", err)
		return false
	}

	return true
}

// ldapCheckUserExists checks for existing users with the given username.
// Returns true if a user exists in LDAP with the given username.
func ldapCheckUserExists(username string) bool {
	if !GetLDAPEnabled() {
		return false
	}

	l := ldapConnect()
	defer l.Close()

	result, err := l.Search(&ldap.SearchRequest{
		BaseDN:       fmt.Sprintf("ou=%s,%s", GetLDAPOrganizationalUnit(), GetLDAPDomainComponents()),
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       fmt.Sprintf("(uid=%s)", username),
		Attributes:   []string{"dn"},
		Controls:     nil,
	})

	if err != nil {
		appLog.Printf("error searching LDAP user with username '%s': %s\n", username, err)
		return false
	}

	if len(result.Entries) > 0 {
		return true
	} else {
		return false
	}
}

// ldapConnect connects to the LDAP server and returns the ldap.Conn.
// Returns nil if the connection failed.
func ldapConnect() *ldap.Conn {
	if !GetLDAPEnabled() {
		return nil
	}

	var (
		l   *ldap.Conn
		err error
	)

	l, err = ldap.DialURL(GetLDAPUrl())

	if err != nil {
		appLog.Printf("error trying to connect to LDAP server: %s\n", err)
		return nil
	} else {
		return l
	}
}
