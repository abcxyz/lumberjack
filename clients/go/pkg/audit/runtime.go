// Copyright 2021 Lumberjack authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	api "github.com/abcxyz/lumberjack/clients/go/apis/v1alpha1"
)

// runtimeInfo is a processor that contains information about the application's
// runtime environment. This processor is inspired from:
// https://github.com/googleapis/google-cloud-go/blob/master/logging/resource.go.
type runtimeInfo struct {
	// monitoredResource is a *structpb.Value representation of
	// *mrpb.MonitoredResource. We use a *structpb.Value representation because we
	// store the runtime information in the `Metadata` field of the audit log
	// request payload.
	monitoredResource *structpb.Value
}

// newRuntimeInfo creates a new runtimeInfo processor. On initialization, this
// processor detects the runtime environment and stores it as a *structpb.Value.
// It's safe for unit tests to call this function because different runtime
// environments do not cause flakiness. In other words, detecting the runtime
// environment never returns an error, no matter what environment is. Only
// converting an *mrpb.MonitoredResource to *structpb.Value can return an error.
func newRuntimeInfo(ctx context.Context) (*runtimeInfo, error) {
	var mr *mrpb.MonitoredResource
	switch {
	// AppEngine, Functions, CloudRun, Kubernetes are detected first,
	// as metadata.OnGCE() erroneously returns true on these runtimes.
	case isAppEngine(ctx):
		mr = detectAppEngineResource(ctx)
	case isCloudFunction(ctx):
		mr = detectCloudFunction(ctx)
	case isCloudRun(ctx):
		mr = detectCloudRunResource(ctx)
	case isKubernetesEngine(ctx):
		mr = detectKubernetesResource(ctx)
	case metadata.OnGCE():
		mr = detectGCEResource(ctx)
	}
	if mr == nil {
		return &runtimeInfo{}, nil
	}
	val, err := toStructVal(mr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert *mrpb.MonitoredResource to *structpb.Value: %w", err)
	}
	return &runtimeInfo{val}, nil
}

// isAppEngine returns true for both standard and flex.
func isAppEngine(ctx context.Context) bool {
	_, service := os.LookupEnv("GAE_SERVICE")
	_, version := os.LookupEnv("GAE_VERSION")
	_, instance := os.LookupEnv("GAE_INSTANCE")

	return service && version && instance
}

func detectAppEngineResource(ctx context.Context) *mrpb.MonitoredResource {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil
	}
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil
	}

	return &mrpb.MonitoredResource{
		Type: "gae_app",
		Labels: map[string]string{
			"project_id":  projectID,
			"module_id":   os.Getenv("GAE_SERVICE"),
			"version_id":  os.Getenv("GAE_VERSION"),
			"instance_id": os.Getenv("GAE_INSTANCE"),
			"runtime":     os.Getenv("GAE_RUNTIME"),
			"zone":        zone,
		},
	}
}

func isCloudFunction(ctx context.Context) bool {
	// Reserved envvars in older function runtimes, e.g. Node.js 8, Python 3.7 and Go 1.11.
	_, name := os.LookupEnv("FUNCTION_NAME")
	_, region := os.LookupEnv("FUNCTION_REGION")
	_, entry := os.LookupEnv("ENTRY_POINT")

	// Reserved envvars in newer function runtimes.
	_, target := os.LookupEnv("FUNCTION_TARGET")
	_, signature := os.LookupEnv("FUNCTION_SIGNATURE_TYPE")
	_, service := os.LookupEnv("K_SERVICE")
	return (name && region && entry) || (target && signature && service)
}

func detectCloudFunction(ctx context.Context) *mrpb.MonitoredResource {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil
	}
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil
	}
	// Newer functions runtimes store name in K_SERVICE.
	functionName, exists := os.LookupEnv("K_SERVICE")
	if !exists {
		functionName, _ = os.LookupEnv("FUNCTION_NAME")
	}
	return &mrpb.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"project_id":    projectID,
			"region":        regionFromZone(zone),
			"function_name": functionName,
		},
	}
}

func isCloudRun(ctx context.Context) bool {
	_, config := os.LookupEnv("K_CONFIGURATION")
	_, service := os.LookupEnv("K_SERVICE")
	_, revision := os.LookupEnv("K_REVISION")
	return config && service && revision
}

func detectCloudRunResource(ctx context.Context) *mrpb.MonitoredResource {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil
	}
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil
	}
	return &mrpb.MonitoredResource{
		Type: "cloud_run_revision",
		Labels: map[string]string{
			"project_id":         projectID,
			"location":           regionFromZone(zone),
			"service_name":       os.Getenv("K_SERVICE"),
			"revision_name":      os.Getenv("K_REVISION"),
			"configuration_name": os.Getenv("K_CONFIGURATION"),
		},
	}
}

func isKubernetesEngine(ctx context.Context) bool {
	clusterName, err := metadata.InstanceAttributeValueWithContext(ctx, "cluster-name")
	// Note: InstanceAttributeValue can return "", nil
	if err != nil || clusterName == "" {
		return false
	}
	return true
}

func detectKubernetesResource(ctx context.Context) *mrpb.MonitoredResource {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil
	}
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil
	}
	clusterName, err := metadata.InstanceAttributeValueWithContext(ctx, "cluster-name")
	if err != nil {
		return nil
	}
	namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	namespaceName := ""
	if err == nil {
		namespaceName = string(namespaceBytes)
	}
	return &mrpb.MonitoredResource{
		Type: "k8s_container",
		Labels: map[string]string{
			"cluster_name":   clusterName,
			"location":       zone,
			"project_id":     projectID,
			"pod_name":       os.Getenv("HOSTNAME"),
			"namespace_name": namespaceName,
			// To get the `container_name` label, users need to explicitly provide it.
			"container_name": os.Getenv("CONTAINER_NAME"),
		},
	}
}

func detectGCEResource(ctx context.Context) *mrpb.MonitoredResource {
	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil
	}
	id, err := metadata.InstanceIDWithContext(ctx)
	if err != nil {
		return nil
	}
	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil
	}
	name, err := metadata.InstanceNameWithContext(ctx)
	if err != nil {
		return nil
	}
	return &mrpb.MonitoredResource{
		Type: "gce_instance",
		Labels: map[string]string{
			"project_id":    projectID,
			"instance_id":   id,
			"instance_name": name,
			"zone":          zone,
		},
	}
}

func regionFromZone(zone string) string {
	cutoff := strings.LastIndex(zone, "-")
	if cutoff > 0 {
		return zone[:cutoff]
	}
	return zone
}

func toStructVal(monitoredResource *mrpb.MonitoredResource) (*structpb.Value, error) {
	b, err := protojson.Marshal(monitoredResource)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %+v to JSON: %w", monitoredResource, err)
	}
	s := &structpb.Struct{}
	if err := protojson.Unmarshal(b, s); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	val := &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: s},
	}
	return val, nil
}

// Process stores the application's GCP runtime information in the audit log
// request. More specifically, in the Payload.Metadata under the key
// "originating_resource".
func (p *runtimeInfo) Process(ctx context.Context, logReq *api.AuditLogRequest) error {
	if p == nil || p.monitoredResource == nil {
		return nil
	}
	// Add monitored resource to Payload.Metadata as JSON.
	if logReq.GetPayload().GetMetadata() == nil {
		logReq.Payload.Metadata = &structpb.Struct{}
	}
	if logReq.GetPayload().GetMetadata().GetFields() == nil {
		logReq.Payload.Metadata.Fields = map[string]*structpb.Value{}
	}
	logReq.Payload.Metadata.Fields["originating_resource"] = p.monitoredResource
	return nil
}
