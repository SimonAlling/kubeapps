package cmdbackend

import(
	"testing"
	"net/http/httptest"
	"github.com/kubeapps/kubeapps/pkg/helm3agent/helm3agentmockup"
)


func TestCmdBackend(t *testing.T) {

	helm3AgentMockup := helm3agentmockup.NewHelm3AgentMockup()

	if (helm3AgentMockup == nil) {
		t.Errorf("Could not create a helm3 agent mockup")
	}

	helm3Proxy := Helm3AgentProxy{
		LogLimit: 3,
		Agent: helm3AgentMockup,
	}

	req := httptest.NewRequest("GET", "http://foo.bar", nil)

	response := httptest.NewRecorder()

	expectedResponse := `{"data":[{"releaseName":"red-frog 0","version":"","namespace":"","status":"","chart":""},{"releaseName":"red-frog 1","version":"","namespace":"","status":"","chart":""},{"releaseName":"red-frog 2","version":"","namespace":"","status":"","chart":""}]}`
	helm3Proxy.ListAllReleases(response, req)

	if response.Body.String() == "" {
		t.Errorf("Received nothing at all")
	} else if response.Body.String() != expectedResponse {
		t.Errorf("Received %s\nExpected %s", response.Body, expectedResponse)
	} 
}