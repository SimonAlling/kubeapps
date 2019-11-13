package agent

import (
	"github.com/kubeapps/kubeapps/pkg/proxy"
)

type HelmClient interface {
	ListReleases(namespace string, releaseListLimit int, status string) ([]proxy.AppOverview, error)
}
