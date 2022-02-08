FROM golang:1.17 as build

WORKDIR /go/src/github.com/webdevops/kube-pool-manager

# Compile
COPY ./ /go/src/github.com/webdevops/kube-pool-manager
RUN make dependencies
RUN make test
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
