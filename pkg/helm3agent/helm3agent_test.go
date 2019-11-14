package helm3agent

import(
	"io/ioutil"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/time"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/chart"
	"testing"
	"helm.sh/helm/v3/pkg/action"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
)

func testGenerateConfiguration(string, string) *action.Configuration {
	config := new(action.Configuration)
	config.Releases = storage.Init(driver.NewMemory())
	config.KubeClient = &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}}
	return config
}

func TestHelm3Agent(t *testing.T) {
	t.Logf("Running Test Helm 3 Agent")

	// Use the test version of generateConfiguration, restore afterwards
	testGenerateConfig = testGenerateConfiguration
	defer func() { testGenerateConfig = nil } ()

	// Create a helm3Agent with a faked KubeClient
	helm3Agent := NewHelm3Agent("Invalid token")

	if helm3Agent == nil {
		t.Errorf("Could not create a helm3Agent")
	}

	createNewRelease(helm3Agent.config.Releases, "crazy-frog", "default")
	createNewRelease(helm3Agent.config.Releases, "funny-bunny", "another namespace")
	createNewRelease(helm3Agent.config.Releases, "mad-magpie", "default")


	apps, err := helm3Agent.ListAllReleases("", 5, "")

	if err != nil {
		t.Errorf("Could not find any releases in any namespace")
	}

	if len(apps) != 3 {
		t.Errorf("Found %d releases, expected 3", len(apps))
	}

	namespace := "default"
	apps, err = helm3Agent.ListReleases(namespace, 5, "")
	
	if err != nil {
		t.Errorf("Could not find any releases in namespace " + namespace)
	}

	if len(apps) != 2 {
		t.Errorf("Found %d releases, expected 2", len(apps))
	}

}

func createNewRelease(store *storage.Storage, name, namespace string) {
	now := time.Now()
	// Just fill in the basics, and 
	rel := release.Release{
		Name : name,
		Namespace: namespace,
		Version: 1,
		Info : &release.Info{
			FirstDeployed: now,
			LastDeployed:  now,
			Status:        release.StatusDeployed,
			Description:   "Named test release",
		},
		Chart : &chart.Chart{
			Metadata : &chart.Metadata{
				Icon : "Icon " + name,
			},
		},
	}

	store.Create(&rel)
}