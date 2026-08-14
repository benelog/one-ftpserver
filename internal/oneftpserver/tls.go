package oneftpserver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"
)

// certificateValidity is how long the generated certificate lasts. The server
// is meant to be short lived, but a year avoids surprising anyone who leaves
// one running.
const certificateValidity = 365 * 24 * time.Hour

// selfSignedTLSConfig builds a certificate in memory, for this run only.
// Keeping it out of the file system is what lets --ssl stay a single flag:
// there is no key store to create beforehand and no file left behind.
func selfSignedTLSConfig() (*tls.Config, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate a private key: %w", err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, fmt.Errorf("cannot generate a serial number: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "one-ftpserver"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(certificateValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           certificateIPs(),
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("cannot generate a certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{der},
			PrivateKey:  key,
		}},
		MinVersion: tls.VersionTLS12,
	}, nil
}

// certificateIPs lists the addresses the certificate is valid for: the
// loopback ones and every address of the machine, so that the name a client
// connects with matches whichever interface it came in through.
func certificateIPs() []net.IP {
	ips := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		network, ok := addr.(*net.IPNet)
		if ok && !network.IP.IsLoopback() {
			ips = append(ips, network.IP)
		}
	}

	return ips
}

// localIP is the address to show in the usage examples: the first non
// loopback address of the machine, so that the printed commands can be run
// from another machine on the same network. It falls back to the loopback
// address when the machine has no other one.
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		network, ok := addr.(*net.IPNet)
		if ok && !network.IP.IsLoopback() && network.IP.To4() != nil {
			return network.IP.String()
		}
	}

	return "127.0.0.1"
}
