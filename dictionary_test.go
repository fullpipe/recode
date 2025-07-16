package recode

import (
	"crypto/rand"
	"log"
	"math"
	r "math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name    string
		words   []string
		wantErr bool
	}{
		{
			"no words",
			[]string{},
			true,
		},
		{
			"one word",
			[]string{"foo"},
			true,
		},
		{
			"two words",
			[]string{"foo", "bar"},
			false,
		},
		{
			"tree words",
			[]string{"foo", "bar", "fizz"},
			true,
		},
		{
			"four words",
			[]string{"foo", "bar", "fizz", "buzz"},
			false,
		},
		{
			"not trimmed words",
			[]string{"foo", "bar", "fizz", "buzz "},
			true,
		},
		{
			"not unique words",
			[]string{"foo", "bar", "fizz", "fizz"},
			true,
		},
		{
			"empty word",
			[]string{"foo", "bar", "fizz", ""},
			true,
		},
		{
			"emoji",
			[]string{"foo", "bar", "buzz", "👍"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDictionary(tt.words)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDic_Encode(t *testing.T) {
	tests := []struct {
		name    string
		words   []string
		data    []byte
		want    []string
		wantErr bool
	}{
		{
			"zero one",
			[]string{"0", "1"},
			[]byte("42"),
			[]string{"0", "0", "0", "1", "1", "0", "1", "0", "0", "0", "0", "1", "1", "0", "0", "1", "0"},
			false,
		},
		{
			"base",
			[]string{"foo", "bar", "fizz", "buzz"},
			[]byte("1"),
			[]string{"fizz", "foo", "buzz", "foo", "bar"},
			false,
		},
		{
			"nice",
			[]string{"🍇", "🍈", "🍉", "🍊", "🍋", "🍌", "🍍", "🥭", "🍎", "🍐", "🍑", "🍒", "🍓", "🫐", "🥝", "🍅", "🫒", "🥥", "🥑", "🍆", "🥔", "🥕", "🌽", "🌶️", "🫑", "🥒", "🥬", "🥦", "🧄", "🧅", "🥜", "🫘"},
			[]byte("nice!"),
			[]string{"🫒", "🫐", "🥒", "🥔", "🌽", "🍍", "🥒", "🍐", "🍈"},
			false,
		},
		{
			"empty data gives just checksum",
			[]string{"foo", "bar", "fizz", "buzz"},
			[]byte{},
			[]string{"fizz"},
			false,
		},
		{
			"nice with bib39",
			Bip39Dictionary,
			[]byte("nice!"),
			[]string{"kit", "hover", "enrich", "sun", "dumb"},
			false,
		},
		{
			"my own random dictionary",
			[]string{"my", "own", "random", "words", "to", "have", "more", "fun"},
			[]byte("nice!"),
			[]string{"have", "words", "words", "to", "more", "to", "have", "to", "words", "words", "own", "random", "random", "my", "fun"},
			false,
		},
		{
			// 0101010 1001  00000111111 11111000000 01111111110 01010001000 00000010101 0 00101010 11
			// festival among way lemon extra actor betray
			"uses full dictionary to encode",
			Bip39Dictionary,
			[]byte{7, 255, 1, 255, 40, 128, 42, 42},
			[]string{"festival", "among", "way", "lemon", "extra", "actor", "betray"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDictionary(tt.words)
			assert.NoError(t, err)

			got, err := d.Encode(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.want, got)

			dec, err := d.Decode(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, dec)
		})
	}
}

func TestDic_Encode_Wallet(t *testing.T) {
	rec, err := NewDictionary([]string{"🍇", "🍈", "🍉", "🍊", "🍋", "🍌", "🍍", "🥭", "🍎", "🍐", "🍑", "🍒", "🍓", "🫐", "🥝", "🍅", "🫒", "🥥", "🥑", "🍆", "🥔", "🥕", "🌽", "🌶️", "🫑", "🥒", "🥬", "🥦", "🧄", "🧅", "🥜", "🫘"})
	assert.NoError(t, err)

	// 128 bit
	entropy := []byte{138, 252, 148, 132, 177, 104, 151, 15, 191, 157, 140, 195, 148, 64, 81, 116}

	salat, err := rec.Encode(entropy)
	assert.NoError(t, err)

	assert.Equal(t, []string{"🍒", "🥥", "🍒", "🥜", "🍐", "🍐", "🍈", "🍌", "🥥", "🫐", "🍉", "🍒", "🫒", "🫘", "🍅", "🧄", "🧅", "🥥", "🍆", "🍈", "🥒", "🍎", "🫒", "🍉", "🥥", "🥝", "🍆"}, salat)

	decodedEntropy, err := rec.Decode(salat)
	assert.NoError(t, err)
	assert.Equal(t, entropy, decodedEntropy)
}

// https://www.blockplate.com/pages/bip-39-wordlist
func TestDic_Decode(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		mnemonic []string
		want     []byte
		wantErr  bool
	}{
		{
			"valid",
			Bip39Dictionary,
			[]string{"festival", "among", "way", "lemon", "extra", "actor", "betray"},
			[]byte{7, 255, 1, 255, 40, 128, 42, 42},
			false,
		},
		{
			"empty",
			Bip39Dictionary,
			[]string{},
			nil,
			true,
		},
		{
			"invalid word for checksum",
			Bip39Dictionary,
			[]string{"WTF", "among", "way", "lemon", "extra", "actor", "betray"},
			nil,
			true,
		},
		{
			"invalid word in mnemonic",
			Bip39Dictionary,
			[]string{"festival", "WTF", "way", "lemon", "extra", "actor", "betray"},
			nil,
			true,
		},
		{
			"invalid checksum",
			Bip39Dictionary,
			[]string{"fire", "among", "way", "lemon", "extra", "actor", "betray"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := NewDictionary(tt.words)
			got, err := d.Decode(tt.mnemonic)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dic.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dic.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDic_Random(t *testing.T) {
	for i := 0; i < 10000; i++ {
		wordsExp := r.IntN(10) + 1
		wordsNum := int(math.Pow(2, float64(wordsExp)))
		words := make([]string, 0, wordsNum)
		for j := 0; j < wordsNum; j++ {
			words = append(words, randomWord())
		}

		d, err := NewDictionary(words)
		assert.NoError(t, err)

		l := r.IntN(512)
		data := make([]byte, l)
		encoded, err := d.Encode(data)
		assert.NoError(t, err)

		decoded, err := d.Decode(encoded)
		assert.NoError(t, err)

		assert.EqualValues(t, data, decoded)
	}
}

func randomWord() string {
	l := r.IntN(32) + 32
	b := make([]byte, l)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(b))
}

func Test_dictionary_idxToBitString(t *testing.T) {
	tests := []struct {
		name       string
		maxBitsLen int
		idx        int
		want       string
	}{
		{
			"one",
			4,
			1,
			"0001",
		},
		{
			"two",
			4,
			2,
			"0010",
		},
		{
			"two",
			5,
			2,
			"00010",
		},
		{
			"dish",
			11,
			505,
			"00111111001",
		},
		{
			"spy",
			11,
			1690,
			"11010011010",
		},
		{
			"spy",
			12,
			1690,
			"011010011010",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := idxToBitString(tt.idx, tt.maxBitsLen); got != tt.want {
				t.Errorf("dictionary.idxToBitString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tailBitsLenInChecksum(t *testing.T) {
	tests := []struct {
		name          string
		bitsBatchSize int
		want          int
	}{
		{
			"for 11 bits in batch (bip39) its 4, as possible padding length 10 < 2^4",
			11,
			4,
		},
		{
			// with 32 words we have 5 bits in batch
			// with 5 bits max tail is 4 bits long
			// to encode 4 to bits we need next to it 2^N
			// and its 8 = 2^3
			"32 words",
			5,
			3,
		},
		{
			"2 words",
			1,
			0,
		},
		{
			"4 words",
			2,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tailBitsLenInChecksum(tt.bitsBatchSize))
		})
	}
}
