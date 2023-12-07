package cipher

import (
	"bytes"
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

// Descriptor defines
type Descriptor struct {
	Name string            //file name
	Info *x509.Certificate //basic input info

	CertDER  []byte
	KeyPKCS8 []byte
}

type SelfSigner struct {
	algorithm elliptic.Curve // used when generate private key
	random    io.Reader      // used when generate private key
	snGen     SerialNumberGenerator

	caCertificate *x509.Certificate // ca certificate
	caPrivateKey  *ecdsa.PrivateKey // ca private key
	caCertDER     []byte            // DER-encoded ca certificate buf, i.e. raw cert
	caKeyPKCS8    []byte            // DER-encoded PKCS8 ca private key buf, i.e. raw key
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
	serialNumber, err := g.Generate()
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

	// export DER-formatted cert and private key
	g.caCertDER = derBytes
	g.caKeyPKCS8 = privateKeyBytes

	g.caCertificate = ca
	g.caPrivateKey = privateKey

	return nil
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

// CreateSignCertificates creates and signs certificates according to the
// given descriptors and returns number of certificates successfully generated.
func (g *SelfSigner) CreateSignCertificates(descriptors []*Descriptor) (error, int) {
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

		count++
		desc.CertDER = cert
		desc.KeyPKCS8 = key
	}

	return nil, count
}

func (g *SelfSigner) Generate() (*big.Int, error) {
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

func (g *SelfSigner) writePemFilePair(file string, cert, key []byte) error {
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key})

	//write pem file for cert, including publick key
	f, err := os.Create(file + "-cert.pem")
	if err != nil {
		log.Errorln("create cert pem file failed:", err)
		return err
	}

	_, err = f.Write(certPem)
	if err != nil {
		log.Errorln("write pem file failed:", err)
		return err
	}

	//write pem file for private key
	f, err = os.Create(file + "-key.pem")
	if err != nil {
		log.Errorln("create key pem file failed:", err)
		return err
	}

	_, err = f.Write(keyPem)
	if err != nil {
		log.Errorln("write pem file failed:", err)
		return err
	}

	return nil
}

func (g *SelfSigner) WriteCACertFiles(file string) error {
	return g.writePemFilePair(file, g.caCertDER, g.caKeyPKCS8)
}

func (g *SelfSigner) WriteFiles(descriptors []*Descriptor) error {
	for _, desc := range descriptors {
		err := g.writePemFilePair(desc.Name, desc.CertDER, desc.KeyPKCS8)
		if err != nil {
			return err
		}
	}

	return nil
}

func sequentialBytes(n int) io.Reader {
	sequence := make([]byte, n)
	for i := 0; i < n; i++ {
		sequence[i] = byte(i)
	}
	return bytes.NewReader(sequence)
}

func randProvider() io.Reader {
	return rand.Reader
}
