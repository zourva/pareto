package cipher

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"math/big"
	"os"
)

type EncodeType = int

const (
	EncodeDER EncodeType = iota
	EncodePEM
)

// Descriptor defines
type Descriptor struct {
	Name string            //file name
	Info *x509.Certificate //basic input info

	EncodeType  EncodeType
	CertEncoded []byte
	KeyEncoded  []byte
}

type SelfSigner struct {
	algorithm elliptic.Curve // used when generate private key
	random    io.Reader      // used when generate private key
	snGen     SerialNumberGenerator

	caCertificate *x509.Certificate // ca certificate
	caPrivateKey  *ecdsa.PrivateKey // ca private key

	// used only when generate a ca cert-key pair
	caCertPEM []byte // PEM-encoded ca certificate buf
	caKeyPEM  []byte // PEM-encoded PKCS8 ca private key buf
}

func NewSelfSigner() *SelfSigner {
	g := &SelfSigner{
		algorithm: elliptic.P256(),
		random:    rand.Reader,
		snGen:     &randSerialNum{},
	}

	return g
}

func (g *SelfSigner) UseAlgorithm(alg elliptic.Curve) {
	g.algorithm = alg
}

func (g *SelfSigner) UseRandProvider(rand io.Reader) {
	g.random = rand
}

func (g *SelfSigner) UseSerialNumberGenerator(gen SerialNumberGenerator) {
	g.snGen = gen
}

// LoadCAFromFiles loads certificate and private key data from
// the given file path. Note that both are required to be PEM-encoded.
func (g *SelfSigner) LoadCAFromFiles(certFile, keyFile string) error {
	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return err
	}

	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}

	return g.LoadCAFromBuf(certBytes, keyBytes)
}

// LoadCAFromBuf loads certificate and private key data from
// the given slices. Note that both are required to be PEM-encoded.
func (g *SelfSigner) LoadCAFromBuf(certPEM, keyPEM []byte) error {
	tlsCert, err := LoadKeyAndCertificate(keyPEM, certPEM)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	g.caCertificate = cert
	g.caPrivateKey = tlsCert.PrivateKey.(*ecdsa.PrivateKey)

	return nil
}

// CACert returns the PEM-encoded ca certificate data.
func (g *SelfSigner) CACert() []byte {
	return g.caCertPEM
}

// CAPrivateKey returns the PEM-encoded ca private key data.
func (g *SelfSigner) CAPrivateKey() []byte {
	return g.caKeyPEM
}

// GenerateCA creates a deterministic certificate authority.
// The following fields input are extracted:
//
//	NotBefore - mandatory
//	NotAfter  - mandatory
//	Subject   - mandatory
//	IPAddresses - optional
//	EmailAddresses - optional
func (g *SelfSigner) GenerateCA(info *x509.Certificate) error {
	if info == nil ||
		info.NotBefore.IsZero() ||
		info.NotAfter.IsZero() ||
		len(info.Subject.String()) == 0 {
		return errors.New("necessity must not be nil")
	}

	// generate private-public key pair
	privateKey, err := g.genPrivateKey()
	if err != nil {
		return err
	}

	// generate serial number
	serialNumber, err := g.genCASerialNumber()
	if err != nil {
		return err
	}

	// copy info
	ca := &x509.Certificate{
		NotBefore:      info.NotBefore,
		NotAfter:       info.NotAfter,
		Subject:        info.Subject,
		IPAddresses:    info.IPAddresses,
		EmailAddresses: info.EmailAddresses,

		IsCA:         true,
		SerialNumber: serialNumber,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// set parent == template so to make it self-signed.
	// privateKey.PublicKey holds the generated public key.
	// the returned slice is the certificate in DER encoding.
	derBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Errorln("x509 CreateCertificate failed:", err)
		return err
	}

	// convert private key to PKCS #8, DER form
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Errorln("x509 MarshalPKCS8PrivateKey failed:", err)
		return err
	}

	// export PEM-formatted cert and private key
	g.caCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	g.caKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})

	g.caCertificate = ca
	g.caPrivateKey = privateKey

	return nil
}

// CreateSelfSignedCertificates creates and signs certificates according to the
// given descriptors and returns number of certificates successfully generated.
func (g *SelfSigner) CreateSelfSignedCertificates(descriptors []*Descriptor) (error, int) {
	count := 0
	for _, desc := range descriptors {
		if len(desc.Name) == 0 {
			return errors.New("name is mandatory"), 0
		}

		if desc.Info == nil ||
			desc.Info.NotBefore.IsZero() ||
			desc.Info.NotAfter.IsZero() ||
			len(desc.Info.Subject.String()) == 0 {
			return errors.New("info NotBefore/NotAfter/Subject is mandatory"), 0
		}

		cert, key, err := g.CreateSelfSigned(desc.Info)
		if err != nil {
			return err, count
		}

		err, desc.CertEncoded, desc.KeyEncoded = g.encode(desc.EncodeType, cert, key)
		if err != nil {
			return err, count
		}

		count++
	}

	return nil, count
}

// CreateSelfSigned creates a certificate based on the given input.
// The following fields input are extracted:
//
//		NotBefore - mandatory
//		NotAfter  - mandatory
//		Subject   - mandatory
//		IPAddresses - optional
//		EmailAddresses - optional
//	 DNSNames - optional
func (g *SelfSigner) CreateSelfSigned(info *x509.Certificate) (cert, key []byte, err error) {
	if g.caCertificate == nil || g.caPrivateKey == nil {
		log.Errorln("certificate and key of ca are not initialized yet")
		return nil, nil, errors.New("ca certificate and key not valid")
	}

	pk, err := g.genPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := g.snGen.Generate()
	if err != nil {
		return
	}

	template := x509.Certificate{
		NotBefore:      info.NotBefore,
		NotAfter:       info.NotAfter,
		Subject:        info.Subject,
		EmailAddresses: info.EmailAddresses,
		IPAddresses:    info.IPAddresses,
		DNSNames:       info.DNSNames,

		SerialNumber: serialNumber,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth},
	}

	// raw DER encoding certificate content
	cert, err = x509.CreateCertificate(rand.Reader, &template,
		g.caCertificate, &pk.PublicKey, g.caPrivateKey)
	if err != nil {
		log.Errorln("x509 CreateCertificate failed:", err)
		return
	}

	key, err = x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return
	}

	return
}

func (g *SelfSigner) encode(t EncodeType, certDER, keyDER []byte) (err error, encCert []byte, encKey []byte) {
	switch t {
	case EncodeDER:
		return nil, certDER, keyDER
	case EncodePEM:
		encCert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
		encKey = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
		return nil, encCert, encKey
	default:
		return errors.New("not supported encoding"), nil, nil
	}
}

func (g *SelfSigner) genCASerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		log.Errorln("generate serial number failed:", err)
		return nil, err
	}

	return sn, nil
}

func (g *SelfSigner) genPrivateKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(g.algorithm, g.random)
	if err != nil {
		log.Errorln("generate private key failed:", err)
		return nil, err
	}

	return key, nil
}

// writePemFilePair write a .crt and .key file separately for certificate
// and private key buffer.
func (g *SelfSigner) writePemFilePair(file string, cert, key []byte) error {
	//write pem file for cert, including public key
	f, err := os.Create(file + ".crt")
	if err != nil {
		log.Errorln("create cert pem file failed:", err)
		return err
	}

	_, err = f.Write(cert)
	if err != nil {
		log.Errorln("write pem file failed:", err)
		return err
	}

	//write pem file for private key
	f, err = os.Create(file + ".key")
	if err != nil {
		log.Errorln("create key pem file failed:", err)
		return err
	}

	_, err = f.Write(key)
	if err != nil {
		log.Errorln("write pem file failed:", err)
		return err
	}

	return nil
}

func (g *SelfSigner) WriteCACertFiles(file string) error {
	return g.writePemFilePair(file, g.caCertPEM, g.caKeyPEM)
}

func (g *SelfSigner) WriteFiles(descriptors []*Descriptor) error {
	for _, desc := range descriptors {
		err := g.writePemFilePair(desc.Name, desc.CertEncoded, desc.KeyEncoded)
		if err != nil {
			return err
		}
	}

	return nil
}
