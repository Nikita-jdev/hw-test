package hw10programoptimization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	suffix := "." + domain

	dec := json.NewDecoder(r)
	for {
		var user User
		err := dec.Decode(&user)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("get users error: %w", err)
		}

		email := user.Email
		if !hasDomain(email, suffix) {
			continue
		}

		at := strings.LastIndexByte(email, '@')
		result[strings.ToLower(email[at+1:])]++
	}
	return result, nil
}

func hasDomain(email, suffix string) bool {
	return len(email) >= len(suffix) && strings.EqualFold(email[len(email)-len(suffix):], suffix)
}
