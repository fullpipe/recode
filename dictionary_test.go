package recode

import (
	"crypto/rand"
	"log"
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
			"not trimmed words",
			[]string{"foo", "bar "},
			true,
		},
		{
			"not unique words",
			[]string{"foo", "bar", "foo"},
			true,
		},
		{
			"empty word",
			[]string{"foo", "bar", ""},
			true,
		},
		{
			"ok",
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
			"base",
			[]string{"foo", "bar", "buzz"},
			[]byte("1"),
			[]string{"foo", "foo", "bar", "buzz", "foo", "foo", "bar", "buzz", "foo", "foo"},
			false,
		},
		{
			"nice",
			[]string{"🍇", "🍈", "🍉", "🍊", "🍋", "🍌", "🍍", "🥭", "🍎", "🍐", "🍑", "🍒", "🍓", "🫐", "🥝", "🍅", "🫒", "🥥", "🥑", "🍆", "🥔", "🥕", "🌽", "🌶️", "🫑", "🥒", "🥬", "🥦", "🧄", "🧅", "🥜", "🫘", "🌰", "🫚", "🫛"},
			[]byte("nice!"),
			[]string{"🍇", "🥦", "🍆", "🍇", "🥑", "🫑", "🥦", "🍇", "🍇", "🥔", "🫚", "🍇", "🍍"},
			false,
		},
		{
			"empty data gives just checksum",
			[]string{"foo", "bar", "buzz"},
			[]byte{},
			[]string{"buzz", "buzz"},
			false,
		},
		{
			"nice with bib39",
			Bip39Dictionary,
			[]byte("nice!"),
			[]string{"abandon", "system", "normal", "raw", "drill"},
			false,
		},
		{
			"my own random dictionary",
			[]string{"my", "own", "random", "dictionary", "to", "have", "more", "fun", "with", "words"},
			[]byte("nice!"),
			[]string{"my", "more", "fun", "my", "my", "more", "words", "my", "more", "my", "my", "more", "more", "my", "have", "my", "my", "with", "my", "more", "random"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := NewDictionary(tt.words)
			got, err := d.Encode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dic.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dic.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDic_Decode(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		mnemonic []string
		want     []byte
		wantErr  bool
	}{
		{
			"base",
			[]string{"foo", "bar", "buzz"},
			[]string{"foo", "foo", "bar", "buzz", "foo", "foo", "bar", "buzz", "foo", "foo"},
			[]byte("1"),
			false,
		},
		{
			"invalid words",
			[]string{"foo", "bar", "buzz"},
			[]string{"WTF", "foo", "bar", "buzz", "foo", "foo", "bar", "buzz", "foo", "foo"},
			nil,
			true,
		},
		{
			"invalid checksumm",
			[]string{"foo", "bar", "buzz"},
			[]string{"foo", "foo", "bar", "buzz", "foo", "foo", "bar", "buzz", "foo", "bar"},
			nil,
			true,
		},
		{
			"nice with bib39",
			Bip39Dictionary,
			[]string{"abandon", "system", "normal", "raw", "drill"},
			[]byte("nice!"),
			false,
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
		wordsNum := r.IntN(100) + 2
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
