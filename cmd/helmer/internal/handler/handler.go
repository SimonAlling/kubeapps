package handler

import (
	"net/http"
	"strings"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	log "github.com/sirupsen/logrus"
)

const (
	tokenPrefix       = "Bearer "
	tokenPrefixLength = len(tokenPrefix)
)

type Helmer struct {
	HelmAgent agent.HelmAgent
	ListLimit int
}

type dependentHandler func(h *Helmer, w http.ResponseWriter, req *http.Request, vars handlerutil.Params)

// A best effort at extracting the actual token from the Authorization header.
// We assume that the token is either preceded by tokenPrefix or not preceded by anything at all.
func extractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[tokenPrefixLength:]
	} else {
		return headerValue
	}
}

func With(h *Helmer) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
			namespace := vars["namespace"]
			if h.HelmAgent.Config == nil {
				log.Info("HALLOJ With sketans config var nil, let's make a new one")
				token := extractToken(req.Header.Get("Authorization"))
				h.HelmAgent.Config = agent.NewConfig(token, namespace)
				log.Info("HALLOJ With config created")
			}
			f(h, w, req, vars)
		}
	}
}

func ListReleases(h *Helmer, w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
	log.Info("HALLOJ ListAllReleases funkar")
	apps, err := h.HelmAgent.ListReleases(vars["namespace"], h.ListLimit, req.URL.Query().Get("statuses"))
	log.Info("HALLOJ ListAllReleases vi klarade ListReleases")
	if err != nil {
		log.Info("HALLOJ err was non nil in ListAllReleases")
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	log.Info("HALLOJ ListReleases nu ska vi response hahah")
	response.NewDataResponse(apps).Write(w)
}

func ListAllReleases(h *Helmer, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(h, w, req, map[string]string{"namespace": ""})
}
