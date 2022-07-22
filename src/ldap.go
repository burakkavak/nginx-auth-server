package main

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

// TODO: ldaps not tested yet

// ldapAuthenticate returns true if the user was successfully authenticated with the LDAP server
func ldapAuthenticate(username string, password string) bool {
	if !GetLDAPEnabled() {
		return false
	}

	var (
		l   *ldap.Conn
		err error
	)

	l, err = ldap.DialURL(GetLDAPUrl())

	if err != nil {
		log.Printf("error trying to connect to LDAP server: %s\n", err)
		return false
	}

	defer l.Close()

	err = l.Bind(fmt.Sprintf("CN=%s,ou=%s,%s", username, GetLDAPOrganizationalUnit(), GetLDAPDomainComponents()), password)

	if err != nil {
		log.Printf("error validating credentials: %s\n", err)
		return false
	}

	return true
}
