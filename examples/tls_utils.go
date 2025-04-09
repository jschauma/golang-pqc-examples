package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"time"
)

const logo = "CyberArk AI Center of Excellence"
// getTlsCurveIDName returns the name of the curve based on the curveID
func getTlsCurveIDName(curveID tls.CurveID) (string, error) {
	curveName := ""
	switch curveID {
	case tls.CurveP256:
		curveName = "P256"
	case tls.CurveP384:
		curveName = "P384"
	case tls.CurveP521:
		curveName = "P521"
	case tls.X25519:
		curveName = "X25519"
	case 0x11ec:
		curveName = "X25519MLKEM768"
	case 0x6399:
		curveName = "X25519Kyber768Draft00"
	default:
		return "", fmt.Errorf("unknown curve ID: 0x%x", uint16(curveID))
	}
	return curveName, nil
}

// CreateSelfSignedKeyAndCertFiles generates a private key and a self-signed certificate
// and saves them to the specified files
func CreateSelfSignedKeyAndCertFiles(keyFileName, certFileName string) error {

	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}

	// Save the private key to a file
	privateKeyFile, err := os.Create(keyFileName)
	if err != nil {
		return fmt.Errorf("error creating private key file: %v", err)
	}
	defer privateKeyFile.Close()

	// Encode the private key to the PEM format
	err = pem.Encode(privateKeyFile, privateKeyPEM)
	if err != nil {
		return fmt.Errorf("error encoding private key to PEM: %v", err)
	}

	// Create a template for the certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{logo},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	// Create a self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// convert certificate DER to PEM
	cert := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	// Save the certificate to a file
	certFile, err := os.Create(certFileName)
	if err != nil {
		return fmt.Errorf("error creating cert file:%v", err)
	}

	defer func() error{
		err = certFile.Close()
		if err != nil {
			return fmt.Errorf("error closing cert file:%v ", err)			
		}
		return nil
	} ()

	// Encode the certificate to the PEM format
	err = pem.Encode(certFile, cert)
	if err != nil {
		return fmt.Errorf("error encoding cert to PEM:%v", err)
	}

	return nil
}

// getRequestCurveID returns the curve ID of the request
func getRequestCurveID(r *http.Request) (tls.CurveID, error) {
	if r.TLS == nil {
		return 0, fmt.Errorf("the request is not a TLS connection")
	}

	// Access the private 'testingOnlyCurveID' field using reflection
	connState := reflect.ValueOf(*r.TLS)
	curveIDField := connState.FieldByName("testingOnlyCurveID")

	if !curveIDField.IsValid() {
		return 0, fmt.Errorf("the curve ID field is not found")
	}

	// Convert the reflected value to tls.CurveID
	return tls.CurveID(curveIDField.Uint()), nil
}
