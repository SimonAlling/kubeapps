package handler

import (
	"net/http"
	"strings"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
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
	configs := map[string]*action.Configuration{}
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
			namespace := vars["namespace"]
			config := configs[namespace]
			// If there is no existing config for the requested namespace, we'll create one:
			log.Info("")
			log.Info("HALLOJ WITH: this is config:")
			log.Info(config)
			if config == nil {
				log.Infof("Creating new config for namespace '%s' ...", namespace)
				token := extractToken(req.Header.Get("Authorization"))
				config = agent.NewConfig(token, namespace)
				configs[namespace] = config // Woah! Incredibly imperative!!!
			}
			h.HelmAgent.Config = config
			f(h, w, req, vars)
		}
	}
}

func ListReleases(h *Helmer, w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
	apps, err := h.HelmAgent.ListReleases(vars["namespace"], h.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func ListAllReleases(h *Helmer, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(h, w, req, map[string]string{"namespace": ""})
}
