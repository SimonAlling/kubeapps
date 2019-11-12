package cmdbackend

import(
	"net/http"
	"github.com/kubeapps/common/response"
	log "github.com/sirupsen/logrus"
	"github.com/kubeapps/kubeapps/pkg/helm3agent"
)

type Helm3AgentProxy struct {
	LogLimit int
	Agent helm3agent.Helm3AgentIf
}

func (h *Helm3AgentProxy) ListAllReleases(writer http.ResponseWriter, req *http.Request) {

	app, error := h.Agent.ListAllReleases("", h.LogLimit, "")
	log.Printf("Nu har vi fått en release.")

	if error != nil {
		response.NewErrorResponse(http.StatusUnprocessableEntity, error.Error()).Write(writer)
		return
	} 
	response.NewDataResponse(app).Write(writer)
}

