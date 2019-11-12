package helm3agentmockup

import(
	"strconv"
	log "github.com/sirupsen/logrus"
	"github.com/kubeapps/kubeapps/pkg/helm3agent"
)

type Helm3AgentMockup struct {
	Nothing int
}

func Helm3AgentMockupNew() (*Helm3AgentMockup) {
	log.Printf("Creating a Helm3 agent mockup")
	return &Helm3AgentMockup{5}
}

func (h *Helm3AgentMockup) ListAllReleases(namespace string, releaseListLimit int, status string) ([]helm3agent.AppOverview, error) {
	appOverview := make([]helm3agent.AppOverview, releaseListLimit)

	log.Printf("Mockup ListAllRelease, ReleaseListLimit = %d", releaseListLimit)
	for i := 0; i < releaseListLimit; i++ {
		appOverview[i].ReleaseName = "red-frog " + strconv.Itoa(i)
	}
	log.Printf("Mockup ListAllRelease done")
	return appOverview, nil
}