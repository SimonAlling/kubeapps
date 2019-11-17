package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/helmer/internal/handler"
	theEpicAgentModule "github.com/kubeapps/kubeapps/pkg/agent"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"helm.sh/helm/v3/pkg/action"
)

func main() {
	fmt.Println("HALLOJ cmd/helmer/main.go")
	fmt.Printf("HALLOJ %v\n", new(action.Configuration))

	listLimit := 10 // TODO

	h := handler.Helmer{
		HelmAgent: theEpicAgentModule.NewHelmAgent(),
		ListLimit: listLimit,
	}

	r := mux.NewRouter()
	withHelmer := handler.With(&h)

	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases/").Handler(negroni.New(
		negroni.Wrap(withHelmer(handler.ListAllReleases)),
	))
	apiv1.Methods("GET").Path("/releases").Handler(negroni.New(
		negroni.Wrap(withHelmer(handler.ListAllReleases)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/").Handler(negroni.New(
		negroni.Wrap(withHelmer(handler.ListReleases)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		negroni.Wrap(withHelmer(handler.ListReleases)),
	))
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

	n := negroni.Classic()
	n.UseHandler(r)

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

	// Catch SIGINT and SIGTERM
	// Set up channel on which to send signal notifications.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	log.Debug("Set system to get notified on signals")
	s := <-c
	log.Infof("Received signal: %v. Waiting for existing requests to finish", s)
	// Set a timeout value high enough to let k8s terminationGracePeriodSeconds to act
	// accordingly and send a SIGKILL if needed
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Info("All requests have been served. Exiting")
	os.Exit(0)
}
