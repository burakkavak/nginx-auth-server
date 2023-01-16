package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Encrypt encrypts given data with a given passphrase and returns the encrypted data.
// This function is used to encrypt the TOTP secret with the user password before saving it to the database.
func Encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)

	if err != nil {
		appLog.Fatal(err.Error())
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		appLog.Fatal(err.Error())
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext
}

// Decrypt decrypts given data with a given passphrase and returns the decrypted data.
// This function is used to decrypt the TOTP secret with the user password.
func Decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)

	if err != nil {
		appLog.Fatal(err.Error())
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		appLog.Fatal(err.Error())
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		appLog.Fatal(err.Error())
	}

	return plaintext
}

// GenerateRandomBytes generated a random number of bytes and returns them.
// Returns an error if it fails
func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
