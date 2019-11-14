package helm3agent

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	
)


type AppOverview struct {
	ReleaseName   string			 `json:"releaseName"`
	Version       string         `json:"version"`
	Namespace     string         `json:"namespace"`
	Icon          string         `json:"icon,omitempty"`
	Status        string         `json:"status"`
	Chart         string         `json:"chart"`
}

type Helm3AgentIf interface {
	ListAllReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error)	
}

type Helm3Agent struct {
	config *action.Configuration
}

func NewHelm3Agent(token string) *Helm3Agent {
	return &Helm3Agent{
		config: generateConfiguration(token, ""),
	}
}

// To be able to replace this function during unit test
var testGenerateConfig func(token, namespace string) *action.Configuration

// generateConfiguration generates a configuration from within the context of the pod
func generateConfiguration(token, namespace string) *action.Configuration {
	// Will only execute during unit test
	if testGenerateConfig != nil {
		return testGenerateConfig(token, namespace)
	}

	actionConfig := new(action.Configuration)

	config, err := rest.InClusterConfig()
	if err != nil {
			panic(err.Error())
	}
	config.BearerToken = token
	config.BearerTokenFile = ""

	clientset, err := kubernetes.NewForConfig(config)

	var store *storage.Storage
	switch os.Getenv("HELM_DRIVER") {
	case "secret", "secrets", "":
			d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
			d.Log = klog.Infof
			store = storage.Init(d)
	case "configmap", "configmaps":
			d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
			d.Log = klog.Infof
			store = storage.Init(d)
	case "memory":
			d := driver.NewMemory()
			store = storage.Init(d)
	default:
			// Not sure what to do here.
			panic("Unknown driver in HELM_DRIVER: " + os.Getenv("HELM_DRIVER"))
	}

	actionConfig.RESTClientGetter = nil
	actionConfig.KubeClient = kube.New(nil)
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof

	return actionConfig
}

// ListReleases lists the current releases in a given namespace
func (agent *Helm3Agent) ListReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error) {
	listCmd := action.NewList(agent.config)

	return runListCmd(listCmd, namespace)
}

// ListAllReleases lists the current releases in all namespaces
func (agent *Helm3Agent) ListAllReleases(namespace string, releaseListLimit int, status string) ([]AppOverview, error) {
	listCmd := action.NewList(agent.config)
	listCmd.AllNamespaces = true
	listCmd.Limit = releaseListLimit

	return runListCmd(listCmd, namespace)
}


func runListCmd(listCmd *action.List, namespace string) ([]AppOverview, error) {

	releases, err := listCmd.Run()

	if err != nil {
		return nil, err
	}

	apps := make([]AppOverview, 0, len(releases))

	for _, v := range releases {
		if  listCmd.AllNamespaces || namespace == v.Namespace {
			app := AppOverview{
				ReleaseName: v.Name,
				Version: "1.2.3",
				Icon: v.Chart.Metadata.Icon,
				Namespace: v.Namespace,
				Status: v.Info.Status.String(),
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}
