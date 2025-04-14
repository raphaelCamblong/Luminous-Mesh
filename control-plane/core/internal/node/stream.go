package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/logger"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/metrics"
	pb "github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/proto"
	"go.uber.org/zap"
)

type StreamHandler struct {
	nodeID         string
	nodeManager    *Manager
	metricsManager *metrics.Manager
	commandChan    chan *pb.ControlPlaneCommand
	done           chan struct{}
	mu             sync.RWMutex
}

func NewStreamHandler(
	nodeID string,
	nodeManager *Manager,
	metricsManager *metrics.Manager,
) *StreamHandler {
	return &StreamHandler{
		nodeID:         nodeID,
		nodeManager:    nodeManager,
		metricsManager: metricsManager,
		commandChan:    make(chan *pb.ControlPlaneCommand, 100),
		done:           make(chan struct{}),
	}
}

// HandleStream handles the bidirectional stream
func (h *StreamHandler) HandleStream(stream pb.NodeService_StreamConnectionServer) error {
	// Start command sender
	go h.sendCommands(stream)

	// Process incoming status updates
	for {
		select {
		case <-h.done:
			return nil
		default:
			update, err := stream.Recv()
			if err != nil {
				logger.L().Error("Failed to receive status update",
					zap.String("node_id", h.nodeID),
					zap.Error(err),
				)
				close(h.done)
				return err
			}

			if err := h.handleStatusUpdate(update); err != nil {
				logger.L().Error("Failed to handle status update",
					zap.String("node_id", h.nodeID),
					zap.Error(err),
				)
			}
		}
	}
}

// handleStatusUpdate processes node status updates
func (h *StreamHandler) handleStatusUpdate(update *pb.NodeStatusUpdate) error {
	// Update node status
	if err := h.nodeManager.UpdateNodeStatus(h.nodeID, update.Status); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	// Update metrics
	node, err := h.nodeManager.GetNode(h.nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Update node status metrics
	h.metricsManager.UpdateNodeStatus(
		h.nodeID,
		node.BasicInfo.Hostname,
		update.Status.State.String(),
	)

	// Update resource metrics
	for _, resource := range update.Status.Resources {
		switch resource.Name {
		case "CPU":
			h.metricsManager.UpdateNodeResources(
				h.nodeID,
				node.BasicInfo.Hostname,
				resource.UsagePercentage,
				0,
				0,
			)
		case "Memory":
			h.metricsManager.UpdateNodeResources(
				h.nodeID,
				node.BasicInfo.Hostname,
				0,
				resource.UsagePercentage,
				0,
			)
		case "Disk":
			h.metricsManager.UpdateNodeResources(
				h.nodeID,
				node.BasicInfo.Hostname,
				0,
				0,
				resource.UsagePercentage,
			)
		}
	}

	// Process metrics reports
	for _, metric := range update.Metrics {
		switch metric.MetricName {
		case "node_active_tasks":
			h.metricsManager.UpdateNodeTasks(
				h.nodeID,
				node.BasicInfo.Hostname,
				metric.Value,
				0,
				0,
			)
		case "node_completed_tasks_total":
			h.metricsManager.UpdateNodeTasks(
				h.nodeID,
				node.BasicInfo.Hostname,
				0,
				uint64(metric.Value),
				0,
			)
		case "node_failed_tasks_total":
			h.metricsManager.UpdateNodeTasks(
				h.nodeID,
				node.BasicInfo.Hostname,
				0,
				0,
				uint64(metric.Value),
			)
		}
	}

	return nil
}

// sendCommands sends commands to the node
func (h *StreamHandler) sendCommands(stream pb.NodeService_StreamConnectionServer) {
	for {
		select {
		case <-h.done:
			return
		case cmd := <-h.commandChan:
			if err := stream.Send(cmd); err != nil {
				logger.L().Error("Failed to send command",
					zap.String("node_id", h.nodeID),
					zap.Error(err),
				)
				close(h.done)
				return
			}
		}
	}
}

// SendCommand sends a command to the node
func (h *StreamHandler) SendCommand(cmd *pb.ControlPlaneCommand) error {
	select {
	case h.commandChan <- cmd:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("command channel full")
	}
}

// Close closes the stream handler
func (h *StreamHandler) Close() {
	close(h.done)
}
