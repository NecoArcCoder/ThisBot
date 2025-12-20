package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// duration foramt: "YYYY-MM-dd"
func GenerateRootCA(organization string, duration string) bool {
	// Split duration
	var durationAry = strings.Split(duration, "-")

	year, err := strconv.ParseInt(durationAry[0], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in year segment, it[%d] must be >=0\n", year)
		return false
	}
	month, err := strconv.ParseInt(durationAry[1], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in month segment, it[%d] must be >=0\n", month)
		return false
	}
	day, err := strconv.ParseInt(durationAry[2], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in day segment, it[%d] must be >=0\n", day)
		return false
	}
	// Check organization
	if len(organization) == 0 {
		fmt.Printf("[💀] Error in organization segment, it[%s] must not be empty\n", organization)
		return false
	}

	rootKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	rootCertTmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().AddDate(int(year), int(month), int(day)), // One year validation
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	rootCertDER, err := x509.CreateCertificate(rand.Reader, &rootCertTmpl, &rootCertTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		fmt.Println("[💀] Failed to create CA root certificate")
		return false
	}

	// Save root CA
	if err = os.WriteFile("root.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCertDER}), 0644); err != nil {
		fmt.Println("[💀] Failed to create root.pem")
		return false
	}

	if err = os.WriteFile("root.key", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey)}), 0600); err != nil {
		fmt.Println("[💀] Failed to create root.key")
		return false
	}
	fmt.Println("[✅] Generate CA root certificate successfully")
	return true
}

func ResignCertificate(organization string, duration string, domains []string, ips []net.IP) bool {
	// Split duration
	var durationAry = strings.Split(duration, "-")

	year, err := strconv.ParseInt(durationAry[0], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in year segment, it[%d] must be >=0\n", year)
		return false
	}
	month, err := strconv.ParseInt(durationAry[1], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in month segment, it[%d] must be >=0\n", month)
		return false
	}
	day, err := strconv.ParseInt(durationAry[2], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in day segment, it[%d] must be >=0\n", day)
		return false
	}
	// Check organization
	if len(organization) == 0 {
		fmt.Printf("[💀] Error in organization segment, it[%s] must not be empty\n", organization)
		return false
	}

	// Read root CA private key(root.key)
	rootKeyPem, err := os.ReadFile(DefaultRootKeyPath)
	if err != nil {
		fmt.Println("[💀] Failed to read root.key")
		return false
	}
	rootKeyBlock, _ := pem.Decode(rootKeyPem)
	rootKey, err := x509.ParsePKCS1PrivateKey(rootKeyBlock.Bytes)
	if err != nil {
		fmt.Println("[💀] Failed to parse root.key")
		return false
	}
	// Read root CA certificate(root.pem)
	rootCertPem, err := os.ReadFile(DefaultRootCertPath)
	if err != nil {
		fmt.Println("[💀] Failed to read root.pem")
		return false
	}
	rootCertBlock, _ := pem.Decode(rootCertPem)
	rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		fmt.Println("[💀] Failed to parse root.pem")
		return false
	}
	serverKeyPem, err := os.ReadFile(DefaultServerKeyPath)
	if err != nil {
		fmt.Println("[💀] Failed to read server.key")
		return false
	}
	serverKeyBlock, _ := pem.Decode(serverKeyPem)
	if serverKeyBlock == nil || serverKeyBlock.Type != "RSA PRIVATE KEY" {
		fmt.Println("[💀] Invalid server.key format")
		return false
	}
	serverKey, err := x509.ParsePKCS1PrivateKey(serverKeyBlock.Bytes)
	if err != nil {
		fmt.Println("[💀] Failed to parse server.key")
		return false
	}
	// Create a x509 certificate template
	serverCertTmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   strconv.FormatInt(time.Now().UTC().UnixMilli(), 10),
			Organization: []string{organization},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().AddDate(int(year), int(month), int(day)),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		// SAN
		DNSNames:    domains,
		IPAddresses: ips,
	}
	// Sign server certificate by Root CA
	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverCertTmpl, rootCert, &serverKey.PublicKey, rootKey)
	if err != nil {
		fmt.Println("[💀] Failed to sign server certificate")
		return false
	}
	// Save server.crt
	serverCrtPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertDER,
	})
	// Backup the old one
	os.Rename(DefaultServerCertPath, "./server.crt.bak")
	// Save new server.crt
	if err = os.WriteFile(DefaultServerCertPath, serverCrtPem, 0644); err != nil {
		fmt.Println("[💀] Failed to save server.key")
		return false
	}
	fmt.Println("[✅] Generate new signed server certificate successfully")
	return true
}

func GenerateCertificate(organization string, duration string, domains []string, ips []net.IP) bool {
	// Split duration
	var durationAry = strings.Split(duration, "-")

	year, err := strconv.ParseInt(durationAry[0], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in year segment, it[%d] must be >=0\n", year)
		return false
	}
	month, err := strconv.ParseInt(durationAry[1], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in month segment, it[%d] must be >=0\n", month)
		return false
	}
	day, err := strconv.ParseInt(durationAry[2], 10, 64)
	if err != nil {
		fmt.Printf("[💀] Error in day segment, it[%d] must be >=0\n", day)
		return false
	}
	// Check organization
	if len(organization) == 0 {
		fmt.Printf("[💀] Error in organization segment, it[%s] must not be empty\n", organization)
		return false
	}

	// Read root CA private key
	rootKeyPEM, err := os.ReadFile(DefaultRootKeyPath)
	if err != nil {
		fmt.Println("[💀] Failed to read root.key")
		return false
	}
	rootKeyBlock, _ := pem.Decode(rootKeyPEM)
	rootKey, err := x509.ParsePKCS1PrivateKey(rootKeyBlock.Bytes)
	if err != nil {
		fmt.Println("[💀] Failed to parse root.key")
		return false
	}
	// Read root CA certificate
	rootCertPEM, err := os.ReadFile(DefaultRootCertPath)
	if err != nil {
		fmt.Println("[💀] Failed to read root.pem")
		return false
	}
	rootCertBlock, _ := pem.Decode(rootCertPEM)
	rootCert, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		fmt.Println("[💀] Failed to parse root.pem")
		return false
	}

	// Generate new server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("[💀] Failed to generate server private key")
		return false
	}

	serverCertTmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   strconv.FormatInt(time.Now().UTC().UnixMilli(), 10),
			Organization: []string{organization},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().AddDate(int(year), int(month), int(day)),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		// SAN
		DNSNames:    domains,
		IPAddresses: ips,
	}

	// Sign server certificate by Root CA
	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverCertTmpl, rootCert, &serverKey.PublicKey, rootKey)
	if err != nil {
		fmt.Println("[💀] Failed to sign server certificate")
		return false
	}
	// Save server.crt
	serverCrtPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertDER,
	})
	if err = os.WriteFile(DefaultServerCertPath, serverCrtPEM, 0644); err != nil {
		fmt.Println("[💀] Failed to save server.crt")
		return false
	}
	// Save server.key
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})
	if err = os.WriteFile(DefaultServerKeyPath, serverKeyPEM, 0644); err != nil {
		fmt.Println("[💀] Failed to save server.key")
		return false
	}
	fmt.Println("[✅] Server certificate generated")
	return true
}
