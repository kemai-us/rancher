package planner

import (
	"context"

	"github.com/rancher/channelserver/pkg/model"
	"github.com/rancher/norman/types/convert"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/rancher/pkg/channelserver"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2"
)

func filterConfigData(config map[string]interface{}, controlPlane *rkev1.RKEControlPlane, entry *planEntry) {
	var (
		isServer = isControlPlane(entry) || isEtcd(entry)
		release  = channelserver.GetReleaseConfigByRuntimeAndVersion(context.TODO(),
			rke2.GetRuntime(controlPlane.Spec.KubernetesVersion),
			controlPlane.Spec.KubernetesVersion)
	)

	for k, v := range config {
		if newV, ok := filterField(isServer, k, v, release); ok {
			config[k] = newV
		} else {
			delete(config, k)
		}
	}
}

func filterField(isServer bool, k string, v interface{}, release model.Release) (interface{}, bool) {
	if v == nil {
		return nil, false
	}

	field, fieldFound := release.AgentArgs[k]
	if !fieldFound && isServer {
		field, fieldFound = release.ServerArgs[k]
	}

	// can't find arg
	if !fieldFound {
		return nil, false
	}

	switch v.(type) {
	case string:
	case bool:
	case []interface{}:
	default:
		// unknown type
		return nil, false
	}

	if field.Type == "boolean" {
		return convert.ToBool(v), true
	}

	return v, true
}
