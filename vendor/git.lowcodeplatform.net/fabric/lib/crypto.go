package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/models"
)

// Пример использования
//func main() {
//	key := []byte("LKHlhb899Y09olUi")
//	encryptMsg, _ := encrypt(key, "Hello World")
//	msg, _ := decrypt(key, encryptMsg)
//	fmt.Println(msg) // Hello World
//}

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

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Encrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := Pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

func Decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
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

// GenXServiceKey создаем токен
func GenXServiceKey(domain string, projectKey []byte, tokenInterval time.Duration) (token string, err error) {
	t := models.XServiceKey{
		Domain:  domain,
		Expired: time.Now().Add(tokenInterval).Unix(),
	}
	strJson, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("error Marshal XServiceKey, err: %s", err)
	}

	token, err = Encrypt(projectKey, string(strJson))
	if err != nil {
		return "", fmt.Errorf("error Encrypt XServiceKey, err: %s", err)
	}

	return token, nil
}

// CheckXServiceKey берем из заголовка X-Service-Key. если он есть, то он должен быть расшифровать
// и валидируем содержимое
func CheckXServiceKey(domain string, projectKey []byte, xServiceKey string) bool {
	var xsKeyValid bool
	var xsKey models.XServiceKey

	if xServiceKey == "" {
		return false
	}

	v, err := Decrypt(projectKey, xServiceKey)
	err = json.Unmarshal([]byte(v), &xsKey)
	if err != nil {
		return false
	}

	if xsKey.Domain == domain && xsKey.Expired > time.Now().Unix() {
		xsKeyValid = true
	}
	if !xsKeyValid {
		if xsKey.Domain == string(projectKey) && xsKey.Expired > time.Now().Unix() {
			xsKeyValid = true
		}
	}

	return xsKeyValid
}
