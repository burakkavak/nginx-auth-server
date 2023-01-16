package main

// This file handles the caching of authentications. Once a user successfully authenticated, the plaintext cookie
// value and the corresponding cookie is saved to the cache. This cache persists for the runtime of the application.

var (
	cache = make(map[string]*Cookie)
)

// SaveCookieToCache saves a cookie and the corresponding plaintext cookie value to the cache.
// This dramatically decreases latency for future requests, since the plain cookie value does not need to
// be matched to the argon2 hash in the database for every request.
func SaveCookieToCache(cookie *Cookie, plainCookieValue string) {
	cache[plainCookieValue] = cookie
}

// GetCookieFromCache returns the cookie corresponding to the given plaintext cookie value.
// Returns nil if no cookie was found.
func GetCookieFromCache(plainCookieValue string) *Cookie {
	return cache[plainCookieValue]
}

// DeleteCookieFromCache deletes a existing cookie from the cache.
func DeleteCookieFromCache(cookie *Cookie) {
	if cookie != nil {
		delete(cache, cookie.Value)
	}
}
