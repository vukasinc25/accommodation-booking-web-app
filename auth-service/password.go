package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Blacklist map[string]bool

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckHashedPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func NewBlacklistFromURL() (Blacklist, error) {
	url := "https://raw.githubusercontent.com/OWASP/passfault/master/wordlists/wordlists/500-worst-passwords.txt"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bl := make(Blacklist)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		item := strings.TrimSpace(scanner.Text())
		bl.Add(item)
	}

	return bl, nil
}

func (bl Blacklist) Add(item string) {
	bl[item] = true
}

// IsBlacklisted checks if an item is in the blacklist.
func (bl Blacklist) IsBlacklisted(item string) bool {
	return bl[item]
}
