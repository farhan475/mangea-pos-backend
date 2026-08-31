package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := map[string]string{
		"admin123": "",
		"kasir123": "",
		"owner123": "",
	}

	for password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", password, err)
			continue
		}
		passwords[password] = string(hash)
	}

	fmt.Println("Bcrypt hashes for default users:")
	fmt.Println("=================================")
	fmt.Printf("admin123: %s\n", passwords["admin123"])
	fmt.Printf("kasir123: %s\n", passwords["kasir123"])
	fmt.Printf("owner123: %s\n", passwords["owner123"])
}
