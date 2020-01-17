package ale

import (
	"regexp"
	"strings"
)

var ownerMatchRegexp = regexp.MustCompile(`^kubernetes://.*/namespaces/([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)/([a-z0-9]+(?:-[a-z0-9]+)*)$`)

func (re *Reporter) IsIdentifiable() {}

func (re *Reporter) SetAttributes(attrs map[string]interface{}) {
	if value, ok := attrs["NAME"].(string); ok {
		re.Name = value
	}
	if value, ok := attrs["NAMESPACE"].(string); ok {
		re.Namespace = value
	}
	if value, ok := attrs["WORKLOAD_NAME"].(string); ok {
		re.Workload = value
	}
	if value, ok := attrs["SERVICE_ACCOUNT"].(string); ok {
		re.ServiceAccount = value
	}
	if value, ok := attrs["CLUSTER_ID"].(string); ok {
		re.ClusterID = value
	}
	if value, ok := attrs["INSTANCE_IPS"].(string); ok {
		re.InstanceIPs = strings.Split(value, ",")
	}
	if value, ok := attrs["OWNER"].(string); ok {
		m := ownerMatchRegexp.FindStringSubmatch(value)
		if len(m) == 4 {
			re.Owner = &Owner{
				Raw:       value,
				Type:      m[2],
				Name:      m[3],
				Namespace: m[1],
			}
		}
	}
	if value, ok := attrs["ISTIO_VERSION"].(string); ok {
		re.IstioVersion = value
	}
	if value, ok := attrs["MESH_ID"].(string); ok {
		re.MeshID = value
	}

	if re.Metadata == nil {
		re.Metadata = make(map[string]string)
	}

	if labels, ok := attrs["LABELS"].(map[string]interface{}); ok {
		for k, v := range labels {
			if value, ok := v.(string); ok {
				re.Metadata[k] = value
			}
		}
	}

	for k, v := range attrs {
		if value, ok := v.(string); ok {
			re.Metadata[k] = value
		}
	}
}
