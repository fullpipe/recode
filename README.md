# recode

[![test](https://github.com/fullpipe/recode/actions/workflows/test.yml/badge.svg)](https://github.com/fullpipe/recode/actions/workflows/test.yml)
[![lint](https://github.com/fullpipe/recode/actions/workflows/lint.yml/badge.svg)](https://github.com/fullpipe/recode/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/fullpipe/recode.svg)](https://pkg.go.dev/github.com/fullpipe/recode)

Package recode provides functionality to encode and decode any byte data into a mnemonic and back, using a custom word list.

## Example

```go
package main

import (
    "log"
    "github.com/fullpipe/recode"
)

func main() {
    rec, err := recode.NewDictionary(
        []string{"my", "own", "random", "dictionary", "to", "have", "more", "fun", "with", "words"},
    )
    if err != nil {
        log.Fatal(err)
    }

    // Encode the byte data
    mnemonic, err := rec.Encode([]byte("nice!"))
    if err != nil {
        log.Fatal(err)
    }

    log.Println(mnemonic) // my more fun my my more words my more ...

    // Decode the mnemonic back to byte data
    decoded, err := rec.Decode(mnemonic)
    if err != nil {
        log.Fatal(err)
        return
    }
    log.Println(string(decoded)) // nice!
}
```

You can use more familiar dictionaries like `bip39` or `slip39`:

```go
recBip, _ := recode.NewDictionary(recode.Bip39Dictionary)
recSlip, _ := recode.NewDictionary(recode.Slip39Dictionary)
```

**Beware**: The resulting mnemonic will differ from the original bip39 and slip39!

But who needs bip39 if you can use fruits & vegetables?

```go
...
entropy, _ := bip39.NewEntropy(256)

fruits, _ := recode.NewDictionary([]string{"🍇", "🍈", "🍉", "🍊", "🍋", "🍌", "🍍", "🥭", "🍎", "🍐", "🍑", "🍒", "🍓", "🫐", "🥝", "🍅", "🫒", "🥥", "🥑", "🍆", "🥔", "🥕", "🌽", "🌶️", "🫑", "🥒", "🥬", "🥦", "🧄", "🧅", "🥜", "🫘", "🌰", "🫚", "🫛"})

salat, _ := fruits.Encode(entropy)

log.Println(string(salat)) // 🍇 🥦 🍆 🍇 🥑 🫑 🥦 🍇 🥔 🫚 🍇 🍍 🍇 🌽 🍑 ...
```

## Features

- **Custom Word List**: Use your own set of words for encoding and decoding.
- **Flexible**: Works with any byte data, of any length.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any bugs or feature requests.
