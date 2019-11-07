package helm3agent


type AppOverview struct {
	ReleaseName string;
}

type Helm3AgentIf interface {
	ListAllReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error)	
}