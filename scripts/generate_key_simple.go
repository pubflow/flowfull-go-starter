package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

func main() {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		fmt.Printf("Error generating key pair: %v\n", err)
		return
	}

	publicKeyHex := hex.EncodeToString(publicKey)
	privateKeyHex := hex.EncodeToString(privateKey)

	fmt.Println("===========================================")
	fmt.Println("🔑 PASETO v4 Key Pair Generated")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("📢 Public Key:")
	fmt.Println(publicKeyHex)
	fmt.Println()
	fmt.Println("🔒 Private Key (add this to your .env file):")
	fmt.Println(privateKeyHex)
	fmt.Println()
	fmt.Println("📝 Add to .env:")
	fmt.Printf("PASETO_PRIVATE_KEY=%s\n", privateKeyHex)
	fmt.Println()
	fmt.Println("⚠️  Keep the private key secret!")
	fmt.Println("===========================================")
}
