package cipher

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var defaultCA = &x509.Certificate{
	NotBefore: time.Now(),
	NotAfter:  time.Now().AddDate(100, 0, 0),
	Subject: pkix.Name{
		Country:            []string{"CN"},
		Organization:       []string{"xxx Co.,Ltd."},
		OrganizationalUnit: []string{"www.example.com"},
		Province:           []string{"Chongqing"},
		CommonName:         "example CA",
	},
	EmailAddresses: []string{"ca@example.com"},
}

var descs = []*Descriptor{
	{
		Name:       "device",
		EncodeType: EncodeDER,
		Info: &x509.Certificate{
			NotBefore: time.Now(),
			NotAfter:  time.Now().AddDate(10, 0, 0),
			Subject: pkix.Name{
				Country:      []string{"CN"},
				Organization: []string{"xxx technologies co.,ltd."},
				Province:     []string{"Beijing"},
				CommonName:   "xxx device",
			},
			EmailAddresses: []string{"dev@xxx.com"},
		},
	}, {
		Name:       "server",
		EncodeType: EncodePEM,
		Info: &x509.Certificate{
			NotBefore: time.Now(),
			NotAfter:  time.Now().AddDate(20, 0, 0),
			Subject: pkix.Name{
				Country:      []string{"CN"},
				Organization: []string{"xxx technologies co.,ltd."},
				Province:     []string{"Beijing"},
				CommonName:   "xxx serfver",
			},
			EmailAddresses: []string{"dev@xxx.com"},
			DNSNames:       []string{"www.example.com"},
		},
	},
}

func TestNewSelfSigner(t *testing.T) {
	s := NewSelfSigner()
	assert.NotNil(t, s)

	assert.NoError(t, s.GenerateCA(defaultCA))

	t.Logf("cert: %s\n", s.CACert())
	t.Logf("key: %s\n", s.CAPrivateKey())

	assert.NoError(t, s.WriteCACertFiles("ca"))

	err, i := s.CreateSelfSignedCertificates(descs)
	assert.NoError(t, err)
	assert.Equal(t, i, 2)

	assert.NoError(t, s.WriteFiles(descs))
}

func TestSelfSigner_LoadAndGenerate(t *testing.T) {
	s := NewSelfSigner()
	assert.NoError(t, s.LoadCAFromFiles("ca.crt", "ca.key"))

	err, i := s.CreateSelfSignedCertificates(descs)
	assert.NoError(t, err)
	assert.Equal(t, i, 2)

	assert.NoError(t, s.WriteFiles(descs))
}

func TestSelfSigner_CreateSelfSigned(t *testing.T) {
	s := NewSelfSigner()
	assert.NotNil(t, s)

	_, _, err := s.CreateSelfSigned(&x509.Certificate{
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(20, 0, 0),
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Organization: []string{"xxx technologies co.,ltd."},
			Province:     []string{"Beijing"},
			CommonName:   "xxx serfver",
		},
		EmailAddresses: []string{"dev@xxx.com"},
		DNSNames:       []string{"www.example.com"},
	})
	assert.Error(t, err)
}
