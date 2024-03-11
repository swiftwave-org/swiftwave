package containermanger

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

// InitializeAsManager initializes the swarm as a manager
func (m Manager) InitializeAsManager() error {
	_, err := m.client.SwarmInit(m.ctx, swarm.InitRequest{
		ForceNewCluster: true,
		ListenAddr:      "0.0.0.0:2377",
	})
	return err
}

// JoinSwarm joins the swarm
func (m Manager) JoinSwarm(address string, token string) error {
	// Try to leave swarm if already joined
	_ = m.client.SwarmLeave(m.ctx, true)
	return m.client.SwarmJoin(m.ctx, swarm.JoinRequest{
		JoinToken:   token,
		ListenAddr:  "0.0.0.0:2377",
		RemoteAddrs: []string{address},
	})
}

// LeaveSwarm leaves the swarm
func (m Manager) LeaveSwarm() error {
	return m.client.SwarmLeave(m.ctx, true)
}

// RemoveNode removes a node from the swarm
func (m Manager) RemoveNode(hostname string) error {
	return m.client.NodeRemove(m.ctx, hostname, types.NodeRemoveOptions{
		Force: true,
	})
}

// PromoteToManager promotes a node to manager
func (m Manager) PromoteToManager(hostname string) error {
	// fetch node
	node, _, err := m.client.NodeInspectWithRaw(m.ctx, hostname)
	if err != nil {
		return err
	}
	// promote node
	node.Spec.Role = swarm.NodeRoleManager
	return m.client.NodeUpdate(m.ctx, hostname, node.Version, node.Spec)
}

// DemoteToWorker demotes a node to worker
func (m Manager) DemoteToWorker(hostname string) error {
	// fetch node
	node, _, err := m.client.NodeInspectWithRaw(m.ctx, hostname)
	if err != nil {
		return err
	}
	// demote node
	node.Spec.Role = swarm.NodeRoleWorker
	return m.client.NodeUpdate(m.ctx, hostname, node.Version, node.Spec)
}

// ListNodes lists all nodes
func (m Manager) ListNodes() (*map[string]swarm.Node, error) {
	nodes, err := m.client.NodeList(m.ctx, types.NodeListOptions{})
	if err != nil {
		return nil, err
	}
	nodeMap := make(map[string]swarm.Node)
	for _, node := range nodes {
		nodeMap[node.Description.Hostname] = node
	}
	return &nodeMap, nil
}

// MarkNodeAsActive marks a node as active
func (m Manager) MarkNodeAsActive(hostname string) error {
	// fetch node
	node, _, err := m.client.NodeInspectWithRaw(m.ctx, hostname)
	if err != nil {
		return err
	}
	// mark node as active
	node.Spec.Availability = swarm.NodeAvailabilityActive
	return m.client.NodeUpdate(m.ctx, hostname, node.Version, node.Spec)
}

// MarkNodeAsDrained marks a node as drained
func (m Manager) MarkNodeAsDrained(hostname string) error {
	// fetch node
	node, _, err := m.client.NodeInspectWithRaw(m.ctx, hostname)
	if err != nil {
		return err
	}
	// mark node as drained
	node.Spec.Availability = swarm.NodeAvailabilityDrain
	return m.client.NodeUpdate(m.ctx, hostname, node.Version, node.Spec)
}

// GenerateManagerJoinToken generates a manager join token
func (m Manager) GenerateManagerJoinToken() (token string, err error) {
	// fetch swarm info
	info, err := m.client.SwarmInspect(m.ctx)
	if err != nil {
		return "", err
	}
	// return token and address
	return info.JoinTokens.Manager, nil
}

// GenerateWorkerJoinToken generates a worker join token
func (m Manager) GenerateWorkerJoinToken() (token string, err error) {
	// fetch swarm info
	info, err := m.client.SwarmInspect(m.ctx)
	if err != nil {
		return "", err
	}
	// return token and address
	return info.JoinTokens.Worker, nil
}
