module github.com/webdevops/kube-pool-manager

go 1.15

require (
	github.com/jessevdk/go-flags v1.4.0
	github.com/operator-framework/operator-sdk v0.8.2
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.6.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2 // indirect
)
