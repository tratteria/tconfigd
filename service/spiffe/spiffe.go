package spiffe

import (
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func FetchSpiffeIdFromX509(x509Source *workloadapi.X509Source) (spiffeid.ID, error) {
	svid, err := x509Source.GetX509SVID()
	if err != nil {
		return spiffeid.ID{}, fmt.Errorf("failed to get X.509 SVID: %w", err)
	}

	return svid.ID, nil
}
