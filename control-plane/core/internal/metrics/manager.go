package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Manager handles metrics collection and reporting
type Manager struct {
	mu sync.RWMutex

	// Node metrics
	nodeCount  *prometheus.GaugeVec
	nodeStatus *prometheus.GaugeVec

	// Resource metrics
	cpuUsage    *prometheus.GaugeVec
	memoryUsage *prometheus.GaugeVec
	diskUsage   *prometheus.GaugeVec

	// Task metrics
	activeTasks    *prometheus.GaugeVec
	completedTasks *prometheus.CounterVec
	failedTasks    *prometheus.CounterVec
}

// NewManager creates a new metrics manager
func NewManager() (*Manager, error) {
	m := &Manager{}

	m.initMetrics()

	// Register metrics with Prometheus
	prometheus.MustRegister(
		m.nodeCount,
		m.nodeStatus,
		m.cpuUsage,
		m.memoryUsage,
		m.diskUsage,
		m.activeTasks,
		m.completedTasks,
		m.failedTasks,
	)

	return m, nil
}

func (m *Manager) initMetrics() {
	m.nodeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_nodes_total",
			Help: "Total number of nodes in the mesh",
		},
		[]string{"state"},
	)

	m.nodeStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_node_status",
			Help: "Current status of nodes",
		},
		[]string{"node_id", "hostname", "state"},
	)

	m.cpuUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_node_cpu_usage",
			Help: "CPU usage percentage by node",
		},
		[]string{"node_id", "hostname"},
	)

	m.memoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_node_memory_usage_bytes",
			Help: "Memory usage in bytes by node",
		},
		[]string{"node_id", "hostname"},
	)

	m.diskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_node_disk_usage_bytes",
			Help: "Disk usage in bytes by node",
		},
		[]string{"node_id", "hostname"},
	)

	m.activeTasks = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "luminous_mesh_node_active_tasks",
			Help: "Number of active tasks by node",
		},
		[]string{"node_id", "hostname"},
	)

	m.completedTasks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "luminous_mesh_node_completed_tasks_total",
			Help: "Total number of completed tasks by node",
		},
		[]string{"node_id", "hostname"},
	)

	m.failedTasks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "luminous_mesh_node_failed_tasks_total",
			Help: "Total number of failed tasks by node",
		},
		[]string{"node_id", "hostname"},
	)
}

// UpdateNodeCount updates the total node count by state
func (m *Manager) UpdateNodeCount(state string, count float64) {
	m.nodeCount.With(prometheus.Labels{
		"state": state,
	}).Set(count)
}

// UpdateNodeStatus updates a node's status metrics
func (m *Manager) UpdateNodeStatus(nodeID, hostname, state string) {
	m.nodeStatus.With(prometheus.Labels{
		"node_id":  nodeID,
		"hostname": hostname,
		"state":    state,
	}).Set(1)
}

// UpdateNodeResources updates a node's resource metrics
func (m *Manager) UpdateNodeResources(nodeID, hostname string, cpu, memory, disk float64) {
	labels := prometheus.Labels{
		"node_id":  nodeID,
		"hostname": hostname,
	}

	m.cpuUsage.With(labels).Set(cpu)
	m.memoryUsage.With(labels).Set(memory)
	m.diskUsage.With(labels).Set(disk)
}

// UpdateNodeTasks updates a node's task metrics
func (m *Manager) UpdateNodeTasks(nodeID, hostname string, active float64, completed, failed uint64) {
	labels := prometheus.Labels{
		"node_id":  nodeID,
		"hostname": hostname,
	}

	m.activeTasks.With(labels).Set(active)
	m.completedTasks.With(labels).Add(float64(completed))
	m.failedTasks.With(labels).Add(float64(failed))
}

// RemoveNodeMetrics removes all metrics for a node
func (m *Manager) RemoveNodeMetrics(nodeID, hostname string) {
	labels := prometheus.Labels{
		"node_id":  nodeID,
		"hostname": hostname,
	}

	m.nodeStatus.Delete(labels)
	m.cpuUsage.Delete(labels)
	m.memoryUsage.Delete(labels)
	m.diskUsage.Delete(labels)
	m.activeTasks.Delete(labels)
}
