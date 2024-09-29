package hashcash

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	version1       = "1"
	delimiter      = ":"
	timeFormat     = "20060102150405"
	hashcashFormat = "1:%d:%s:%s::%s:%s" // official format 1:zeroBits:date:senderData:extension:randomSeed:counter
	zeroBit        = '0'
)

var (
	ErrIncorrectVersion            = errors.New("incorrect version")
	ErrIncorrectHeaderFormat       = errors.New("incorrect header format")
	ErrHashLengthLessThanZeroBits  = errors.New("hash length cannot be less than zeroBit bits")
	ErrZeroBitsMustBeMoreThanZero  = errors.New("zeroBit bits must be more than zeroBit")
	ErrZeroBitsMustBeNumber        = errors.New("zeroBit bits must be a number")
	ErrMaxTriesExceeded            = errors.New("calculation exceeded max tries limit")
	ErrIncorrectDate               = errors.New("incorrect date")
	ErrCounterMustBePositiveNumber = errors.New("counter must be a positive number")
)

type HashCash struct {
	zeroBits   int
	date       time.Time
	senderData string
	extension  string
	randomSeed []byte
	counter    int
}

func NewHashCash(zeroBits int, sender string) (*HashCash, error) {
	if zeroBits <= 0 {
		return nil, ErrZeroBitsMustBeMoreThanZero
	}

	rand, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		return nil, err
	}

	return &HashCash{
		zeroBits:   zeroBits,
		date:       time.Now().UTC().Truncate(time.Second),
		senderData: sender,
		extension:  "",
		randomSeed: rand.Bytes(),
		counter:    0,
	}, nil
}

func NewHashCashFromString(data string) (*HashCash, error) {
	var err error

	parts := strings.Split(data, delimiter)
	if len(parts) < 7 {
		return nil, ErrIncorrectHeaderFormat
	}

	version := parts[0]
	if version != version1 {
		return nil, ErrIncorrectVersion
	}

	zeroBits, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, ErrZeroBitsMustBeNumber
	}

	if zeroBits <= 0 {
		return nil, ErrZeroBitsMustBeMoreThanZero
	}

	date, err := time.ParseInLocation(timeFormat, parts[2], time.UTC)
	if err != nil {
		return nil, ErrIncorrectDate
	}

	counterString, err := base64.StdEncoding.DecodeString(parts[len(parts)-1])
	if err != nil {
		return nil, fmt.Errorf("read counter error: %v", err)
	}

	counter, err := strconv.Atoi(string(counterString))
	if err != nil {
		return nil, ErrCounterMustBePositiveNumber
	}

	if counter < 0 {
		return nil, ErrCounterMustBePositiveNumber
	}

	randomSeed, err := base64.StdEncoding.DecodeString(parts[len(parts)-2])
	if err != nil {
		return nil, fmt.Errorf("read random seed error: %v", err)
	}

	extension := parts[len(parts)-3]

	sender := strings.Join(parts[3:len(parts)-3], delimiter)

	return &HashCash{
		zeroBits:   zeroBits,
		senderData: sender,
		date:       date,
		extension:  extension,
		randomSeed: randomSeed,
		counter:    counter,
	}, nil
}

func (h *HashCash) Hash() (string, error) {
	var err error

	hash := sha1.New()

	_, err = hash.Write([]byte(h.String()))
	if err != nil {
		return "", fmt.Errorf("sha1 write error: %v", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (h *HashCash) String() string {
	return fmt.Sprintf(
		hashcashFormat,
		h.zeroBits,
		h.date.Format(timeFormat),
		h.senderData,
		base64.StdEncoding.EncodeToString(h.randomSeed),
		base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(h.counter))),
	)
}

func (h *HashCash) Validate() (bool, error) {
	hash, err := h.Hash()
	if err != nil {
		return false, err
	}

	if len(hash) < h.zeroBits {
		return false, ErrHashLengthLessThanZeroBits
	}

	for _, s := range hash[:h.zeroBits] {
		if s != zeroBit {
			return false, nil
		}
	}

	return true, nil
}

func (h *HashCash) Calculate(maxTries int) error {
	if maxTries > 0 {
		h.counter = 0
		for h.counter <= maxTries {
			ok, err := h.Validate()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			h.counter++
		}
	}

	return ErrMaxTriesExceeded
}

func (h *HashCash) Counter() int {
	return h.counter
}

func (h *HashCash) CheckSender(clientID string) bool {
	return h.senderData == clientID
}

func (h *HashCash) Expired(ttl time.Duration) bool {
	return h.date.Add(ttl).Before(time.Now().UTC())
}
