package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/config"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/pkg/certs"
	"google.golang.org/grpc/metadata"
)

// Manager authentication and certificate operations
type Manager struct {
	config     *config.AuthConfig
	caKey      *rsa.PrivateKey
	caCert     *x509.Certificate
	tokenCache sync.Map
	mu         sync.RWMutex
}

func NewManager() (*Manager, error) {
	cfg := config.Get().Core.Auth

	caCert, caKey, err := certs.LoadCA(cfg.CACertPath, cfg.CAKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA: %w", err)
	}

	return &Manager{
		config: &cfg,
		caKey:  caKey,
		caCert: caCert,
	}, nil
}

func (m *Manager) ValidateBootstrapToken(token string) error {
	// TODO: Implement bootstrap token validation
	if token == "" {
		return fmt.Errorf("empty bootstrap token")
	}
	return nil
}

// SignCSR signs a certificate signing request
func (m *Manager) SignCSR(csrBytes []byte) ([]byte, error) {
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("invalid CSR signature: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:       certs.GenerateSerialNumber(),
		Subject:            csr.Subject,
		PublicKey:          csr.PublicKey,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		SignatureAlgorithm: csr.SignatureAlgorithm,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, m.caCert, csr.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return certPEM, nil
}

func (m *Manager) GenerateAuthToken(nodeID string) (string, int64, error) {
	expiry := time.Now().Add(m.config.TokenDuration).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"node_id": nodeID,
		"exp":     expiry,
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(m.config.TokenSecret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	m.tokenCache.Store(nodeID, tokenString)
	return tokenString, expiry, nil
}

func (m *Manager) ValidateAuthToken(nodeID, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.TokenSecret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["node_id"] != nodeID {
			return fmt.Errorf("token node ID mismatch")
		}
		return nil
	}

	return fmt.Errorf("invalid token")
}

func (m *Manager) ValidateCertificate(certBytes []byte) error {
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	if err := cert.CheckSignatureFrom(m.caCert); err != nil {
		return fmt.Errorf("certificate not signed by our CA: %w", err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf("certificate is not valid at this time (not before: %s, not after: %s)", cert.NotBefore, cert.NotAfter)
	}

	return nil
}

func (m *Manager) RotateToken(nodeID, currentToken string) (string, int64, error) {
	if err := m.ValidateAuthToken(nodeID, currentToken); err != nil {
		return "", 0, fmt.Errorf("invalid current token: %w", err)
	}

	return m.GenerateAuthToken(nodeID)
}

func (m *Manager) GetTokenExpiry() int64 {
	return time.Now().Add(m.config.TokenDuration).Unix()
}

func (m *Manager) NodeIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found in context")
	}

	nodeID := md.Get("node_id")
	if len(nodeID) == 0 {
		return "", fmt.Errorf("node ID not found in metadata")
	}
	return nodeID[0], nil
}
