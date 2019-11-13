package handler

import (
	"net/http"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
)

type Helmer struct {
	AgentClient agent.HelmClient
	ListLimit   int
}

func (h *Helmer) ListAllReleases(w http.ResponseWriter, req *http.Request) {
	apps, err := h.AgentClient.ListReleases("", h.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}
