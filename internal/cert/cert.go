package cert

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

type CertOpt struct {
	Host       string
	ValidFor   time.Duration
	IsCA       *bool
	RSABits    int
	ECDSACurve string
	ED25519Key bool

	CertPath string
	KeyPath  string
}

func Generate(opt CertOpt) error {
	if opt.Host == "" {
		// use for SANs
		opt.Host = "localhost,localhost.localdomain,lvh.me,*.lvh.me,[::1],127.0.0.1,fe80::1"
	}
	if opt.ValidFor == 0 {
		opt.ValidFor = 30 * 24 * time.Hour
	}
	if opt.RSABits <= 0 {
		opt.RSABits = 2048
	}
	if opt.IsCA == nil {
		yes := true
		opt.IsCA = &yes
	}
	if opt.CertPath == "" {
		opt.CertPath = "server.crt"
	}
	if opt.KeyPath == "" {
		opt.KeyPath = "server.key"
	}

	var priv interface{}
	var err error
	switch opt.ECDSACurve {
	case "":
		if opt.ED25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, opt.RSABits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		log.Fatalf("Unrecognized elliptic curve: %q", opt.ECDSACurve)
	}
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(opt.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	// TODO: confirm SANs?
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(opt.Host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if *opt.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := os.Create(opt.CertPath)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", opt.CertPath, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", opt.CertPath, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", opt.CertPath, err)
	}

	keyOut, err := os.OpenFile(opt.KeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", opt.KeyPath, err)
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", opt.KeyPath, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", opt.KeyPath, err)
	}

	return nil
}

func Verify(opt CertOpt) bool {
	now := time.Now()

	// verify cert
	certBytes, err := ioutil.ReadFile(opt.CertPath)
	if err != nil {
		// TODO: error handling
		return false
	}

	certBlock, _ := pem.Decode(certBytes)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		// TODO: error handling
		return false
	}
	if cert.NotAfter.Before(now) {
		// expired
		return false
	}

	// TODO: verify key

	return true
}
