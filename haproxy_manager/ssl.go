package haproxymanager

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

// UpdateSSL : Add SSL certificate to HAProxy
func (s Manager) UpdateSSL(transactionId string, domain string, privateKey []byte, fullChain []byte) error {
	_ = transactionId
	// Create a new buffer
	var buffer bytes.Buffer
	// Add the full chain
	buffer.Write(fullChain)
	// Add a new line
	buffer.WriteString("\n")
	// Add the private key
	buffer.Write(privateKey)

	// Is update
	updateSSLRequired := false

	ioReader := bytes.NewReader(buffer.Bytes())
	domainSanitizedName := strings.ReplaceAll(domain, ".", "_") + ".pem"

	// Try to Upload the file
	res, err := s.uploadSSL("/services/haproxy/storage/ssl_certificates", domainSanitizedName, ioReader)
	if err != nil {
		return errors.New("error while uploading ssl certificate :" + err.Error())
	}
	if !isValidStatusCode(res.StatusCode) {
		if res.StatusCode != 409 {
			return errors.New("error while uploading ssl certificate with status code" + strconv.Itoa(res.StatusCode))
		} else {
			// Already exists, so update
			updateSSLRequired = true
		}
	}
	if updateSSLRequired {
		ioReader := bytes.NewReader(buffer.Bytes())
		// Try to Upload the file
		res, err := s.replaceSSL("/services/haproxy/storage/ssl_certificates/"+domainSanitizedName, domainSanitizedName, ioReader)
		if err != nil {
			return errors.New("error while updating ssl certificate :" + err.Error())
		}
		if !isValidStatusCode(res.StatusCode) {
			return errors.New("error while updating ssl certificate with status code" + strconv.Itoa(res.StatusCode))
		}
	}
	return nil
}
