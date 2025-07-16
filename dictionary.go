package recode

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

type dictionary struct {
	bitsToWord    map[string]string
	wordToBits    map[string]string
	bitsToInt     map[string]int
	bitsBatchSize int
	wordsChecksum []byte
	checksumLen   int
	// how many bit in checksum are for tail len
	// bitsBatchSize = checksumLen + tailChecksumLen
	tailChecksumLen int
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

	bitLenRaw := math.Log2(float64(len(words)))
	if bitLenRaw != float64(int(bitLenRaw)) {
		return nil, errors.New("dictionary should be complete, len(words) == 2^N")
	}
	bitsBatchSize := int(bitLenRaw)

	bitsToWord := make(map[string]string, len(words))
	wordToBits := make(map[string]string, len(words))
	bitsToInt := make(map[string]int, len(words))
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

		bitWord := idxToBitString(i, bitsBatchSize)
		bitsToWord[bitWord] = word
		wordToBits[word] = bitWord
		bitsToInt[bitWord] = i

		h.Write([]byte(word))
	}

	tailChecksumLen := tailBitsLenInChecksum(bitsBatchSize)
	checksumLen := bitsBatchSize - tailChecksumLen

	return &dictionary{
		bitsToWord:      bitsToWord,
		wordToBits:      wordToBits,
		bitsToInt:       bitsToInt,
		bitsBatchSize:   bitsBatchSize,
		wordsChecksum:   h.Sum(nil),
		checksumLen:     checksumLen,
		tailChecksumLen: tailChecksumLen,
	}, nil
}

func tailBitsLenInChecksum(bitsBatchSize int) int {
	tailChecksumLen := 0
	if bitsBatchSize > 1 {
		tailChecksumLen = int(math.Ceil(math.Log2(float64(bitsBatchSize))))
	}

	return tailChecksumLen
}

// padByteSlice returns a byte slice of the given size with contents of the
// given slice left padded and any empty spaces filled with 0's.
func padByteSlice(slice []byte, length int) []byte {
	offset := length - len(slice)
	if offset <= 0 {
		return slice
	}
	newSlice := make([]byte, length)
	copy(newSlice[offset:], slice)
	return newSlice
}

func idxToBitString(idx int, bitLen int) string {
	n := big.NewInt(int64(idx))
	b := padByteSlice(n.Bytes(), 2)

	str := fmt.Sprintf("%08b", b[0]) + fmt.Sprintf("%08b", b[1])

	return str[16-bitLen:]
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

	bits := bitsBuilder.String()

	// how many bits we should take from last word
	tailLen := len(bits) % d.bitsBatchSize
	tailLenBits := idxToBitString(tailLen, d.bitsBatchSize)
	tailLenBits = tailLenBits[len(tailLenBits)-d.tailChecksumLen:]

	// add checksum at the begining
	// so when decoding we dont care about its paddings
	bits = cs + tailLenBits + bits

	for i := 0; i < len(bits)-tailLen; i += d.bitsBatchSize {
		lb := bits[i : i+d.bitsBatchSize]
		word, ok := d.bitsToWord[lb]
		if !ok {
			return mnemonic, errors.New("this should not exists")
		}

		mnemonic = append(mnemonic, word)
	}

	if tailLen > 0 {
		tailBits := bits[len(bits)-tailLen:]
		tailBits += strings.Repeat("1", d.bitsBatchSize-tailLen)
		tailWord, ok := d.bitsToWord[tailBits]
		if !ok {
			return mnemonic, errors.New("this should not exists")
		}
		mnemonic = append(mnemonic, tailWord)
	}

	return mnemonic, nil
}

func (d *dictionary) Decode(mnemonic []string) ([]byte, error) {
	if len(mnemonic) == 0 {
		return nil, errors.New("empty mnemonic")
	}

	checksumTailBits, ok := d.wordToBits[mnemonic[0]]
	if !ok {
		return nil, errors.New("invalid mnemonic words")
	}

	checksum, tailLenBits := checksumTailBits[:d.checksumLen], checksumTailBits[d.checksumLen:]

	tailLen := 0
	if d.tailChecksumLen > 0 {
		tailLenBits = strings.Repeat("0", d.bitsBatchSize-d.tailChecksumLen) + tailLenBits
		tailLen, ok = d.bitsToInt[tailLenBits]
		if !ok {
			return nil, errors.New("invalid tail")
		}
	}

	var bitsBuilder strings.Builder
	for i := 1; i < len(mnemonic); i++ {
		wordBits, ok := d.wordToBits[mnemonic[i]]
		if !ok {
			return nil, errors.New("invalid mnemonic word")
		}
		bitsBuilder.WriteString(wordBits)
	}

	bitString := bitsBuilder.String()
	if tailLen > 0 {
		paddingLen := d.bitsBatchSize - tailLen
		bitString = bitString[:len(bitString)-paddingLen]
	}

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

// checksum calculates bit string one word length
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
	str := fmt.Sprintf("%08b", sum[0]) + fmt.Sprintf("%08b", sum[1])

	return str[:d.checksumLen], nil
}

var _ Recoder = &dictionary{}
