package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/big"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unsafe"
	"uuid"

	"github.com/goccy/go-json"
	"github.com/godruoyi/go-snowflake"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/roysitumorang/laukpauk/keys"
	"github.com/roysitumorang/laukpauk/pools"
	"github.com/sqids/sqids-go"
)

const (
	numbers             = "0123456789"
	base58alphabets     = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	lowerCasedAlphabets = "123456789abcdefghijkmnopqrstuvwxyz"
)

var (
	env,
	jwtIssuer string
	kafkaBrokers           []string
	loginMaxFailedAttempts int
	loginLockoutDuration,
	accessTokenAge time.Duration
	shortUniqueIDs *sqids.Sqids
	privateKey     *rsa.PrivateKey
	InitHelper     = sync.OnceValue(func() (err error) {
		location, ok := os.LookupEnv("TZ")
		if !ok || location == "" {
			return errors.New("env TZ is required")
		}
		timezone, err := time.LoadLocation(location)
		if err != nil {
			return
		}
		time.Local = timezone
		if env, ok = os.LookupEnv("ENV"); !ok {
			return errors.New("env ENV is required")
		}
		if env == "" {
			env = "development"
		}
		if jwtIssuer, ok = os.LookupEnv("JWT_ISSUER"); !ok || jwtIssuer == "" {
			return errors.New("env JWT_ISSUER is required")
		}
		envKafkaBrokers, ok := os.LookupEnv("KAFKA_BROKERS")
		if !ok || envKafkaBrokers == "" {
			return errors.New("env KAFKA_BROKERS is required")
		}
		kafkaBrokers = strings.Split(envKafkaBrokers, ";")
		envLoginMaxFailedAttempts, ok := os.LookupEnv("LOGIN_MAX_FAILED_ATTEMPTS")
		if !ok || envLoginMaxFailedAttempts == "" {
			return errors.New("env LOGIN_MAX_FAILED_ATTEMPTS is required")
		}
		if loginMaxFailedAttempts, err = strconv.Atoi(envLoginMaxFailedAttempts); err != nil || loginMaxFailedAttempts < 1 {
			return errors.New("env LOGIN_MAX_FAILED_ATTEMPS requires a positive integer")
		}
		envLoginLockoutDuration, ok := os.LookupEnv("LOGIN_LOCKOUT_DURATION")
		if !ok || envLoginLockoutDuration == "" {
			return errors.New("env LOGIN_LOCKOUT_DURATION is required")
		}
		if loginLockoutDuration, err = time.ParseDuration(envLoginLockoutDuration); err != nil {
			return
		}
		envSqidsMinLength, ok := os.LookupEnv("SQIDS_MIN_LENGTH")
		if !ok || envSqidsMinLength == "" {
			return errors.New("env SQIDS_MIN_LENGTH is required")
		}
		sqidsMinLength, err := strconv.ParseUint(envSqidsMinLength, 10, 8)
		if err != nil || sqidsMinLength < 1 {
			return fmt.Errorf("env SQIDS_MIN_LENGTH requires a positive integer, min. 1, max %d", math.MaxUint8)
		}
		if shortUniqueIDs, err = sqids.New(sqids.Options{
			Alphabet:  lowerCasedAlphabets,
			MinLength: uint8(sqidsMinLength),
		}); err != nil {
			return
		}
		envAccesTokenAge, ok := os.LookupEnv("ACCESS_TOKEN_AGE")
		if !ok || envAccesTokenAge == "" {
			return errors.New("env ACCESS_TOKEN_AGE is required")
		}
		if accessTokenAge, err = time.ParseDuration(envAccesTokenAge); err != nil {
			return
		}
		privateKey, err = keys.InitPrivateKey()
		return
	})
)

func String2ByteSlice(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

func ByteSlice2String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func GenerateSnowflakeID() uint64 {
	return snowflake.ID()
}

func EncodeShortUniqueIDs(numbers ...uint64) (string, error) {
	n := len(numbers)
	if n == 0 {
		return "", nil
	}
	return shortUniqueIDs.Encode(numbers)
}

func DecodeShortUniqueIDs(id string) uint64 {
	if id == "" {
		return 0
	}
	unsignedNumbers := shortUniqueIDs.Decode(id)
	if len(unsignedNumbers) == 0 {
		return 0
	}
	return unsignedNumbers[0]
}

func GenerateUniqueID() string {
	return uuid.New().String()
}

func GetEnv() string {
	return env
}

func Transcode(input, output any) error {
	buffer, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(buffer, output)
}

func GetJwtIssuer() string {
	return jwtIssuer
}

func GenerateAccessToken(id, subject, audience string, createdAt, expiredAt time.Time) (string, error) {
	numericDate := jwt.NewNumericDate(createdAt)
	var claims jwt.RegisteredClaims
	claims.ID = id
	claims.Subject = subject
	claims.Audience = append(claims.Audience, audience)
	claims.Issuer = jwtIssuer
	claims.IssuedAt = numericDate
	claims.NotBefore = numericDate
	claims.ExpiresAt = jwt.NewNumericDate(expiredAt)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func GetKafkaBrokers() []string {
	return kafkaBrokers
}

// RandomString generate random string
func RandomString(length int) string {
	randomBytes := make([]byte, length)
	for {
		if _, err := rand.Read(randomBytes); err == nil {
			break
		}
	}
	for i := range length {
		randomBytes[i] = base58alphabets[randomBytes[i]%58]
	}
	return ByteSlice2String(randomBytes)
}

// RandomNumber generate random number
func RandomNumber(length int) string {
	randomBytes := make([]byte, length)
	for {
		if _, err := rand.Read(randomBytes); err == nil {
			break
		}
	}
	for i := range length {
		randomBytes[i] = numbers[randomBytes[i]%10]
	}
	return ByteSlice2String(randomBytes)
}

func ValidPassword(password string) bool {
	var hasUpperCase,
		hasLowerCase,
		hasNumber,
		hasSymbol bool
	length := len(password)
	for _, char := range password {
		hasUpperCase = hasUpperCase || unicode.IsUpper(char)
		hasLowerCase = hasLowerCase || unicode.IsLower(char)
		hasNumber = hasNumber || unicode.IsNumber(char)
		hasSymbol = hasSymbol || unicode.IsPunct(char) || unicode.IsSymbol(char)
	}
	return hasUpperCase &&
		hasLowerCase &&
		hasNumber &&
		hasSymbol &&
		length >= 8
}

func Base64Decode(input string) (string, error) {
	output, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		if _, ok := err.(base64.CorruptInputError); ok {
			err = errors.New("malformed input")
		}
		return "", err
	}
	return ByteSlice2String(output), nil
}

func GetLoginMaxFailedAttempts() int {
	return loginMaxFailedAttempts
}

func GetLoginLockoutDuration() time.Duration {
	return loginLockoutDuration
}

func GetAccessTokenAge() time.Duration {
	return accessTokenAge
}

func Sprintf(format string, a ...any) string {
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	_, _ = fmt.Fprintf(sb, format, a...)
	return sb.String()
}

func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 0 {
		return ""
	}
	return unsafe.String(unsafe.StringData(s), maxLength)
}

func PhotoDirectory() string {
	return os.Getenv("PHOTO_DIRECTORY")
}

func PhotoPath(filename string, baseURLs ...string) string {
	var baseURL string
	if len(baseURLs) > 0 {
		baseURL = baseURLs[0]
	}
	return Sprintf("%s/%s/%s", baseURL, PhotoDirectory(), filename)
}

func PhotoLocation(filename string) string {
	return Sprintf(".%s", PhotoPath(filename))
}

func SaveFile(c fiber.Ctx, file *multipart.FileHeader, mimetype *string, filename string) error {
	if file == nil {
		return nil
	}
	dst := PhotoLocation(filename)
	err := c.SaveFile(file, dst)
	if err != nil {
		return err
	}
	if mimetype != nil {
		switch *mimetype {
		case "image/jpeg":
			if err = Jpeg2Webp(dst, dst); err != nil {
				return err
			}
		case "image/png":
			if err = Png2Webp(dst, dst); err != nil {
				return err
			}
		}
	}
	return os.Chmod(dst, 0644)
}

func Jpeg2Webp(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = input.Close()
	}()
	image, err := jpeg.Decode(input)
	if err != nil {
		return err
	}
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = output.Close()
	}()
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 80)
	if err != nil {
		return err
	}
	return webp.Encode(output, image, options)
}

func Png2Webp(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = input.Close()
	}()
	image, err := png.Decode(input)
	if err != nil {
		return err
	}
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = output.Close()
	}()
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 80)
	if err != nil {
		return err
	}
	if err = webp.Encode(output, image, options); err != nil {
		return err
	}
	return webp.Encode(output, image, options)
}

func RandomInteger(bound int64) int64 {
	max := big.NewInt(bound)
	var (
		n   *big.Int
		err error
	)
	for {
		if n, err = rand.Int(rand.Reader, max); err == nil {
			break
		}
	}
	return n.Int64()
}

func ValidLatitude(latitude float64) bool {
	return latitude >= -90 && latitude <= 90
}

func ValidLongitude(longitude float64) bool {
	return longitude >= -180 && longitude <= 180
}

func Encrypt(plaintext string) (string, error) {
	key := String2ByteSlice(os.Getenv("ENCRYPTION_SECRET_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, String2ByteSlice(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(cryptoText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	key := String2ByteSlice(os.Getenv("ENCRYPTION_SECRET_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", err
	}
	return ByteSlice2String(plaintext), nil
}
