package dataplaneregistry

import (
	"strconv"
	"sync"
	"time"

	"github.com/tratteria/tconfigd/common"
)

const (
	EXPIRATION_BUFFER   = 2 * time.Minute
	EXPIRATION_DURATION = time.Duration(common.DATA_PLANE_HEARTBEAT_INTERVAL_MINUTES)*time.Minute + EXPIRATION_BUFFER
	CLEANUP_INTERVAL    = 5 * EXPIRATION_DURATION
)

type Component struct {
	IpAddress     string
	Port          int
	ServiceName   string
	LastHeartbeat time.Time
}

type Manager interface {
	Register(string, int, string, string)
	UpdateHeartbeat(string, int, string, string)
	GetActiveEntries(string, string) []*Component
}

type Retriever interface {
	GetActiveEntries(string, string) []*Component
	GetAgentServices(string) []string
}

type Registry struct {
	entries map[string]map[string]map[string]*Component
	mutex   sync.RWMutex
}

func NewRegistry() *Registry {
	am := &Registry{
		entries: make(map[string]map[string]map[string]*Component),
	}

	go am.cleanupExpiredEntires()

	return am
}

func (am *Registry) Register(ip string, port int, serviceName string, namespace string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.entries[namespace] == nil {
		am.entries[namespace] = make(map[string]map[string]*Component)
	}

	if am.entries[namespace][serviceName] == nil {
		am.entries[namespace][serviceName] = make(map[string]*Component)
	}

	key := ip + strconv.Itoa(port)

	am.entries[namespace][serviceName][key] = &Component{
		IpAddress:     ip,
		Port:          port,
		ServiceName:   serviceName,
		LastHeartbeat: time.Now(),
	}
}

func (am *Registry) UpdateHeartbeat(ip string, port int, serviceName string, namespace string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := ip + strconv.Itoa(port)

	if entry, ok := am.entries[namespace][serviceName][key]; ok {
		entry.LastHeartbeat = time.Now()
	} // TODO: return error if an entry not found
}

func (am *Registry) GetActiveEntries(serviceName string, namespace string) []*Component {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeEntries []*Component

	now := time.Now()

	serviceEntries, ok := am.entries[namespace][serviceName]
	if !ok {
		return activeEntries
	}

	for _, entry := range serviceEntries {
		if now.Sub(entry.LastHeartbeat) <= EXPIRATION_DURATION {
			activeEntries = append(activeEntries, entry)
		}
	}

	return activeEntries
}

func (am *Registry) GetAgentServices(namespace string) []string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	services := make([]string, 0, len(am.entries[namespace]))

	for service := range am.entries[namespace] {
		if service != common.TRATTERIA_SERVICE_NAME {
			services = append(services, service)
		}
	}

	return services
}

func (am *Registry) cleanupExpiredEntires() {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	defer ticker.Stop()

	for {
		<-ticker.C
		am.removeExpiredEntries()
	}
}

func (am *Registry) removeExpiredEntries() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	now := time.Now()

	for _, namespaceEntries := range am.entries {
		for _, serviceEntries := range namespaceEntries {
			for key, entry := range serviceEntries {
				if now.Sub(entry.LastHeartbeat) > EXPIRATION_DURATION {
					delete(serviceEntries, key)
				}
			}
		}
	}
}
