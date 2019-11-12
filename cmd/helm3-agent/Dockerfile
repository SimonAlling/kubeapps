FROM golang:1.13 as builder
COPY . /go/src/github.com/kubeapps/kubeapps
WORKDIR /go/src/github.com/kubeapps/kubeapps

ARG VERSION
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-X main.version=$VERSION" ./cmd/helm3-agent

FROM scratch
COPY --from=builder /go/src/github.com/kubeapps/kubeapps/helm3-agent /agent
EXPOSE 8080
CMD ["/agent"]
