// Copyright 2022 Enmotech Inc. All rights reserved.

package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/utils"
)

// Encrypt Generates the ciphertext for the given string.
// If the encryption fails, the original characters will be returned.
// If the passed string is empty, return empty directly.
func Encrypt(text string) string {
	if text == "" {
		return ""
	}

	curTime := time.Now().Format("01021504")
	salt := md5encode(curTime)
	key := salt[:8] + curTime

	cipherText, err := aesEncrypt(text, key)
	if err != nil {
		lg.Debug("failed to encrypt text '%s': %s", utils.MaskString(text), err)
		return text
	}
	return base64.StdEncoding.EncodeToString([]byte(salt[:8] + shiftEncode(curTime) + cipherText))
}

func Decrypt(rawCipher string) string {
	if rawCipher == "" {
		return ""
	}

	dec, err := base64.StdEncoding.DecodeString(rawCipher)
	if err != nil {
		lg.Debug("failed to base64 decode cipher text '%s': %s", rawCipher, err)
		return rawCipher
	}

	key := string(dec[:8]) + shiftDecode(string(dec[8:16]))
	text := string(dec[16:])
	res, err := aesDecrypt(text, key)
	if err != nil {
		lg.Debug("failed to decypt cipher '%s': %s", text, err)
		return rawCipher
	}
	return res
}

func md5encode(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func shiftEncode(s string) string {
	rs := make([]string, 0, len(s))
	for _, c := range s[:] {
		// start with '<'
		rs = append(rs, fmt.Sprintf("%c", c+12))
	}
	return strings.Join(rs, "")
}

func shiftDecode(s string) string {
	rs := make([]string, 0, len(s))
	for _, c := range s[:] {
		rs = append(rs, fmt.Sprintf("%c", c-12))
	}
	return strings.Join(rs, "")
}

func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func aesEncrypt(text string, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	msg := pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

func aesDecrypt(text string, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
