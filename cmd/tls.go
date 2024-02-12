package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	tlsCmd.AddCommand(tlsEnableCmd)
	tlsCmd.AddCommand(tlsDisableCmd)
	tlsCmd.AddCommand(generateCertificateCommand)
	tlsCmd.AddCommand(renewCertificatesCommand)
	generateCertificateCommand.Flags().String("domain", "", "Domain name for which to generate the certificate")
}

var tlsCmd = &cobra.Command{
	Use:   "tls",
	Short: "Manage TLS for swiftwave service",
	Long:  `Manage TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var tlsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable TLS for swiftwave service",
	Long:  `Enable TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		if systemConfig.ServiceConfig.UseTLS {
			printSuccess("TLS is already enabled")
			return
		}
		// Check if some certificate is already present
		if isFolderEmpty(systemConfig.ServiceConfig.SSLCertificateDir) {
			printError("No TLS certificate found")
			printInfo("Use `swiftwave tls generate-certificate` to generate a new certificate")
			return
		}
		systemConfig.ServiceConfig.UseTLS = true
		err := systemConfig.WriteToFile(configFilePath)
		if err != nil {
			printError("Failed to update config")
			printError(err.Error())
			return
		}
		printSuccess("TLS has been enabled")
		restartSysctlService("swiftwave")
	},
}

var tlsDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable TLS for swiftwave service",
	Long:  `Disable TLS for swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		if !systemConfig.ServiceConfig.UseTLS {
			printSuccess("TLS is already disabled")
			return
		}
		systemConfig.ServiceConfig.UseTLS = false
		err := systemConfig.WriteToFile(configFilePath)
		if err != nil {
			printError("Failed to update config")
			printError(err.Error())
			return
		}
		printSuccess("TLS has been disabled")
		restartSysctlService("swiftwave")
	},
}

var generateCertificateCommand = &cobra.Command{
	Use:   "generate-certificate",
	Short: "Generate TLS certificate for swiftwave endpoints",
	Long: `This command generates TLS certificate for swiftwave endpoints.
	It's not for generating certificates for domain of hosted applications`,
	Run: func(cmd *cobra.Command, args []string) {
		// If domain is not provided, use the domain from config
		domain := cmd.Flag("domain").Value.String()
		if strings.TrimSpace(domain) == "" {
			domain = systemConfig.ServiceConfig.AddressOfCurrentNode
		}
		//// Start http-01 challenge server
		echoServer := echo.New()
		echoServer.HideBanner = true
		echoServer.Pre(middleware.RemoveTrailingSlash())
		// Initiating database client
		dbClient, err := getDBClient()
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
		// Check if there is already someone listening on port 80
		isServerStarted := false
		isPort80Blocked := checkIfPortIsInUse("80")
		isServicePortBlocked := checkIfPortIsInUse(strconv.Itoa(systemConfig.ServiceConfig.BindPort))
		if isPort80Blocked {
			if isServicePortBlocked {
				printInfo("Already running swiftwave service will be used to solve http-01 challenge")
			} else {
				printError("Please stop the service running on port 80 temporarily")
				return
			}
		} else {
			isServerStarted = true
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
		}
		// Generate private key
		privateKey, err := generatePrivateKey()
		if err != nil {
			printError("Failed to generate private key")
			os.Exit(1)
			return
		}
		// Generate the certificate
		certificate, err := sslManager.ObtainCertificate(domain, privateKey)
		if err != nil {
			println(err.Error())
			printError("Failed to generate certificate")
			os.Exit(1)
			return
		}
		if isServerStarted {
			// Stop the http-01 challenge server
			err = echoServer.Server.Shutdown(context.Background())
			if err != nil {
				return
			}
		}
		// Store private key and certificate in the service.ssl_certificate_dir/<domain> folder
		dir := systemConfig.ServiceConfig.SSLCertificateDir + "/" + domain
		if !checkIfFolderExists(dir) {
			err = createFolder(dir)
			if err != nil {
				printError("Failed to create folder " + dir)
				os.Exit(1)
				return
			}
		}
		// Store private key
		err = os.WriteFile(dir+"/private.key", []byte(privateKey), 0644)
		if err != nil {
			printError("Failed to store private key")
			os.Exit(1)
			return
		}
		// Store certificate
		err = os.WriteFile(dir+"/certificate.crt", []byte(certificate), 0644)
		if err != nil {
			printError("Failed to store certificate")
			os.Exit(1)
			return
		}
		// Print success message
		printSuccess("Successfully generated TLS certificate for " + domain)
		// Restart swiftwave service
		restartSysctlService("swiftwave")
	},
}

var renewCertificatesCommand = &cobra.Command{
	Use:   "renew-certificates",
	Short: "Renew TLS certificates for swiftwave endpoints",
	Long: `This command renews TLS certificates for swiftwave endpoints.
	It's not for renewing certificates for domain of hosted applications`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find-out domain names for which certificates are to be renewed
		files, err := os.ReadDir(systemConfig.ServiceConfig.SSLCertificateDir)
		if err != nil {
			printError("Failed to read SSL certificate directory")
			return
		}
		// find out current executable
		executablePath, err := os.Executable()
		if err != nil {
			printError("Failed to find out current executable path")
			return
		}

		for _, file := range files {
			domain := file.Name()
			certPath := filepath.Join(systemConfig.ServiceConfig.SSLCertificateDir, domain, "certificate.crt")
			isRenewalRequired, err := isRenewalImminent(certPath)
			if err != nil {
				printError("> " + domain + ": " + err.Error())
				continue
			}
			if isRenewalRequired {
				cmd := exec.Command(executablePath, "tls", "generate-certificate", "--domain", domain)
				// fetch exit code
				err := cmd.Run()
				if err != nil {
					printError("> " + domain + ": " + err.Error())
				}
				if cmd.ProcessState.Success() {
					printSuccess("> " + domain + ": certificate has been renewed")
				} else {
					printError("> " + domain + ": failed to renew certificate")
				}
			}
		}
		printInfo("Renewal process has been completed")
	},
}

// private functions

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

func isFolderEmpty(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return true
	}
	return len(files) == 0
}

func restartSysctlService(serviceName string) {
	// check if service is running
	// read the output of systemctl is-active <service_name>
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return
	}
	if strings.TrimSpace(string(output)) == "active" {
		// restart the service
		cmd = exec.Command("systemctl", "restart", serviceName)
		err = cmd.Run()
		if err != nil {
			return
		}
		printSuccess(serviceName + " has been restarted")
	}
}

func daysUntilExpiration(certPath string) (int, error) {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return 0, err
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return 0, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0, err
	}

	daysRemaining := int(cert.NotAfter.Sub(time.Now()).Hours() / 24)
	return daysRemaining, nil
}

func isRenewalImminent(certPath string) (bool, error) {
	daysRemaining, err := daysUntilExpiration(certPath)
	if err != nil {
		return false, err
	}

	return daysRemaining <= 30, nil
}
