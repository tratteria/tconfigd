package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/dataplaneregistry"
	"github.com/tratteria/tconfigd/tratteriacontroller"
	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
	"go.uber.org/zap"
)

const (
	TRATTERIA_JWKS_ENDPOINT = ".well-known/jwks.json"
)

type Service struct {
	dataPlaneRegistryManager dataplaneregistry.Manager
	tratteriaController      *tratteriacontroller.TratteriaController
	x509Source               *workloadapi.X509Source
	tratteriaSpiffeId        spiffeid.ID
	logger                   *zap.Logger
}

func NewService(dataPlaneRegistryManager dataplaneregistry.Manager, tratteriaController *tratteriacontroller.TratteriaController, x509Source *workloadapi.X509Source, tratteriaSpiffeId spiffeid.ID, logger *zap.Logger) *Service {
	return &Service{
		dataPlaneRegistryManager: dataPlaneRegistryManager,
		tratteriaController:      tratteriaController,
		x509Source:               x509Source,
		tratteriaSpiffeId:        tratteriaSpiffeId,
		logger:                   logger,
	}
}

type registrationResponse struct {
	HeartBeatIntervalMinutes int                                 `json:"heartBeatIntervalMinutes"`
	VerificationRules        *tratteria1alpha1.VerificationRules `json:"verificationRules,omitempty"`
	GenerationRules          *tratteria1alpha1.GenerationRules   `json:"generationRules,omitempty"`
}

func (s *Service) RegisterService(ipaddress string, port int, serviceName string, namespace string) (*registrationResponse, error) {
	response := &registrationResponse{
		HeartBeatIntervalMinutes: common.DATA_PLANE_HEARTBEAT_INTERVAL_MINUTES,
	}

	if serviceName == common.TRATTERIA_SERVICE_NAME {
		generationRules, err := s.tratteriaController.Controller.GetActiveGenerationRules(namespace)
		if err != nil {
			return nil, err
		}

		response.GenerationRules = generationRules
	} else {
		verificationRules, err := s.tratteriaController.Controller.GetActiveVerificationRules(serviceName, namespace)
		if err != nil {
			return nil, err
		}

		response.VerificationRules = verificationRules
	}

	s.dataPlaneRegistryManager.Register(ipaddress, port, serviceName, namespace)

	return response, nil
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

	tlsConfig := tlsconfig.MTLSClientConfig(s.x509Source, s.x509Source, tlsconfig.AuthorizeID(s.tratteriaSpiffeId))

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	for _, instance := range tratteriaInstances {
		url := fmt.Sprintf("https://%s/%s", instance.IpAddress, TRATTERIA_JWKS_ENDPOINT)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request for URL %s: %w", url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS from URL %s: %w", url, err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-ok status code %d from URL %s", resp.StatusCode, url)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body from URL %s: %w", url, err)
		}

		set, err := jwk.Parse(body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JWKS from URL %s: %w", url, err)
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
