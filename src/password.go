package main

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"runtime"
	"strings"
)

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// GeneratePassword generates a password with the given length, a minimum number of numbers, and a minimum
// number of uppercase characters.
// Refer to https://golangbyexample.com/generate-random-password-golang/
func GeneratePassword(passwordLength, minNum, minUpperCase int) string {
	const (
		lowerCharSet = "abcdedfghijklmnopqrst"
		upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberSet    = "0123456789"
		allCharSet   = lowerCharSet + upperCharSet + numberSet
	)

	var password strings.Builder

	// Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	// Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}

// GenerateHash derives an argon2id-hash from the given password.
func GenerateHash(password string) string {
	// use double the number of (virtual) cores as the argon2-parallelism parameter
	parallelism := runtime.NumCPU() * 2

	if parallelism > 255 {
		parallelism = 255
	}

	// argon2 generation parameters
	params := &argonParams{
		memory:      64 * 1024,
		iterations:  2,
		parallelism: uint8(parallelism),
		saltLength:  16,
		keyLength:   32,
	}

	salt, err := GenerateRandomBytes(params.saltLength)

	if err != nil {
		appLog.Fatal("an error occurred while trying to generate a random salt.")
	}

	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.memory, params.iterations, params.parallelism, b64Salt, b64Hash)
}

// CompareHashAndPassword compares the given argon2id-hash to the given password.
// If the password matches the hash, nil is returned.
func CompareHashAndPassword(encodedHash string, password string) error {
	params, salt, hash, err := decodeArgonHash(encodedHash)

	if err != nil {
		// error trying to decode argon hash
		// the hash may be a (legacy) bcrypt hash
		return bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	}

	otherHash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}

	return errors.New("password does not match against argon2 hash")
}

// decodeArgonHash deconstruct the given argon2id-hash into it's components.
// Example for a valid parameter: $argon2id$v=19$m=65536,t=2,p=32$dCk......8cM
func decodeArgonHash(encodedHash string) (p *argonParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")

	if len(vals) != 6 {
		return nil, nil, nil, errors.New("the encoded hash is not in the correct format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)

	if err != nil {
		return nil, nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version of argon2")
	}

	p = &argonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)

	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])

	if err != nil {
		return nil, nil, nil, err
	}

	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])

	if err != nil {
		return nil, nil, nil, err
	}

	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

// CheckPasswordRequirements verifies if a given plaintext password satisfies the minimum requirements.
func CheckPasswordRequirements(password string) error {
	if len(password) < 6 {
		return errors.New("error: password length must exceed 6 or more")
	}

	return nil
}
