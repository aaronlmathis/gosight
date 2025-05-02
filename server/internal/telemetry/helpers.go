package telemetry

import "github.com/aaronlmathis/gosight/shared/model"

// TODO - do this better.
func MergeDimensionsWithMeta(base map[string]string, meta *model.Meta) map[string]string {
	out := make(map[string]string, len(base)+20)

	// Copy base dimensions first
	for k, v := range base {
		out[k] = v
	}

	// Helper to safely set label if not already present and not empty
	set := func(k, v string) {
		if v != "" && out[k] == "" {
			out[k] = v
		}
	}

	// System / agent / host
	set("agent_id", meta.AgentID)
	set("host_id", meta.HostID)
	set("endpoint_id", meta.EndpointID)
	set("hostname", meta.Hostname)
	set("ip_address", meta.IPAddress)
	set("os", meta.OS)
	set("platform", meta.Platform)
	set("arch", meta.Architecture)

	// Cloud
	set("cloud_provider", meta.CloudProvider)
	set("region", meta.Region)
	set("zone", meta.AvailabilityZone)
	set("instance_id", meta.InstanceID)
	set("instance_type", meta.InstanceType)
	set("account_id", meta.AccountID)
	set("project_id", meta.ProjectID)

	// Container / k8s
	set("container_id", meta.ContainerID)
	set("container_name", meta.ContainerName)
	set("pod_name", meta.PodName)
	set("namespace", meta.Namespace)
	set("cluster_name", meta.ClusterName)
	set("image_id", meta.ContainerImageID)
	set("image_name", meta.ContainerImageName)

	// App
	set("application", meta.Application)
	set("environment", meta.Environment)
	set("service", meta.Service)
	set("version", meta.Version)
	set("deployment_id", meta.DeploymentID)

	// Tags (custom metadata)
	for k, v := range meta.Tags {
		if v != "" && out[k] == "" {
			out[k] = v
		}
	}

	return out
}
