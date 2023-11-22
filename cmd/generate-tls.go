package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	generateTLSCommand.Flags().String("domain", "", "Domain name for which to generate the certificate")
}

var generateTLSCommand = &cobra.Command{
	Use:   "generate-tls",
	Short: "Generate TLS certificates for swiftwave endpoints",
	Long: `This command generates TLS certificates for swiftwave endpoints.
	It's not for generating certificates for domain of hosted applications`,
	Run: func(cmd *cobra.Command, args []string) {
		// If domain is not provided, use the domain from config
		domain := cmd.Flag("domain").Value.String()
		if strings.TrimSpace(domain) == "" {
			domain = systemConfig.ServiceConfig.AddressOfCurrentNode
		}
		// Check if there is already someone listening on port 80
		if checkIfPortIsInUse("80") {
			printError("Port 80 is already in use, please stop the process and try again")
			return
		}
		//// Start http-01 challenge server
		echoServer := echo.New()
		echoServer.HideBanner = true
		echoServer.Pre(middleware.RemoveTrailingSlash())
		// Initiating database client
		dbDialect := postgres.Open(systemConfig.PostgresqlConfig.DSN())
		dbClient, err := gorm.Open(dbDialect, &gorm.Config{})
		if err != nil {
			printError("Failed to connect to database")
			return
		}
		// Initiating SSL Manager
		options := SSL.ManagerOptions{
			IsStaging:                 systemConfig.LetsEncryptConfig.StagingEnvironment,
			Email:                     systemConfig.LetsEncryptConfig.EmailID,
			AccountPrivateKeyFilePath: systemConfig.LetsEncryptConfig.AccountPrivateKeyPath,
		}
		sslManager := SSL.Manager{}
		err = sslManager.Init(context.Background(), *dbClient, options)
		if err != nil {
			printError("Failed to initiate SSL Manager")
			return
		}
		// Start the server
		go func(sslManager *SSL.Manager) {
			sslManager.InitHttpHandlers(echoServer)
			err := echoServer.Start(":80")
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
				printSuccess("http-01 challenge server has been stopped")
				} else {
					printError("Failed to start http-01 challenge server")
					os.Exit(1)
				}
			}
		}(&sslManager)
		// Generate private key
		privateKey, err := generatePrivateKey()
		if err != nil {
			printError("Failed to generate private key")
			return
		}
		// Generate the certificate
		certificate, err := sslManager.ObtainCertificate(domain, privateKey)
		if err != nil {
			println(err.Error())
			printError("Failed to generate certificate")
			return
		}
		// Stop the http-01 challenge server
		echoServer.Server.Shutdown(context.Background())
		// Store private key and certificate in the service.ssl_certificate_dir/<domain> folder
		dir := systemConfig.ServiceConfig.SSLCertificateDir + "/" + domain
		if !checkIfFolderExists(dir) {
			err = createFolder(dir)
			if err != nil {
				printError("Failed to create folder " + dir)
				return
			}
		}
		// Store private key
		err = os.WriteFile(dir+"/private.key", []byte(privateKey), 0644)
		if err != nil {
			printError("Failed to store private key")
			return
		}
		// Store certificate
		err = os.WriteFile(dir+"/certificate.crt", []byte(certificate), 0644)
		if err != nil {
			printError("Failed to store certificate")
			return
		}
		// Print success message
		printSuccess("Successfully generated TLS certificate for " + domain)
	},
}

func generatePrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", errors.New("unable to generate private key")
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemKey := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyBytes = pem.EncodeToMemory(&pemKey)
	return string(privateKeyBytes), nil
}
