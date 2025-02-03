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

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
	ErrNoServiceKey        = errors.New("empty service key")
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
// domain - домен, для которого будет выдан токен (валидируется)
// client - наименование клиента, которому был выдан токен (не валидируется)
func GenXServiceKey(domain string, projectKey []byte, tokenInterval time.Duration, client string) (token string, err error) {
	t := models.XServiceKey{
		Domain:  domain,
		Expired: time.Now().Add(tokenInterval).Unix(),
		Client:  client,
	}

	return encodeServiceKey(t, projectKey)
}

// CheckXServiceKey берем из заголовка X-Service-Key. если он есть, то он должен быть расшифровать
// и валидируем содержимое
// client - возвращает имя клиента, которому был выдан токен (опционально)
func CheckXServiceKey(domain string, projectKey []byte, xServiceKey string) (valid bool, client string) {
	var xsKeyValid bool
	xsKey, err := decodeServiceKey(projectKey, xServiceKey)
	if err != nil {
		return false, ""
	}

	if xsKey.Domain == domain && xsKey.Expired > time.Now().Unix() {
		xsKeyValid = true
	}
	if !xsKeyValid {
		if xsKey.Domain == string(projectKey) && xsKey.Expired > time.Now().Unix() {
			xsKeyValid = true
		}
	}

	return xsKeyValid, xsKey.Client
}

// SetCheckCert устанавливает поле CheckCert в токене
func SetCheckCert(projectKey []byte, xServiceKey string, checkCert bool) (token string, err error) {
	xsKey, err := decodeServiceKey(projectKey, xServiceKey)
	if err != nil {
		return token, err
	}
	xsKey.CheckCert = checkCert

	return encodeServiceKey(xsKey, projectKey)
}

// GetCheckCert получает поле CheckCert из токена
func GetCheckCert(projectKey []byte, xServiceKey string) (checkCert bool, err error) {
	xsKey, err := decodeServiceKey(projectKey, xServiceKey)
	return xsKey.CheckCert, nil
}

// SetValidURI устанавливает доступные пути для токена
// пустая строка означает, что достпуны все
// несколько доменов разделяются запятой (без пробелов)
func SetValidURI(uris string, projectKey []byte, xServiceKey string) (token string, err error) {
	xsKey, err := decodeServiceKey(projectKey, xServiceKey)
	if err != nil {
		return token, err
	}
	xsKey.WhiteURI = uris
	return encodeServiceKey(xsKey, projectKey)
}

// IsValidURI проверяет, есть ли дотуп у токена к запрашиваемому пути
func IsValidURI(checkUri string, projectKey []byte, xServiceKey string) (isValidURI bool, err error) {
	xsKey, err := decodeServiceKey(projectKey, xServiceKey)
	if err != nil {
		return false, err
	}

	// Пустая строка = пропустить на все
	if xsKey.WhiteURI == "" {
		return true, nil
	}

	return isValidUri(xsKey, checkUri), nil
}

func isValidUri(xsKey models.XServiceKey, checkUri string) bool {
	uris := strings.Split(xsKey.WhiteURI, ",")
	for _, uri := range uris {
		// Не в префиксе - нет доступа
		if !strings.HasPrefix(checkUri, uri) {
			continue
		}
		lUri := len(uri)
		// Полное вхождение. Разрешен
		if lUri == len(checkUri) {
			return true
		}
		// Если следующий символ / или ? - тоже разрешено
		switch checkUri[lUri] {
		case '/':
			return true
		case '?':
			return true
		}
	}

	return false
}

type paramsArgon2 struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
	salt        []byte
}

func EncryptArgon2(value string, params *paramsArgon2) (string, error) {
	p := &paramsArgon2{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
	if params != nil {
		if params.memory != 0 {
			p.memory = params.memory
		}
		if params.iterations != 0 {
			p.iterations = params.iterations
		}
		if params.parallelism != 0 {
			p.parallelism = params.parallelism
		}
		if params.keyLength != 0 {
			p.keyLength = params.keyLength
		}
		if params.keyLength != 0 {
			p.keyLength = params.keyLength
		}
		if len(params.salt) != 0 {
			p.salt = params.salt
		}
	}

	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	if len(p.salt) != 0 {
		salt = p.salt
	}

	hash := argon2.IDKey([]byte(value), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return base64.RawStdEncoding.EncodeToString([]byte(encodedHash)), nil
}

func CheckArgon2(rawText, cryptoText string) bool {
	encodedHash, _ := base64.RawStdEncoding.Strict().DecodeString(cryptoText)
	p, salt, _, err := decodeHash(string(encodedHash))
	if err != nil {
		return false
	}

	p.salt = salt
	thisHash, _ := EncryptArgon2(rawText, p)
	if thisHash == cryptoText {
		return true
	}

	return false
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func decodeHash(encodedHash string) (p *paramsArgon2, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &paramsArgon2{}
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

// decodeServiceKey расшифровыет модель XServiceKey из токена
func decodeServiceKey(projectKey []byte, xServiceKey string) (xsKey models.XServiceKey, err error) {
	if xServiceKey == "" {
		return xsKey, ErrNoServiceKey
	}

	v, err := Decrypt(projectKey, xServiceKey)
	err = json.Unmarshal([]byte(v), &xsKey)
	return
}

// encodeServiceKey шифрует модель XServiceKey в токен
func encodeServiceKey(xsKey models.XServiceKey, projectKey []byte) (token string, err error) {
	strJson, err := json.Marshal(xsKey)
	if err != nil {
		return "", fmt.Errorf("error Marshal XServiceKey, err: %s", err)
	}

	token, err = Encrypt(projectKey, string(strJson))
	if err != nil {
		return "", fmt.Errorf("error Encrypt XServiceKey, err: %s", err)
	}
	return token, nil
}
