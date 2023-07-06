package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
	// "github.com/mholt/acmez"
)

const tokenFile = "tokens.txt" 

func main() {
	var wg sync.WaitGroup

	ctx := context.Background()
	
	keyData, err := ioutil.ReadFile("private_key2.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Parse the PEM-encoded data
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		fmt.Println("Invalid PEM file or key type")
		return
	}

	// Parse the DER-encoded key data
	accountPrivateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
	}

	account := acme.Account{
		Contact:              []string{"mailto:tanmoysrt@gmail.com"},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	// Create a new ACME client
	client := acmez.Client{
		Client: &acme.Client{
			Directory: "https://acme-staging-v02.api.letsencrypt.org/directory",
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // REMOVE THIS FOR PRODUCTION USE!
					},
				},
			},
		},
		ChallengeSolvers: map[string]acmez.Solver{
			acme.ChallengeTypeHTTP01:    mySolver{}, // provide these!
		},
	}

	// only once
	// account, err = client.GetAccount(ctx, account)
	// if err != nil {
	// 	fmt.Println("new account: %v", err)
	// 	return
	// }
	fmt.Println("Account created")
	fmt.Println(account)

	// Generate Certificate key
	certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	fmt.Println("Generating certificate key...")
	if err != nil {
		fmt.Println("failed to generating certificate key: %v", err)
		return
	}
	err = storeKeyToFile("cert_key.pem", certPrivateKey)
	if err != nil {
		fmt.Println("failed to storing certificate key: %v", err)
		return
	}

	return

	// Start verification server
	wg.Add(1)
	go func() {
		http.HandleFunc("/.well-known/acme-challenge/",  helloHandler)
		fmt.Println("Server listening on port 80...")
		http.ListenAndServe(":80", nil)
	}()

	// Request certificate
	// Obtain a certificate for the domain
	domains := []string{"minc.tanmoy.info"}
	certs, err := client.ObtainCertificate(ctx, account, certPrivateKey, domains)
	if err != nil {
		fmt.Println("failed to obtaining certificate: %v", err)
		return
	}

	for _, cert := range certs {
		fmt.Printf("Certificate %q:\n%s\n\n", cert.URL, cert.ChainPEM)
	}
	fmt.Println("done")
	wg.Wait()
}


func helloHandler(w http.ResponseWriter, r *http.Request) {
	var token string;
	token = strings.ReplaceAll(r.URL.Path, "/.well-known/acme-challenge/", "");
	fmt.Println("Searching "+token+" in "+tokenFile)
	fullToken, err := findTokenByPrefix(token)
	if err != nil {
		fmt.Fprintf(w, "Token not found!")
		return
	}
	fmt.Fprintf(w, fullToken)
}

func storeTokenToFile(token string) error {
	file, err := os.OpenFile(tokenFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(token + "\n")
	if err != nil {
		return err
	}
	writer.Flush()

	return nil
}

// findTokenByPrefix finds a token from the file that matches the given prefix
func findTokenByPrefix(prefix string) (string, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		token := scanner.Text()
		if strings.HasPrefix(token, prefix+".") {
			return token, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("matching token not found")
}

func storeKeyToFile(keyFile string, key *ecdsa.PrivateKey) error {
	// Encode the private key to PEM format
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		fmt.Println(err)
	}

	pemKey := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	// Create the PEM file
	file, err := os.Create(keyFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Write the PEM-encoded key to the file
	err = pem.Encode(file, &pemKey)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

type mySolver struct{}

func (s mySolver) Present(ctx context.Context, chal acme.Challenge) error {
	fmt.Printf("[DEBUG] store token: %s", chal.KeyAuthorization)
	storeTokenToFile(chal.KeyAuthorization)
	return nil
}

func (s mySolver) CleanUp(ctx context.Context, chal acme.Challenge) error {
	fmt.Printf("[DEBUG] cleanup: %s", chal.KeyAuthorization)
	return nil
}