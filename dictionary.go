package recode

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"strings"
)

const checksumLen = 4

type dictionary struct {
	bitsToWord    map[string]string
	wordToBits    map[string]int
	maxBitsLen    int
	wordsChecksum []byte
}

type Recoder interface {
	// Encode converts the input byte slice into a mnemonic.
	Encode(data []byte) ([]string, error)

	// Decode takes a  mnemonic and returns the original byte slice.
	Decode(mnemonic []string) ([]byte, error)
}

// NewDictionary creates a new Recoder instance using the provided slice of words.
// Returns an error if there are any problems with the words.
func NewDictionary(words []string) (Recoder, error) {
	if len(words) < 2 {
		return nil, errors.New("more than 2 words are required")
	}

	bitsToWord := make(map[string]string, len(words))
	wordToBits := make(map[string]int, len(words))
	dups := make(map[string]bool, len(words))
	h := sha256.New()

	for i, word := range words {
		if word != strings.TrimSpace(word) {
			return nil, errors.New("all words should be trimmed")
		}

		if word == "" {
			return nil, errors.New("words should not be empty")
		}

		if dups[word] {
			return nil, errors.New("words should be unique")
		}
		dups[word] = true

		bitsToWord[fmt.Sprintf("%b", i)] = word
		wordToBits[word] = i
		h.Write([]byte(word))
	}

	maxBitsLen := math.Ceil(math.Log2(float64(len(words))))

	return &dictionary{
		bitsToWord:    bitsToWord,
		wordToBits:    wordToBits,
		maxBitsLen:    int(maxBitsLen),
		wordsChecksum: h.Sum(nil),
	}, nil
}

func (d *dictionary) Encode(data []byte) ([]string, error) {
	mnemonic := []string{}

	var bitsBuilder strings.Builder
	for _, b := range data {
		bitsBuilder.WriteString(fmt.Sprintf("%08b", b))
	}

	cs, err := d.checksum(data)
	if err != nil {
		return mnemonic, err
	}

	bits := bitsBuilder.String() + cs

	right, left := 0, d.maxBitsLen
	for right < len(bits) {
		if left > len(bits) {
			left = len(bits)
		}

		lb := bits[right:left]
		word, ok := d.bitsToWord[lb]
		if ok {
			mnemonic = append(mnemonic, word)
			right = left
			left = right + d.maxBitsLen
			continue
		}

		left -= 1
	}

	return mnemonic, nil
}

func (d *dictionary) Decode(mnemonic []string) ([]byte, error) {
	var bitsBuilder strings.Builder
	for _, m := range mnemonic {
		idx, ok := d.wordToBits[m]
		if !ok {
			return nil, errors.New("invalid mnemonic word")
		}
		bitsBuilder.WriteString(fmt.Sprintf("%b", idx))
	}

	bitString := bitsBuilder.String()
	if len(bitString) < checksumLen {
		return nil, errors.New("mnemonic too short for checksum")
	}
	checksum := bitString[len(bitString)-checksumLen:]
	bitString = bitString[:len(bitString)-checksumLen]

	src := []byte(bitString)
	dst := make([]byte, len(src)/8)
	var bitMask byte = 1

	bitCounter := 0
	for b := 0; b < len(bitString)/8; b++ {
		for bit := 0; bit < 8; bit++ {
			dst[b] |= (src[bitCounter] & bitMask) << (7 - bit)
			bitCounter++
		}
	}

	deccs, err := d.checksum(dst)
	if err != nil {
		return nil, err
	}

	if checksum != deccs {
		return nil, errors.New("invalid checksum")
	}

	return dst, nil
}

func (d *dictionary) checksum(data []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	_, err = h.Write(d.wordsChecksum)
	if err != nil {
		return "", err
	}

	sum := h.Sum(nil)

	return fmt.Sprintf("%08b", sum[0])[:checksumLen], nil
}

var _ Recoder = &dictionary{}
