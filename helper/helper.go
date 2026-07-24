package helper

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"fmt"
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

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/keys"
	"github.com/sqids/sqids-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	numbers             = "0123456789"
	base58alphabets     = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	lowerCasedAlphabets = "123456789abcdefghijkmnopqrstuvwxyz"
)

var (
	timezone *time.Location
	env,
	jwtIssuer string
	kafkaBrokers           []string
	loginMaxFailedAttempts int
	loginLockoutDuration,
	accessTokenAge time.Duration
	sqIDs         *sqids.Sqids
	dbWrite       *pgxpool.Pool
	privateKey    *rsa.PrivateKey
	snowflakeNode *snowflake.Node
	bufferPool    = sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
	InitHelper = sync.OnceValue(func() (err error) {
		location, ok := os.LookupEnv("TZ")
		if !ok || location == "" {
			return errors.New("env TZ is required")
		}
		if timezone, err = time.LoadLocation(location); err != nil {
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
		sqidsMinLength, err := strconv.Atoi(envSqidsMinLength)
		if err != nil || sqidsMinLength < 1 || sqidsMinLength > math.MaxUint8 {
			return fmt.Errorf("env SQIDS_MIN_LENGTH requires a positive integer, min. 1, max %d", math.MaxUint8)
		}
		if sqIDs, err = sqids.New(sqids.Options{
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
		if privateKey, err = keys.InitPrivateKey(); err != nil {
			return
		}
		snowflakeNode, err = snowflake.NewNode(1)
		return
	})
)

func InitDbWrite(dbWriteOnly *pgxpool.Pool) {
	dbWrite = dbWriteOnly
}

func BeginTx(ctx context.Context) (pgx.Tx, error) {
	return dbWrite.Begin(ctx)
}

func String2ByteSlice(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

func ByteSlice2String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func GenerateSnowflakeID() int64 {
	return snowflakeNode.Generate().Int64()
}

func EncodeSqids(numbers ...int64) (string, error) {
	n := len(numbers)
	if n == 0 {
		return "", nil
	}
	unsignedNumbers := make([]uint64, len(numbers))
	for i, number := range numbers {
		unsignedNumbers[i] = uint64(number)
	}
	return sqIDs.Encode(unsignedNumbers)
}

func DecodeSqids(id string) int64 {
	if id == "" {
		return 0
	}
	unsignedNumbers := sqIDs.Decode(id)
	if len(unsignedNumbers) == 0 {
		return 0
	}
	return int64(unsignedNumbers[0])
}

func GenerateUniqueID() string {
	return uuid.New().String()
}

func GetEnv() string {
	return env
}

func Transcode(input, output any) error {
	buffer := new(bytes.Buffer)
	if err := json.MarshalWrite(buffer, input); err != nil {
		return err
	}
	return json.UnmarshalRead(buffer, output)
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

func HashPassword(password string) (*string, error) {
	hashByte, err := bcrypt.GenerateFromPassword(String2ByteSlice(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return new(ByteSlice2String(hashByte)), nil
}

func MatchedHashAndPassword(encryptedPassword, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(encryptedPassword, password)
	return !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword)
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
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}()
	_, _ = fmt.Fprintf(buffer, format, a...)
	return buffer.String()
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

func SaveFile(c fiber.Ctx, file *multipart.FileHeader, filename string) error {
	if file == nil {
		return nil
	}
	return c.SaveFile(file, PhotoLocation(filename))
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
