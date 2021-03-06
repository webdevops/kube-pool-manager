FROM golang:1.15 as build

WORKDIR /go/src/github.com/webdevops/kube-pool-manager

# Get deps (cached)
COPY ./go.mod /go/src/github.com/webdevops/kube-pool-manager
COPY ./go.sum /go/src/github.com/webdevops/kube-pool-manager
COPY ./Makefile /go/src/github.com/webdevops/kube-pool-manager
RUN make dependencies

# Compile
COPY ./ /go/src/github.com/webdevops/kube-pool-manager
RUN make test
RUN make lint
RUN make build
RUN ./kube-pool-manager --help

#############################################
# FINAL IMAGE
#############################################
FROM gcr.io/distroless/base
ENV LOG_JSON=1 \
    LEASE_ENABLE=1
COPY --from=build /go/src/github.com/webdevops/kube-pool-manager/kube-pool-manager /
USER 1000:1000
ENTRYPOINT ["/kube-pool-manager"]
