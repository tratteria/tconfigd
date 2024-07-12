package tlscreds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CertsDirPath     = "/etc/webhook/certs"
	CertPath         = "/etc/webhook/certs/tls.crt"
	KeyPath          = "/etc/webhook/certs/tls.key"
	certsFileMode    = os.FileMode(0o644)
	keyFileMode      = os.FileMode(0o600)
	certsDirMode     = os.FileMode(0o755)
	SpiffeWorkloadNewClientTimeout = 15 * time.Second
)

func SetupTLSCertAndKeyFromSPIRE() error {
	spireClient, err := GetSPIREWorkLoadApiClient()
	if err != nil {
		return fmt.Errorf("unable to create spire workload api client: %w", err)
	}

	if err := os.MkdirAll(CertsDirPath, certsDirMode); err != nil {
		return err
	}

	go func() {
		defer spireClient.Close()

		watcher := &svidWatcher{}

		err := spireClient.WatchX509Context(context.Background(), watcher)
		if err != nil && status.Code(err) != codes.Canceled {
			log.Fatalf("Error watching X.509 context: %v", err)
		}
	}()

	if err := WaitForCertificates(); err != nil {
		return err
	}

	return nil
}

func GetSPIREWorkLoadApiClient() (*workloadapi.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), SpiffeWorkloadNewClientTimeout)
	defer cancel()

	spireClient, err := workloadapi.New(ctx)
	if err != nil {
		return nil, err
	}

	return spireClient, nil
}

func WaitForCertificates() error {
	sleep := 500 * time.Millisecond
	maxRetries := 30

	for i := 1; i <= maxRetries; i++ {
		time.Sleep(sleep)

		if _, err := os.Stat(CertPath); err != nil {
			continue
		}

		if _, err := os.Stat(KeyPath); err != nil {
			continue
		}

		return nil
	}

	return errors.New("timed out waiting for trust bundle")
}
