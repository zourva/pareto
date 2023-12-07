package cipher

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSelfSigner(t *testing.T) {
	s := NewSelfSigner()
	assert.NotNil(t, s)

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

	assert.NoError(t, s.GenerateCA(defaultCA))
	assert.NoError(t, s.WriteCACertFiles("ca"))

	descs := []*Descriptor{
		{
			Name: "device",
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
			Name: "server",
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
	err, i := s.CreateSignCertificates(descs)
	assert.NoError(t, err)
	assert.Equal(t, i, 2)

	assert.NoError(t, s.WriteFiles(descs))
}
