package tlscreds

import (
	"log"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type svidWatcher struct{}

func (s *svidWatcher) OnX509ContextWatchError(err error) {
	log.Printf("Error watching for updates: %v", err)
}

func (s *svidWatcher) OnX509ContextUpdate(c *workloadapi.X509Context) {
	pemCerts, pemKey, err := c.DefaultSVID().Marshal()
	if err != nil {
		log.Printf("Unable to marshal X.509 SVID: %v", err)

		return
	}

	if err := os.WriteFile(CertPath, pemCerts, certsFileMode); err != nil {
		log.Printf("Error writing certs file: %v", err)

		return
	}

	if err := os.WriteFile(KeyPath, pemKey, keyFileMode); err != nil {
		log.Printf("Error writing key file: %v", err)

		return
	}
}
