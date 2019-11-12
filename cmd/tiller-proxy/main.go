/*
Copyright (c) 2018 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubeapps/kubeapps/cmd/cmd-backend"
	helm3Agent "github.com/kubeapps/kubeapps/pkg/helm3agent"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	appRepo "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	"github.com/kubeapps/kubeapps/cmd/tiller-proxy/internal/handler"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	tillerProxy "github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	helmChartUtil "k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/tlsutil"
)

var (
	settings    environment.EnvSettings
	proxy       *tillerProxy.Proxy
	kubeClient  kubernetes.Interface
	disableAuth bool
	listLimit   int
	timeout     int64

	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsVerify     bool   // enable TLS and verify remote certificates
	tlsEnable     bool   // enable TLS

	tlsCaCertDefault = fmt.Sprintf("%s/ca.crt", os.Getenv("HELM_HOME"))
	tlsCertDefault   = fmt.Sprintf("%s/tls.crt", os.Getenv("HELM_HOME"))
	tlsKeyDefault    = fmt.Sprintf("%s/tls.key", os.Getenv("HELM_HOME"))

	chartsvcURL string
)

const token = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt1YmVhcHBzLW9wZXJhdG9yLXRva2VuLWg3ZG1yIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Imt1YmVhcHBzLW9wZXJhdG9yIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMTIyMDgzZGYtNDRlMi00Y2QxLWJlMTQtMTY1NmQ4OWY1MWRlIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlZmF1bHQ6a3ViZWFwcHMtb3BlcmF0b3IifQ.q3qRFLc5mcGK632aYnCp0nghVzMngO25iA7t7wHni82EF65O-VkkP-O35-JSsonnUwY-9S5X4PEUiXK3ORo_RS4rMA7c6z8op3SR2CwRWYs91ODUKIYbAxVf9OrjDG0vz2AeeGPHtQW1BPj_byaGdGvL79venydRRY3ogqJAgJwRiLbC75Ghij1FbGx4SgUgHEs7_TJW_bf_-zB89N7DQlZWwuwdFzRUOsiQojJmfE2kNz1HpoH5Ae8DrAmySiLrwGQ5OL8vE-tXUm2LJQV9Hk7-zQxugyK33CZwc4U-aRaXJUNSzzzH8XXmD_JKbHiBJ6uV-2_h-0GDWXNc1Ysvpg"

/*
const token = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt1YmVhcHBzLW9wZXJhdG9yLXRva2VuLXY0d2pzIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Imt1YmVhcHBzLW9wZXJhdG9yIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiODZkODJjYzMtYWQzNC0xMWU5LTkzZDctZjJlOTk2YjQ3NGRlIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlZmF1bHQ6a3ViZWFwcHMtb3BlcmF0b3IifQ.V3YzqvLw-Os035zd3TrnJTMe3egvcbnhhbI0UP_gaqE6Ks_Fwhc5ZYNkds_VCXff_7dzJgzpgbh0ap-NuKVPb_dNcDrpClQ3s-C0jYeOzqdfzpwiRgMgUQaFfroDejtiE_cPrCYMpNG5YTowFzy6L22O6DNZ-VjDjD8fzGbFMfwD3ismIGajDxQ1_LYV898e-RGABk9QPN1GaqqU08VkJuu5bfYqj4wHlO6z6Ie2FLp6bFkR92cLiEMBxA-DWnH4ykB7glz4H0clb0gW3H0aR-0yN1TAr0tflds3J-MxCOHS4pwHCbAf0Hw1VeuFjcOwHNeCjy3QRMKB0tYP2c0RM-jllktkilQWIR7wDKFx6IaYxy1KUY8KxV4IaPHMtJN1lKYnFMSEe73EUyVtYBxgvrjnqnCkcDJZAG0cMgOr0H5SKpM6LI6mh_g5ihUvu7M7tAuUccBpnltzwwOLbE5p0Y_D0y6sSe6_hteRS4o9PCINK2ue7UeAUPAq3jQ7z_AdsLyDeFy2WfSx9aOoX35U_LrcovF5pkmflvILjDTJU8s7Cmoqftdb0ukB4bCIHBGyAb5SzqRvNCxvtZi_4n4ev2eF0K7DH1W9Fr7u2wGPbbCBtP4Eqid6kJDqGvCIliTOOxxtcrDhHTRsiZeQReDfa0qtstrVNqgyn4QTvQEtB7I"
*/

func init() {
	settings.AddFlags(pflag.CommandLine)
	// TLS Flags
	pflag.StringVar(&tlsCaCertFile, "tls-ca-cert", tlsCaCertDefault, "path to TLS CA certificate file")
	pflag.StringVar(&tlsCertFile, "tls-cert", tlsCertDefault, "path to TLS certificate file")
	pflag.StringVar(&tlsKeyFile, "tls-key", tlsKeyDefault, "path to TLS key file")
	pflag.BoolVar(&tlsVerify, "tls-verify", false, "enable TLS for request and verify remote")
	pflag.BoolVar(&tlsEnable, "tls", false, "enable TLS for request")
	pflag.BoolVar(&disableAuth, "disable-auth", false, "Disable authorization check")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	pflag.Int64Var(&timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
	pflag.StringVar(&chartsvcURL, "chartsvc-url", "http://kubeapps-internal-chartsvc:8080", "URL to the internal chartsvc")
}

func main() {
	pflag.Parse()

	// set defaults from environment
	settings.Init(pflag.CommandLine)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to get cluster config: %v", err)
	}

	kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create a kubernetes client: %v", err)
	}

	appRepoClient, err := appRepo.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create an app repository client: %v", err)
	}

	log.Printf("Using tiller host: %s", settings.TillerHost)
	helmOptions := []helm.Option{helm.Host(settings.TillerHost)}
	if tlsVerify || tlsEnable {
		if tlsCaCertFile == "" {
			tlsCaCertFile = settings.Home.TLSCaCert()
		}
		if tlsCertFile == "" {
			tlsCertFile = settings.Home.TLSCert()
		}
		if tlsKeyFile == "" {
			tlsKeyFile = settings.Home.TLSKey()
		}
		log.Printf("Using Key=%q, Cert=%q, CA=%q", tlsKeyFile, tlsCertFile, tlsCaCertFile)
		tlsopts := tlsutil.Options{KeyFile: tlsKeyFile, CertFile: tlsCertFile, InsecureSkipVerify: true}
		if tlsVerify {
			tlsopts.CaCertFile = tlsCaCertFile
			tlsopts.InsecureSkipVerify = false
		}
		tlscfg, err := tlsutil.ClientConfig(tlsopts)
		if err != nil {
			log.Fatal(err)
		}
		helmOptions = append(helmOptions, helm.WithTLS(tlscfg))
	}
	helmClient := helm.NewClient(helmOptions...)
	err = helmClient.PingTiller()
	if err != nil {
		log.Fatalf("Unable to connect to Tiller: %v", err)
	}

	proxy = tillerProxy.NewProxy(kubeClient, helmClient, timeout)
	chartutils := chartUtils.NewChart(kubeClient, appRepoClient, helmChartUtil.LoadArchive, userAgent())

	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	authGate := handler.AuthGate()

	// HTTP Handler
	h := handler.TillerProxy{
		DisableAuth: disableAuth,
		ListLimit:   listLimit,
		ChartClient: chartutils,
		ProxyClient: proxy,
	}

	helm3Agent := helm3Agent.Helm3AgentNew(token)
	cb := cmdbackend.Helm3AgentProxy{
		LogLimit: 1,
		Agent: helm3Agent,
	}

	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithoutParams(cb.ListAllReleases)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.ListReleases)),
	))
	apiv1.Methods("POST").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.CreateRelease)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.GetRelease)),
	))
	apiv1.Methods("PUT").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.OperateRelease)),
	))
	apiv1.Methods("DELETE").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.DeleteRelease)),
	))

	// Chartsvc reverse proxy
	parsedChartsvcURL, err := url.Parse(chartsvcURL)
	if err != nil {
		log.Fatalf("Unable to parse the chartsvc URL: %v", err)
	}
	chartsvcProxy := httputil.NewSingleHostReverseProxy(parsedChartsvcURL)
	chartsvcPrefix := "/chartsvc"
	chartsvcRouter := r.PathPrefix(chartsvcPrefix).Subrouter()
	// Logos don't require authentication so bypass that step
	chartsvcRouter.Methods("GET").Path("/v1/assets/{repo}/{id}/logo").Handler(negroni.New(
		negroni.Wrap(http.StripPrefix(chartsvcPrefix, chartsvcProxy)),
	))
	chartsvcRouter.Methods("GET").Handler(negroni.New(
		authGate,
		negroni.Wrap(http.StripPrefix(chartsvcPrefix, chartsvcProxy)),
	))

	n := negroni.Classic()
	n.UseHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: n,
	}

	go func() {
		log.WithFields(log.Fields{"addr": addr}).Info("Started Tiller Proxy")
		err = srv.ListenAndServe()
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
