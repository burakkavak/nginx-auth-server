package main

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

// This file handles any logic regarding the GeoIP2 database
// and location queries for specific IP addresses.
// Refer to oschwald/geoip2-golang (https://github.com/oschwald/geoip2-golang).

type Location struct {
	City    string
	Country string
}

// GetLocationFromIP returns a Location-struct with the city and country
// for a given IP address. The function returns nil if no location was found,
// the IP was invalid or the database was not configured.
func GetLocationFromIP(ip string) *Location {
	geoIPDatabasePath := GetGeoIPDatabasePath()

	if geoIPDatabasePath == "" || ip == "" {
		return nil
	}

	parsedIPAddress := net.ParseIP(ip)

	if parsedIPAddress == nil {
		appLog.Printf("error: provided string '%s' is not a valid IP address. Could not geolocate (invalid) IP address.", ip)
		return nil
	}

	err := CheckFileReadable(geoIPDatabasePath)

	if err != nil {
		appLog.Printf("error: GeoIP2 database is not readable or does not exist. %s", err)
		return nil
	}

	db, err := geoip2.Open(geoIPDatabasePath)

	if err != nil {
		appLog.Printf("an error occured trying to open the GeoIP2 database. %s", err)
		return nil
	}

	defer db.Close()

	record, err := db.City(parsedIPAddress)

	if err != nil {
		appLog.Printf("error querying GeoIP2 database for IP '%s'. %s", parsedIPAddress, err)
		return nil
	}

	// validate that database didn't return empty values and return location
	city := "?"
	country := "?"

	if record.City.Names["en"] != "" {
		city = record.City.Names["en"]
	}

	if record.Country.Names["en"] != "" {
		city = record.Country.Names["en"]
	}

	return &Location{
		City:    city,
		Country: country,
	}
}
