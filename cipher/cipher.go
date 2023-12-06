package cipher

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
)

var (
	errBlockIsNotCertificate = errors.New("block is not a certificate, unable to load certificates")
	errNoCertificateFound    = errors.New("no certificate found, unable to load certificates")
	errBlockIsNotPrivateKey  = errors.New("block is not a private key, unable to load key")
	errNoPrivateKeyFound     = errors.New("no private key found, unable to load key")
	errUnknownPrivateKeyType = errors.New("unknown key time in PKCS#8 wrapping, unable to load key")
)

// LoadCertificateFromFile loads cert from file
func LoadCertificateFromFile(path string) (*tls.Certificate, error) {
	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadCertificate(certBytes)
}

// LoadCertificate loads cert from bytes
func LoadCertificate(certBytes []byte) (*tls.Certificate, error) {
	var certificate tls.Certificate

	for {
		block, rest := pem.Decode(certBytes)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			return nil, errBlockIsNotCertificate
		}

		certificate.Certificate = append(certificate.Certificate, block.Bytes)
		certBytes = rest
	}

	if len(certificate.Certificate) == 0 {
		return nil, errNoCertificateFound
	}

	return &certificate, nil
}

func LoadKeyFromFile(path string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return LoadKey(keyBytes)
}

// LoadKey loads key from bytes
func LoadKey(keyBytes []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil || !strings.HasSuffix(block.Type, "PRIVATE KEY") {
		return nil, errBlockIsNotPrivateKey
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errUnknownPrivateKeyType
		}
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, errNoPrivateKeyFound
}

func LoadKeyAndCertificateFromFile(keyPath, certPath string) (*tls.Certificate, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return LoadKeyAndCertificate(keyBytes, certBytes)
}

// LoadKeyAndCertificate loads client certificate
func LoadKeyAndCertificate(keyBytes []byte, certBytes []byte) (*tls.Certificate, error) {
	certificate, err := LoadCertificate(certBytes)
	if err != nil {
		return nil, err
	}
	key, err := LoadKey(keyBytes)
	if err != nil {
		return nil, err
	}
	certificate.PrivateKey = key
	return certificate, nil
}

func LoadCertPoolFromFile(path string) (*x509.CertPool, error) {
	caBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadCertPool(caBytes)
}

// LoadCertPool loads cert pool from ca certificate
func LoadCertPool(caBytes []byte) (*x509.CertPool, error) {
	rootCertificate, err := LoadCertificate(caBytes)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	for _, certBytes := range rootCertificate.Certificate {
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			certPool = nil
			return nil, err
		}
		certPool.AddCert(cert)
	}

	return certPool, nil
}
