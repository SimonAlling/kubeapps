package helm3agentmockup

import(
	"strconv"
	"github.com/kubeapps/kubeapps/pkg/helm3agent"
)

type Helm3AgentMockup struct {
	Nothing int
}

func Helm3AgentMockupNew() (*Helm3AgentMockup) {
	return &Helm3AgentMockup{5}
}

func (h *Helm3AgentMockup) ListAllReleases(namespace string, releaseListLimit int, status string) ([]helm3agent.AppOverview, error) {
	appOverview := make([]helm3agent.AppOverview, releaseListLimit)

	for i := 0; i < releaseListLimit; i++ {
		appOverview[i].ReleaseName = "red-frog " + strconv.Itoa(i)
	}
	return appOverview, nil
}