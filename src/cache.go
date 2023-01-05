package main

var (
	cache = make(map[string]*Cookie)
)

func SaveCookieToCache(cookie *Cookie, plainCookieValue string) {
	cache[plainCookieValue] = cookie
}

func GetCookieFromCache(plainCookieValue string) *Cookie {
	return cache[plainCookieValue]
}
