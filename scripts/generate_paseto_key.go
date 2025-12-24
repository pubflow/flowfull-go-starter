package main

import (
	"fmt"

	"github.com/pubflow/flowfull-go-starter/internal/lib/tokens"
)

func main() {
	publicKey, privateKey, err := tokens.GenerateKeyPair()
	if err != nil {
		fmt.Printf("Error generating key pair: %v\n", err)
		return
	}

	fmt.Println("===========================================")
	fmt.Println("PASETO v4 Key Pair Generated")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("Public Key:")
	fmt.Println(publicKey)
	fmt.Println()
	fmt.Println("Private Key (add this to your .env file):")
	fmt.Println(privateKey)
	fmt.Println()
	fmt.Println("Add to .env:")
	fmt.Printf("PASETO_PRIVATE_KEY=%s\n", privateKey)
	fmt.Println()
	fmt.Println("⚠️  Keep the private key secret!")
	fmt.Println("===========================================")
}

