package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/dataplaneregistry"
	"go.uber.org/zap"
)

const (
	TRATTERIA_JWKS_ENDPOINT = "/.well-known/jwks.json"
)

type Service struct {
	dataPlaneRegistryManager dataplaneregistry.Manager
	httpClient               *http.Client
	logger                   *zap.Logger
}

func NewService(dataPlaneRegistryManager dataplaneregistry.Manager, httpClient *http.Client, logger *zap.Logger) *Service {
	return &Service{
		dataPlaneRegistryManager: dataPlaneRegistryManager,
		httpClient:               httpClient,
		logger:                   logger,
	}
}

type registrationResponse struct {
	HeartBeatIntervalMinutes int `json:"heartBeatIntervalMinutes"`
}

func (s *Service) RegisterAgent(ipaddress string, port int, serviceName string, namespace string) *registrationResponse {
	// TODO: return rules belonging to the service
	s.dataPlaneRegistryManager.Register(ipaddress, port, serviceName, namespace)

	return &registrationResponse{
		HeartBeatIntervalMinutes: common.DATA_PLANE_HEARTBEAT_INTERVAL_MINUTES,
	}
}

func (s *Service) RegisterHeartBeat(ipaddress string, port int, serviceName string, namespace string) {
	// TODO: if updateHeartBeat fails, notify agent to register
	// TODO: if an agent is heartbeating with an old rule version id, notify it to fetch the latest rules
	s.dataPlaneRegistryManager.UpdateHeartbeat(ipaddress, port, serviceName, namespace)
}

// TODO: Implement parallel processing of HTTP requests using goroutines.
func (s *Service) CollectJwks(ctx context.Context, namespace string) (jwk.Set, error) {
	tratteriaInstances := s.dataPlaneRegistryManager.GetActiveEntries(common.TRATTERIA_SERVICE_NAME, namespace)
	if len(tratteriaInstances) == 0 {
		return nil, fmt.Errorf("no active tratteria instances found for namespace: %s", namespace)
	}

	allKeys := jwk.NewSet()

	for _, instance := range tratteriaInstances {
		url := fmt.Sprintf("http://%s:%d/%s", instance.IpAddress, instance.Port, TRATTERIA_JWKS_ENDPOINT)

		set, err := jwk.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS from URL %s: %w", url, err)
		}

		for iter := set.Iterate(ctx); iter.Next(ctx); {
			pair := iter.Pair()
			if key, ok := pair.Value.(jwk.Key); ok {
				allKeys.Add(key)
			}
		}
	}

	return allKeys, nil
}
