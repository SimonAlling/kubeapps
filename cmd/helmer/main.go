package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/helmer/internal/handler"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"helm.sh/helm/v3/pkg/action"
)

func main() {
	fmt.Println("HALLOJ cmd/helmer/main.go")
	fmt.Printf("HALLOJ %v\n", new(action.Configuration))

	listLimit := 10 // TODO

	r := mux.NewRouter()

	h := handler.Helmer{
		AgentClient: nil, // TODO
		ListLimit:   listLimit,
	}

	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").Handler(negroni.New(
		negroni.Wrap(handlerutil.WithoutParams(h.ListAllReleases)),
	))
	// apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
	// 	negroni.Wrap(handler.WithParams(h.ListReleases)),
	// ))
	// apiv1.Methods("POST").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
	// 	negroni.Wrap(handler.WithParams(h.CreateRelease)),
	// ))
	// apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
	// 	negroni.Wrap(handler.WithParams(h.GetRelease)),
	// ))
	// apiv1.Methods("PUT").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
	// 	negroni.Wrap(handler.WithParams(h.OperateRelease)),
	// ))
	// apiv1.Methods("DELETE").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
	// 	negroni.Wrap(handler.WithParams(h.DeleteRelease)),
	// ))

	log.Info("HALLOJ routes added")

	n := negroni.Classic()
	n.UseHandler(r)

	log.Info("HALLOJ handler being used by negroni")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Info("HALLOJ port and addr assigned")

	srv := &http.Server{
		Addr:    addr,
		Handler: n,
	}

	go func() {
		log.Info("Started Helmer")
		err := srv.ListenAndServe()
		if err != nil {
			log.Info(err)
		}
	}()
}
