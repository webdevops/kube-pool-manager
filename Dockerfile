FROM golang:1.17-alpine as build

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git

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
FROM gcr.io/distroless/static
ENV LOG_JSON=1 \
    LEASE_ENABLE=1
COPY --from=build /go/src/github.com/webdevops/kube-pool-manager/kube-pool-manager /
USER 1000:1000
ENTRYPOINT ["/kube-pool-manager"]
