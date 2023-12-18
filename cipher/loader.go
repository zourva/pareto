package cipher

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	log "github.com/sirupsen/logrus"
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

// LoadAllCertPool loads system default pool and appends the additional
// ca certs provided in additionalCAFiles to the pool.
// The additional ca cert files are expected to be PEM-encoded.
// Returns the system pool plus the additional.
func LoadAllCertPool(additionalCAFiles []string) (*x509.CertPool, error) {
	// load system default pool
	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Errorln("x509.SystemCertPool failed:", err)
		return nil, err
	}

	for _, file := range additionalCAFiles {
		caBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		pool.AppendCertsFromPEM(caBytes)
		log.Debugf("load cert pool for %s", file)
	}

	return pool, nil
}

// LoadCertPoolFromFile loads certs from the ca file and
// returns as a new cert pool.
// The system cert pool is not loaded and caBytes should be PEM-encoded.
func LoadCertPoolFromFile(file string) (*x509.CertPool, error) {
	caBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return LoadCertPoolFromPEM(caBytes)
}

// LoadCertPoolFromPEM loads certs from the ca certificate and
// returns as a new cert pool.
// The system cert pool is not loaded and caBytes should be PEM-encoded.
func LoadCertPoolFromPEM(caBytes []byte) (*x509.CertPool, error) {
	rootCertificate, err := UnmarshalCert(caBytes)
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

// LoadKeyAndCertificateFromFile loads key and cert for a single pair.
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

// LoadKeyAndCertificate combines a tls.Certificate using the given cert and key binary data.
func LoadKeyAndCertificate(keyBytes []byte, certBytes []byte) (*tls.Certificate, error) {
	certificate, err := UnmarshalCert(certBytes)
	if err != nil {
		return nil, err
	}

	key, err := UnmarshalPemKey(keyBytes)
	if err != nil {
		return nil, err
	}

	certificate.PrivateKey = key

	return certificate, nil
}

// LoadCertificateFromFile loads a single PEM-encoded cert from the file.
func LoadCertificateFromFile(file string) (*tls.Certificate, error) {
	certBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return UnmarshalCert(certBytes)
}

// LoadKeyFromFile loads a single PMM-encoded private key from the file.
func LoadKeyFromFile(file string) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return UnmarshalPemKey(keyBytes)
}

// UnmarshalCert unmarshal binary data to a tls.Certificate.
func UnmarshalCert(certBytes []byte) (*tls.Certificate, error) {
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

// UnmarshalPemKey unmarshal binary data to a crypto.PrivateKey
func UnmarshalPemKey(keyBytes []byte) (crypto.PrivateKey, error) {
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
