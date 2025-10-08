module github.com/kart/k8s-agent/cluster-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/lib/pq v1.10.9
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v0.28.3
	k8s.io/metrics v0.28.3
)
