package main

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
	"path/filepath"
	"time"
)

type CertificateConfig struct {
	CommonName  string
	DNSNames    []string
	IPAddresses []net.IP
	IsCA        bool
	ValidYears  int
}

func main() {
	fmt.Println("🔐 Generating TLS Certificates for E-commerce Services...")
	fmt.Println()

	// Create base directory
	baseDir := "."
	caDir := filepath.Join(baseDir, "ca")
	if err := os.MkdirAll(caDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create CA directory: %v\n", err)
		os.Exit(1)
	}

	// Generate CA
	fmt.Println("🏭 Generating Certificate Authority (CA)...")
	caCert, caKey, err := generateCA(caDir)
	if err != nil {
		fmt.Printf("❌ Failed to generate CA: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CA Certificate generated: %s\n", filepath.Join(caDir, "ca-cert.pem"))
	fmt.Println()

	// Services to generate certificates for
	services := []string{
		"api-gateway",
		"order-service",
		"payment-service",
		"product-service",
		"user-service",
		"inventory-service",
		"notification-service",
	}

	// Generate service certificates
	for _, service := range services {
		fmt.Printf("🔧 Generating certificate for %s...\n", service)

		serviceDir := filepath.Join(baseDir, service)
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory for %s: %v\n", service, err)
			continue
		}

		config := CertificateConfig{
			CommonName: service,
			DNSNames: []string{
				"localhost",
				service,
				service + ".default.svc.cluster.local",
			},
			IPAddresses: []net.IP{
				net.ParseIP("127.0.0.1"),
			},
			IsCA:       false,
			ValidYears: 1,
		}

		if err := generateServiceCert(serviceDir, config, caCert, caKey); err != nil {
			fmt.Printf("❌ Failed to generate certificate for %s: %v\n", service, err)
			continue
		}

		fmt.Printf("Certificate generated for %s\n", service)
		fmt.Printf("   Cert: %s\n", filepath.Join(serviceDir, "server-cert.pem"))
		fmt.Printf("   Key:  %s\n", filepath.Join(serviceDir, "server-key.pem"))
		fmt.Println()
	}

	fmt.Println("🎉 All certificates generated successfully!")
	fmt.Println()
	fmt.Println("📋 Certificate Summary:")
	fmt.Printf("   CA Certificate: %s\n", filepath.Join(caDir, "ca-cert.pem"))
	fmt.Printf("   Services: %d certificates generated\n", len(services))
	fmt.Println()
	fmt.Println("📝 Next Steps:")
	fmt.Println("   1. Update docker-compose.yaml to mount certificates")
	fmt.Println("   2. Update service code to load TLS certificates")
	fmt.Println("   3. Configure gRPC clients to use TLS")
	fmt.Println()
	fmt.Println("⚠️  Security Warning:")
	fmt.Println("   These are self-signed certificates for DEVELOPMENT only!")
	fmt.Println()
}

func generateCA(dir string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"VN"},
			Province:           []string{"HoChiMinh"},
			Locality:           []string{"HoChiMinh"},
			Organization:       []string{"Ecommerce"},
			OrganizationalUnit: []string{"Development"},
			CommonName:         "Ecommerce Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	// Self-sign CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse CA certificate
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Save CA certificate
	caCertFile, err := os.Create(filepath.Join(dir, "ca-cert.pem"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA cert file: %w", err)
	}
	defer caCertFile.Close()

	if err := pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode CA cert: %w", err)
	}

	// Save CA private key
	caKeyFile, err := os.Create(filepath.Join(dir, "ca-key.pem"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer caKeyFile.Close()

	if err := pem.Encode(caKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to encode CA key: %w", err)
	}

	return caCert, caKey, nil
}

func generateServiceCert(dir string, config CertificateConfig, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {
	// Generate service private key
	serviceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate service key: %w", err)
	}

	// Create service certificate template
	serviceTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:            []string{"VN"},
			Province:           []string{"HoChiMinh"},
			Locality:           []string{"HoChiMinh"},
			Organization:       []string{"Ecommerce"},
			OrganizationalUnit: []string{"Development"},
			CommonName:         config.CommonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(config.ValidYears, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    config.DNSNames,
		IPAddresses: config.IPAddresses,
	}

	// Sign service certificate with CA
	serviceCertDER, err := x509.CreateCertificate(rand.Reader, serviceTemplate, caCert, &serviceKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create service certificate: %w", err)
	}

	// Save service certificate
	serviceCertFile, err := os.Create(filepath.Join(dir, "server-cert.pem"))
	if err != nil {
		return fmt.Errorf("failed to create service cert file: %w", err)
	}
	defer serviceCertFile.Close()

	if err := pem.Encode(serviceCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serviceCertDER,
	}); err != nil {
		return fmt.Errorf("failed to encode service cert: %w", err)
	}

	// Save service private key
	serviceKeyFile, err := os.Create(filepath.Join(dir, "server-key.pem"))
	if err != nil {
		return fmt.Errorf("failed to create service key file: %w", err)
	}
	defer serviceKeyFile.Close()

	if err := pem.Encode(serviceKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serviceKey),
	}); err != nil {
		return fmt.Errorf("failed to encode service key: %w", err)
	}

	return nil
}
