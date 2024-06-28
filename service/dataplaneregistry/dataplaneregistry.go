package dataplaneregistry

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/tconfigderrors"
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
	Register(string, int, string)
	UpdateHeartbeat(string, int, string)
	GetActiveEntries(string) ([]*Component, error)
}

type Retriever interface {
	GetActiveEntries(string) ([]*Component, error)
}

type Registry struct {
	entries map[string]map[string]*Component
	mutex   sync.RWMutex
}

func NewRegistry() *Registry {
	am := &Registry{
		entries: make(map[string]map[string]*Component),
	}

	go am.cleanupExpiredEntires()

	return am
}

func (am *Registry) Register(ip string, port int, serviceName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.entries[serviceName] == nil {
		am.entries[serviceName] = make(map[string]*Component)
	}

	key := ip + strconv.Itoa(port)

	am.entries[serviceName][key] = &Component{
		IpAddress:     ip,
		Port:          port,
		ServiceName:   serviceName,
		LastHeartbeat: time.Now(),
	}
}

func (am *Registry) UpdateHeartbeat(ip string, port int, serviceName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := ip + strconv.Itoa(port)

	if entry, ok := am.entries[serviceName][key]; ok {
		entry.LastHeartbeat = time.Now()
	} // TODO: return error if an entry not found
}

func (am *Registry) GetActiveEntries(serviceName string) ([]*Component, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeEntries []*Component

	now := time.Now()

	serviceEntries, ok := am.entries[serviceName]
	if !ok {
		return nil, fmt.Errorf("%w: service %s", tconfigderrors.ErrNotFound, serviceName)
	}

	for _, entry := range serviceEntries {
		if now.Sub(entry.LastHeartbeat) <= EXPIRATION_DURATION {
			activeEntries = append(activeEntries, entry)
		}
	}

	return activeEntries, nil
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

	for _, serviceEntries := range am.entries {
		for key, entry := range serviceEntries {
			if now.Sub(entry.LastHeartbeat) > EXPIRATION_DURATION {
				delete(serviceEntries, key)
			}
		}
	}
}
