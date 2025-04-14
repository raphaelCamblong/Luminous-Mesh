package lmgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/config"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/logger"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/auth"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/metrics"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/node"
	pb "github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents the gRPC server for the Luminous Mesh control plane
type Server struct {
	pb.UnimplementedNodeServiceServer
	config         *config.CoreConfig
	nodeManager    *node.Manager
	authManager    *auth.Manager
	metricsManager *metrics.Manager
	mu             sync.RWMutex
	grpcServer     *grpc.Server
}

// NewServer creates a new instance of the control plane server
func NewServer() (*Server, error) {
	cfg := config.Get()
	nodeManager, err := node.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create node manager: %w", err)
	}

	authManager, err := auth.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	metricsManager, err := metrics.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics manager: %w", err)
	}

	return &Server{
		config:         &cfg.Core,
		nodeManager:    nodeManager,
		authManager:    authManager,
		metricsManager: metricsManager,
	}, nil
}

// Start initializes and starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	// Load TLS credentials
	creds, err := credentials.NewServerTLSFromFile(
		s.config.TLS.CertFile,
		s.config.TLS.KeyFile,
	)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	)

	// Register services
	s.registerGrpcServices()

	// Start listening
	lis, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.L().Info("Starting gRPC server", zap.String("address", s.config.ListenAddr))

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			logger.L().Error("Failed to serve", zap.Error(err))
		}
	}()

	<-ctx.Done()
	s.Stop()
	return nil
}

func (s *Server) Stop() {
	logger.L().Info("Stopping gRPC server")
	s.grpcServer.GracefulStop()
}

func (s *Server) registerGrpcServices() {
	pb.RegisterNodeServiceServer(s.grpcServer, s)
}

// RegisterNode handles node registration requests
func (s *Server) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	// Validate bootstrap token
	if err := s.authManager.ValidateBootstrapToken(req.BootstrapToken); err != nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "Invalid bootstrap token",
		}, nil
	}

	// Process CSR and generate certificate
	cert, err := s.authManager.SignCSR(req.Csr)
	if err != nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to sign CSR: %v", err),
		}, nil
	}

	// Generate node ID and initial auth token
	nodeID := s.nodeManager.GenerateNodeID()
	authToken, _, err := s.authManager.GenerateAuthToken(nodeID)
	if err != nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: "Failed to generate auth token",
		}, nil
	}

	// Register node
	if err := s.nodeManager.RegisterNode(nodeID, req.BasicInfo); err != nil {
		return &pb.RegisterNodeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to register node: %v", err),
		}, nil
	}

	return &pb.RegisterNodeResponse{
		Success:           true,
		Message:           "Node registered successfully",
		SignedCertificate: cert,
		NodeId:            nodeID,
		InitialAuthToken:  authToken,
		ControlPlaneInfo: &pb.ControlPlaneInfo{
			ApiEndpoint:      s.config.APIEndpoint,
			CaCertificate:    []byte(s.config.TLS.CACert), // TODO: Convert to []byte
			ConnectionParams: s.config.ConnectionParams,
		},
	}, nil
}

// Authenticate handles node authentication requests
func (s *Server) Authenticate(ctx context.Context, req *pb.AuthenticationRequest) (*pb.AuthenticationResponse, error) {
	// Validate auth token
	if err := s.authManager.ValidateAuthToken(req.NodeId, req.AuthToken); err != nil {
		return &pb.AuthenticationResponse{
			Success: false,
			Message: "Invalid auth token",
		}, nil
	}

	// Validate certificate
	if err := s.authManager.ValidateCertificate(req.Certificate); err != nil {
		return &pb.AuthenticationResponse{
			Success: false,
			Message: "Invalid certificate",
		}, nil
	}

	// Create session
	sessionID, err := s.nodeManager.CreateSession(req.NodeId)
	if err != nil {
		return &pb.AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create session: %v", err),
		}, nil
	}

	// Update node info
	if err := s.nodeManager.UpdateNodeInfo(req.NodeId, req.BasicInfo, req.Capabilities); err != nil {
		return &pb.AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update node info: %v", err),
		}, nil
	}

	// Get initial configuration
	config := s.nodeManager.GetNodeConfiguration(req.NodeId)

	return &pb.AuthenticationResponse{
		Success:       true,
		Message:       "Authentication successful",
		SessionId:     sessionID,
		TokenExpiry:   s.authManager.GetTokenExpiry(),
		InitialConfig: config,
	}, nil
}

// StreamConnection handles bidirectional streaming with nodes
func (s *Server) StreamConnection(stream pb.NodeService_StreamConnectionServer) error {
	ctx := stream.Context()
	nodeID, err := s.authManager.NodeIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("node ID not found in context: %w", err)
	}

	// Create stream handler
	handler := node.NewStreamHandler(nodeID, s.nodeManager, s.metricsManager)
	return handler.HandleStream(stream)
}

// RotateToken handles token rotation requests
func (s *Server) RotateToken(ctx context.Context, req *pb.TokenRotationRequest) (*pb.TokenRotationResponse, error) {
	// Validate current token and session
	if err := s.authManager.ValidateAuthToken(req.NodeId, req.CurrentToken); err != nil {
		return nil, fmt.Errorf("invalid auth token")
	}

	if err := s.nodeManager.ValidateSession(req.NodeId, req.SessionId); err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	// Generate new token
	newToken, expiry, err := s.authManager.RotateToken(req.NodeId, req.CurrentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate token: %w", err)
	}

	return &pb.TokenRotationResponse{
		NewToken: newToken,
		Expiry:   expiry,
	}, nil
}
