package cmd

import "github.com/spf13/cobra"

func init(){
	generateTLSCommand.Flags().String("domain", "", "Domain name for which to generate the certificate")
}

var generateTLSCommand = &cobra.Command{
	Use:   "generate-tls",
	Short: "Generate TLS certificates for swiftwave endpoints",
	Long: `This command generates TLS certificates for swiftwave endpoints.
	It's not for generating certificates for domain of hosted applications`,
	Run: func(cmd *cobra.Command, args []string) {
		// If domain is not provided, use the domain from config
		// Create the folder if not exists
		// Check if there is already someone listening on port 80
		// Start http-01 challenge server
		// Generate the certificate
		// Stop the http-01 challenge server
		// Store private key and certificate in the service.ssl_certificate_dir/<domain> folder
	},
}
