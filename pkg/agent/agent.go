package agent

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kubeapps/kubeapps/pkg/proxy"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const driverEnvVar = "HELM_DRIVER"

type HelmAgent struct {
	Config *action.Configuration
}

func NewHelmAgent() HelmAgent {
	return HelmAgent{}
}

func (h *HelmAgent) ListReleases(namespace string, releaseListLimit int, status string) ([]proxy.AppOverview, error) {
	cmd := action.NewList(h.Config)
	klog.Info("HALLOJ ListReleases namespace is " + namespace)
	if namespace == "" {
		cmd.AllNamespaces = true
	}
	cmd.Limit = releaseListLimit
	releases, err := cmd.Run()
	if err != nil {
		return nil, err
	}
	klog.Info("HALLOJ len is " + strconv.Itoa(len(releases)))
	appOverviews := make([]proxy.AppOverview, len(releases))
	for i, r := range releases {
		appOverviews[i] = appOverviewFromRelease(r)
	}
	return appOverviews, nil
}

func appOverviewFromRelease(r *release.Release) proxy.AppOverview {
	return proxy.AppOverview{
		ReleaseName: r.Name,
		Version:     strconv.Itoa(r.Version),
		Icon:        r.Chart.Metadata.Icon,
		Namespace:   r.Namespace,
		Status:      r.Info.Status.String(),
	}
}

func NewConfig(token, namespace string) *action.Configuration {
	klog.Info("HALLOJ NewConfig")
	actionConfig := new(action.Configuration)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	klog.Info("HALLOJ NewConfig har config")
	klog.Info("HALLOJ env var value is " + os.Getenv(driverEnvVar))
	config.BearerToken = token
	config.BearerTokenFile = ""
	clientset, err := kubernetes.NewForConfig(config)
	store := createStorage(os.Getenv(driverEnvVar), namespace, clientset)
	klog.Info("HALLOJ NewConfig store created")

	actionConfig.RESTClientGetter = nil
	actionConfig.KubeClient = kube.New(nil)
	actionConfig.Releases = store
	actionConfig.Log = klog.Infof
	return actionConfig
}

func createStorage(driverType, namespace string, clientset *kubernetes.Clientset) *storage.Storage {
	klog.Info("HALLOJ createStorage")
	var store *storage.Storage
	switch driverType {
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
		// Wat do???
		panic(fmt.Sprintf("Unknown value of environment variable %s: %s", driverEnvVar, driverType))
	}
	return store
}
