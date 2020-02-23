// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backyards

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/banzaicloud/backyards-cli/pkg/helm"
)

type AuthMode string

const (
	AnonymousAuthMode     AuthMode = "anonymous"
	ImpersonationAuthMode AuthMode = "impersonation"
)

type Values struct {
	NameOverride         string                      `json:"nameOverride,omitempty"`
	FullnameOverride     string                      `json:"fullnameOverride,omitempty"`
	ReplicaCount         int                         `json:"replicaCount,omitempty"`
	UseNamespaceResource bool                        `json:"useNamespaceResource,omitempty"`
	Resources            corev1.ResourceRequirements `json:"resources,omitempty"`
	UseIstioResources    bool                        `json:"useIstioResources,omitempty"`

	Ingress struct {
		Enabled     bool              `json:"enabled,omitempty"`
		Annotations map[string]string `json:"annotations,omitempty"`
		Paths       struct {
			Application string `json:"application,omitempty"`
			Web         string `json:"web,omitempty"`
		} `json:"paths,omitempty"`
		BasePath string   `json:"basePath,omitempty"`
		Hosts    []string `json:"hosts,omitempty"`
		TLS      []struct {
			SecretName string   `json:"secretName,omitempty"`
			Hosts      []string `json:"hosts,omitempty"`
		} `json:"tls,omitempty"`
	} `json:"ingress,omitempty"`

	Autoscaling struct {
		Enabled                           bool `json:"enabled,omitempty"`
		MinReplicas                       int  `json:"minReplicas,omitempty"`
		MaxReplicas                       int  `json:"maxReplicas,omitempty"`
		TargetCPUUtilizationPercentage    int  `json:"targetCPUUtilizationPercentage,omitempty"`
		TargetMemoryUtilizationPercentage int  `json:"targetMemoryUtilizationPercentage,omitempty"`
	} `json:"autoscaling,omitempty"`

	Application struct {
		helm.EnvironmentVariables
		Enabled bool       `json:"enabled,omitempty"`
		Image   helm.Image `json:"image,omitempty"`
		Service struct {
			Type string `json:"type,omitempty"`
			Port int    `json:"port,omitempty"`
		} `json:"service,omitempty"`
	} `json:"application,omitempty"`

	Web struct {
		helm.EnvironmentVariables
		Enabled   bool                        `json:"enabled,omitempty"`
		Image     helm.Image                  `json:"image,omitempty"`
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Service   struct {
			Type string `json:"type,omitempty"`
			Port int    `json:"port,omitempty"`
		} `json:"service,omitempty"`
	} `json:"web,omitempty"`

	Istio struct {
		Namespace          string `json:"namespace,omitempty"`
		CRName             string `json:"CRName,omitempty"`
		ServiceAccountName string `json:"serviceAccountName,omitempty"`
	} `json:"istio,omitempty"`

	Prometheus struct {
		Enabled     bool                        `json:"enabled,omitempty"`
		Image       helm.Image                  `json:"image,omitempty"`
		Resources   corev1.ResourceRequirements `json:"resources,omitempty"`
		ExternalURL string                      `json:"externalUrl,omitempty"`
		Config      struct {
			Global struct {
				ScrapeInterval     string `json:"scrapeInterval,omitempty"`
				ScrapeTimeout      string `json:"scrapeTimeout,omitempty"`
				EvaluationInterval string `json:"evaluationInterval,omitempty"`
			} `json:"global,omitempty"`
		} `json:"config,omitempty"`
		Service struct {
			Enabled bool   `json:"enabled,omitempty"`
			Type    string `json:"type,omitempty"`
			Port    int    `json:"port,omitempty"`
		} `json:"service,omitempty"`
		InMesh      bool   `json:"inMesh,omitempty"`
		ClusterName string `json:"clusterName,omitempty"`
	} `json:"prometheus,omitempty"`

	Grafana struct {
		Enabled   bool                        `json:"enabled,omitempty"`
		Image     helm.Image                  `json:"image,omitempty"`
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Security  struct {
			Enabled       bool   `json:"enabled,omitempty"`
			UsernameKey   string `json:"usernameKey,omitempty"`
			SecretName    string `json:"secretName,omitempty"`
			PassphraseKey string `json:"passphraseKey,omitempty"`
		} `json:"security,omitempty"`
		ExternalURL string   `json:"externalUrl,omitempty"`
		Plugins     []string `json:"plugins,omitempty"`
	} `json:"grafana,omitempty"`

	Tracing struct {
		Enabled     bool   `json:"enabled,omitempty"`
		ExternalURL string `json:"externalUrl,omitempty"`
		Provider    string `json:"provider,omitempty"`
		Jaeger      struct {
			Image     helm.Image                  `json:"image,omitempty"`
			Resources corev1.ResourceRequirements `json:"resources,omitempty"`
			Memory    struct {
				MaxTraces string `json:"max_traces,omitempty"`
			} `json:"memory,omitempty"`
			SpanStorageType  string `json:"spanStorageType,omitempty"`
			Persist          bool   `json:"persist,omitempty"`
			StorageClassName string `json:"storageClassName,omitempty"`
			AccessMode       string `json:"accessMode,omitempty"`
		} `json:"jaeger,omitempty"`
		Service struct {
			Annotations  map[string]string `json:"annotations,omitempty"`
			Name         string            `json:"name,omitempty"`
			Type         string            `json:"type,omitempty"`
			ExternalPort int               `json:"externalPort,omitempty"`
		} `json:"service,omitempty"`
		MTLS struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"mtls,omitempty"`
		MultiCluster struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"multiCluster,omitempty"`
	} `json:"tracing,omitempty"`

	IngressGateway struct {
		Enabled     bool `json:"enabled,omitempty"`
		MeshGateway struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"meshgateway,omitempty"`
		Service struct {
			Type string `json:"type,omitempty"`
		} `json:"service,omitempty"`
	} `json:"ingressgateway,omitempty"`

	AuditSink struct {
		Enabled     bool                        `json:"enabled,omitempty"`
		Image       helm.Image                  `json:"image,omitempty"`
		Resources   corev1.ResourceRequirements `json:"resources,omitempty"`
		Tolerations []corev1.Toleration         `json:"tolerations,omitempty"`
		Mode        string                      `json:"mode,omitempty"`
		HTTP        struct {
			Timeout        string `json:"timeout,omitempty"`
			RetryWaitMin   string `json:"retryWaitMin,omitempty"`
			RetryWaitMax   string `json:"retryWaitMax,omitempty"`
			RetryMax       int    `json:"retryMax,omitempty"`
			PanicOnFailure bool   `json:"panicOnFailure,omitempty"`
		} `json:"http,omitempty"`
	} `json:"auditsink,omitempty"`

	CertManager struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"certmanager,omitempty"`

	Auth struct {
		Mode AuthMode `json:"mode,omitempty"`
	} `json:"auth,omitempty"`

	Impersonation struct {
		Enabled bool `json:"enabled,omitempty"`
		Config  struct {
			Users           []string `json:"users,omitempty"`
			Groups          []string `json:"groups,omitempty"`
			ServiceAccounts []string `json:"serviceaccounts,omitempty"`
			Scopes          []string `json:"scopes,omitempty"`
		} `json:"config,omitempty"`
	} `json:"impersonation,omitempty"`

	ALS struct {
		Enabled   bool                        `json:"enabled,omitempty"`
		Image     helm.Image                  `json:"image,omitempty"`
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Service   struct {
			Type string `json:"type,omitempty"`
			Port int    `json:"port,omitempty"`
		} `json:"service,omitempty"`
		MTLS struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"mtls,omitempty"`
		MultiCluster struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"multiCluster,omitempty"`
	} `json:"als,omitempty"`

	KubeStateMetrics struct {
		Enabled   bool                        `json:"enabled,omitempty"`
		Replicas  int                         `json:"replicas,omitempty"`
		Image     helm.Image                  `json:"image,omitempty"`
		Resources corev1.ResourceRequirements `json:"resources,omitempty"`
		Ports     struct {
			Monitoring int `json:"monitoring,omitempty"`
			Telemetry  int `json:"telemetry,omitempty"`
		} `json:"ports,omitempty"`
		Service struct {
			MonitoringPort int `json:"monitoringPort,omitempty"`
			TelemetryPort  int `json:"telemetryPort,omitempty"`
		} `json:"service,omitempty"`
	} `json:"kubestatemetrics,omitempty"`

	TurbonomicImporter struct {
		Enabled    bool                        `json:"enabled,omitempty"`
		Image      helm.Image                  `json:"image"`
		Resources  corev1.ResourceRequirements `json:"resources,omitempty"`
		Turbonomic struct {
			Hostname           string `json:"hostname,omitempty"`
			Username           string `json:"username"`
			Password           string `json:"password"`
			InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
		} `json:"turbonomic"`
	} `json:"turbonomicimporter"`
}

func (values *Values) SetDefaults(releaseName, istioNamespace string) {
	values.NameOverride = releaseName
	values.UseNamespaceResource = true
	values.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}

	values.Ingress.Enabled = false

	values.Web.Enabled = true
	values.Web.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}

	values.Prometheus.Enabled = true
	values.Prometheus.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}
	values.Prometheus.ExternalURL = "/prometheus"
	values.Prometheus.Config.Global.ScrapeInterval = "10s" //nolint
	values.Prometheus.Config.Global.ScrapeTimeout = "10s"
	values.Prometheus.Config.Global.EvaluationInterval = "10s"
	values.Prometheus.InMesh = false

	values.Grafana.Enabled = true
	values.Grafana.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("800m"),
			corev1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}
	values.Grafana.ExternalURL = "/grafana"
	values.Grafana.Security.Enabled = false

	values.Tracing.Enabled = true
	values.Tracing.ExternalURL = "/jaeger"
	values.Tracing.Provider = "jaeger"
	values.Tracing.Jaeger.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}
	values.Tracing.Service.Name = "backyards-zipkin"

	values.Auth.Mode = AnonymousAuthMode
	values.Impersonation.Enabled = false

	values.KubeStateMetrics.Enabled = true

	values.TurbonomicImporter.Enabled = false

	values.TurbonomicImporter.Enabled = false
}
