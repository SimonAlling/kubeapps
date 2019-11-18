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

type dependentHandler func(h agent.Context, w http.ResponseWriter, req *http.Request, vars handlerutil.Params)

// A best effort at extracting the actual token from the Authorization header.
// We assume that the token is either preceded by tokenPrefix or not preceded by anything at all.
func extractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[tokenPrefixLength:]
	} else {
		return headerValue
	}
}

// Written in a curried fashion for convenient usage.
func WithContext(options agent.Options) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
			namespace := vars["namespace"]
			token := extractToken(req.Header.Get("Authorization"))
			h := agent.Context{
				AgentOptions: options,
				ActionConfig: agent.NewConfig(token, namespace),
			}
			f(h, w, req, vars)
		}
	}
}

func ListReleases(h agent.Context, w http.ResponseWriter, req *http.Request, vars handlerutil.Params) {
	apps, err := agent.ListReleases(h, vars["namespace"], req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func ListAllReleases(h agent.Context, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(h, w, req, map[string]string{"namespace": ""})
}
