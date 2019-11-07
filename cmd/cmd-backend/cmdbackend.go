package cmdbackend

import(
	"net/http"
	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/helm3agent"
)

type Helm3AgentProxy struct {
	LogLimit int
	Agent helm3agent.Helm3AgentIf
}

func (h *Helm3AgentProxy) ListAllReleases(writer http.ResponseWriter, req *http.Request) {

	app, error := h.Agent.ListAllReleases("", h.LogLimit, "")
	if error != nil {
		response.NewErrorResponse(http.StatusUnprocessableEntity, error.Error()).Write(writer)
	} else {
		response.NewDataResponse(app).Write(writer)
	}
}

