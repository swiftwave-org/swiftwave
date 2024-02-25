package worker

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
	"time"
)

func (m Manager) SSLGenerate(request SSLGenerateRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch domain
	var domain core.Domain
	err := domain.FindById(ctx, dbWithoutTx, request.DomainId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if not found, return nil as no queue is required
			return nil
		}
		return err
	}
	// If domain is IPv4, don't generate SSL
	if domain.IsIPv4() {
		return nil
	}
	// verify domain points to this server
	isDomainPointingToThisServer := m.ServiceManager.SslManager.VerifyDomain(domain.Name)
	if !isDomainPointingToThisServer {
		if domain.SSLStatus == core.DomainSSLStatusNone {
			// If SSL generation is invoked at the time of domain creation, don't mark it as failed if domain is not pointing to this server
			return nil
		}
		_ = domain.UpdateSSLStatus(ctx, dbWithoutTx, core.DomainSSLStatusFailed)
		return nil
	}
	// generate private key [if not found]
	if domain.SSLPrivateKey == "" {
		privateKey, err := generatePrivateKey()
		if err != nil {
			return err
		}
		domain.SSLPrivateKey = privateKey
		err = domain.Update(ctx, dbWithoutTx)
		if err != nil {
			return err
		}
	}
	// obtain certificate
	fullChain, err := m.ServiceManager.SslManager.ObtainCertificate(domain.Name, domain.SSLPrivateKey)
	if err != nil {
		// don' requeue, if anything happen user can anytime re-request for certificate
		return nil
	}
	// store certificate
	domain.SSLFullChain = fullChain
	// update status
	domain.SSLStatus = core.DomainSSLStatusIssued
	domain.SSLIssuedAt = time.Now()
	domain.SSLIssuer = m.ServiceManager.SslManager.FetchIssuerName()
	// update domain
	err = domain.Update(ctx, dbWithoutTx)
	if err != nil {
		return err
	}
	// generate a new transaction id for haproxy
	transactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
	if err != nil {
		return err
	}
	// upload certificate to haproxy
	err = m.ServiceManager.HaproxyManager.UpdateSSL(transactionId, domain.Name, []byte(domain.SSLPrivateKey), []byte(domain.SSLFullChain))
	if err != nil {
		return err
	}
	// commit transaction
	err = m.ServiceManager.HaproxyManager.CommitTransaction(transactionId)
	if err != nil {
		return err
	}
	return nil
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
