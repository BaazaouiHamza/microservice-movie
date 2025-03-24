package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"movieexample.com/pkg/discovery"
)

type ServiceName string
type InstanceID string

// Registry defines an in-memory service registry.
type Registry struct {
	sync.RWMutex
	serviceAddrs map[ServiceName]map[InstanceID]*serviceInstance
}

type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

var _ discovery.Registry = (*Registry)(nil)

// NewRegistry creates a new in-memory service
// registry instance.
func NewRegistry() *Registry {
	return &Registry{serviceAddrs: map[ServiceName]map[InstanceID]*serviceInstance{}}
}

// Register creates a service record in the registry.
func (r *Registry) Register(ctx context.Context, instanceID string,
	serviceName string, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		r.serviceAddrs[ServiceName(serviceName)] = map[InstanceID]*serviceInstance{}
	}
	r.serviceAddrs[ServiceName(serviceName)][InstanceID(instanceID)] = &serviceInstance{
		hostPort:   hostPort,
		lastActive: time.Now(),
	}
	return nil
}

// Deregister removes a service record from the
// registry.
func (r *Registry) Deregister(ctx context.Context, instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		return nil
	}
	delete(r.serviceAddrs[ServiceName(serviceName)], InstanceID(instanceID))
	return nil
}

// ReportHealthyState is a push mechanism for
// reporting healthy state to the registry.
func (r *Registry) ReportHealthyState(instanceID string, serviceName string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[ServiceName(serviceName)]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.serviceAddrs[ServiceName(serviceName)][InstanceID(instanceID)]; !ok {
		return errors.New("service instance is not registered yet")
	}
	r.serviceAddrs[ServiceName(serviceName)][InstanceID(serviceName)].lastActive = time.Now()
	return nil
}

// ServiceAddresses returns the list of addresses of
// active instances of the given service
func (r *Registry) ServiceAddresses(ctx context.Context,
	serviceName string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.serviceAddrs[ServiceName(serviceName)]) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for _, v := range r.serviceAddrs[ServiceName(serviceName)] {
		if v.lastActive.Before(time.Now().Add((-5 * time.Second))) {
			continue
		}
		res = append(res, v.hostPort)
	}

	return res, nil

}
