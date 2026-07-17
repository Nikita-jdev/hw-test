package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {

	if utf8.RuneCountInString(str) == 0 {
		return "", nil
	}

	var unpackStr strings.Builder
	var previousRune rune
	isInt := false

	for si, v := range str {
		i, err := strconv.Atoi(string(v))
		if err != nil {
			unpackStr.WriteString(string(v))
			previousRune = v
			isInt = false
		} else {
			if isInt {
				return "", ErrInvalidString
			}

			if si == 0 {
				return "", ErrInvalidString
			}

			if i == 0 {
				minusRune := []rune(unpackStr.String())
				unpackStr.Reset()
				unpackStr.WriteString(string(minusRune[:len(minusRune)-1]))
				isInt = true
				continue
			}

			unpackStr.WriteString(strings.Repeat(string(previousRune), i-1))
			isInt = true
		}
	}

	return unpackStr.String(), nil
}
