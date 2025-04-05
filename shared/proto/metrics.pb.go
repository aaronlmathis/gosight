// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.14.0
// source: metrics.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StatisticValues struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Minimum       float64                `protobuf:"fixed64,1,opt,name=minimum,proto3" json:"minimum,omitempty"`
	Maximum       float64                `protobuf:"fixed64,2,opt,name=maximum,proto3" json:"maximum,omitempty"`
	SampleCount   int32                  `protobuf:"varint,3,opt,name=sample_count,json=sampleCount,proto3" json:"sample_count,omitempty"`
	Sum           float64                `protobuf:"fixed64,4,opt,name=sum,proto3" json:"sum,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StatisticValues) Reset() {
	*x = StatisticValues{}
	mi := &file_metrics_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatisticValues) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatisticValues) ProtoMessage() {}

func (x *StatisticValues) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatisticValues.ProtoReflect.Descriptor instead.
func (*StatisticValues) Descriptor() ([]byte, []int) {
	return file_metrics_proto_rawDescGZIP(), []int{0}
}

func (x *StatisticValues) GetMinimum() float64 {
	if x != nil {
		return x.Minimum
	}
	return 0
}

func (x *StatisticValues) GetMaximum() float64 {
	if x != nil {
		return x.Maximum
	}
	return 0
}

func (x *StatisticValues) GetSampleCount() int32 {
	if x != nil {
		return x.SampleCount
	}
	return 0
}

func (x *StatisticValues) GetSum() float64 {
	if x != nil {
		return x.Sum
	}
	return 0
}

type Metric struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Namespace         string                 `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"` // e.g. "host", "container", "nginx"
	Subnamespace      string                 `protobuf:"bytes,2,opt,name=subnamespace,proto3" json:"subnamespace,omitempty"`
	Name              string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Timestamp         *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Value             float64                `protobuf:"fixed64,5,opt,name=value,proto3" json:"value,omitempty"`
	StatisticValues   *StatisticValues       `protobuf:"bytes,6,opt,name=statistic_values,json=statisticValues,proto3" json:"statistic_values,omitempty"`
	Unit              string                 `protobuf:"bytes,7,opt,name=unit,proto3" json:"unit,omitempty"`
	Dimensions        map[string]string      `protobuf:"bytes,8,rep,name=dimensions,proto3" json:"dimensions,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	StorageResolution int32                  `protobuf:"varint,9,opt,name=storage_resolution,json=storageResolution,proto3" json:"storage_resolution,omitempty"`
	Type              string                 `protobuf:"bytes,10,opt,name=type,proto3" json:"type,omitempty"` // e.g. "gauge", "counter"
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *Metric) Reset() {
	*x = Metric{}
	mi := &file_metrics_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_metrics_proto_rawDescGZIP(), []int{1}
}

func (x *Metric) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *Metric) GetSubnamespace() string {
	if x != nil {
		return x.Subnamespace
	}
	return ""
}

func (x *Metric) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Metric) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Metric) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *Metric) GetStatisticValues() *StatisticValues {
	if x != nil {
		return x.StatisticValues
	}
	return nil
}

func (x *Metric) GetUnit() string {
	if x != nil {
		return x.Unit
	}
	return ""
}

func (x *Metric) GetDimensions() map[string]string {
	if x != nil {
		return x.Dimensions
	}
	return nil
}

func (x *Metric) GetStorageResolution() int32 {
	if x != nil {
		return x.StorageResolution
	}
	return 0
}

func (x *Metric) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

type Meta struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// General Host Information
	Hostname      string `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	IpAddress     string `protobuf:"bytes,2,opt,name=ip_address,json=ipAddress,proto3" json:"ip_address,omitempty"`
	Os            string `protobuf:"bytes,3,opt,name=os,proto3" json:"os,omitempty"`
	OsVersion     string `protobuf:"bytes,4,opt,name=os_version,json=osVersion,proto3" json:"os_version,omitempty"`
	KernelVersion string `protobuf:"bytes,5,opt,name=kernel_version,json=kernelVersion,proto3" json:"kernel_version,omitempty"`
	Architecture  string `protobuf:"bytes,6,opt,name=architecture,proto3" json:"architecture,omitempty"`
	// Cloud Provider Specific
	CloudProvider    string `protobuf:"bytes,7,opt,name=cloud_provider,json=cloudProvider,proto3" json:"cloud_provider,omitempty"`
	Region           string `protobuf:"bytes,8,opt,name=region,proto3" json:"region,omitempty"`
	AvailabilityZone string `protobuf:"bytes,9,opt,name=availability_zone,json=availabilityZone,proto3" json:"availability_zone,omitempty"`
	InstanceId       string `protobuf:"bytes,10,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	InstanceType     string `protobuf:"bytes,11,opt,name=instance_type,json=instanceType,proto3" json:"instance_type,omitempty"`
	AccountId        string `protobuf:"bytes,12,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	ProjectId        string `protobuf:"bytes,13,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	ResourceGroup    string `protobuf:"bytes,14,opt,name=resource_group,json=resourceGroup,proto3" json:"resource_group,omitempty"`
	VpcId            string `protobuf:"bytes,15,opt,name=vpc_id,json=vpcId,proto3" json:"vpc_id,omitempty"`
	SubnetId         string `protobuf:"bytes,16,opt,name=subnet_id,json=subnetId,proto3" json:"subnet_id,omitempty"`
	ImageId          string `protobuf:"bytes,17,opt,name=image_id,json=imageId,proto3" json:"image_id,omitempty"`
	ServiceId        string `protobuf:"bytes,18,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	// Containerization/Orchestration
	ContainerId   string `protobuf:"bytes,19,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	ContainerName string `protobuf:"bytes,20,opt,name=container_name,json=containerName,proto3" json:"container_name,omitempty"`
	PodName       string `protobuf:"bytes,21,opt,name=pod_name,json=podName,proto3" json:"pod_name,omitempty"`
	Namespace     string `protobuf:"bytes,22,opt,name=namespace,proto3" json:"namespace,omitempty"` // K8s namespace
	ClusterName   string `protobuf:"bytes,23,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	NodeName      string `protobuf:"bytes,24,opt,name=node_name,json=nodeName,proto3" json:"node_name,omitempty"`
	// Application Specific
	Application  string `protobuf:"bytes,25,opt,name=application,proto3" json:"application,omitempty"`
	Environment  string `protobuf:"bytes,26,opt,name=environment,proto3" json:"environment,omitempty"`
	Service      string `protobuf:"bytes,27,opt,name=service,proto3" json:"service,omitempty"`
	Version      string `protobuf:"bytes,28,opt,name=version,proto3" json:"version,omitempty"`
	DeploymentId string `protobuf:"bytes,29,opt,name=deployment_id,json=deploymentId,proto3" json:"deployment_id,omitempty"`
	// Network Information
	PublicIp         string `protobuf:"bytes,30,opt,name=public_ip,json=publicIp,proto3" json:"public_ip,omitempty"`
	PrivateIp        string `protobuf:"bytes,31,opt,name=private_ip,json=privateIp,proto3" json:"private_ip,omitempty"`
	MacAddress       string `protobuf:"bytes,32,opt,name=mac_address,json=macAddress,proto3" json:"mac_address,omitempty"`
	NetworkInterface string `protobuf:"bytes,33,opt,name=network_interface,json=networkInterface,proto3" json:"network_interface,omitempty"`
	// Custom Metadata
	Tags          map[string]string `protobuf:"bytes,34,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Meta) Reset() {
	*x = Meta{}
	mi := &file_metrics_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Meta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Meta) ProtoMessage() {}

func (x *Meta) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Meta.ProtoReflect.Descriptor instead.
func (*Meta) Descriptor() ([]byte, []int) {
	return file_metrics_proto_rawDescGZIP(), []int{2}
}

func (x *Meta) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *Meta) GetIpAddress() string {
	if x != nil {
		return x.IpAddress
	}
	return ""
}

func (x *Meta) GetOs() string {
	if x != nil {
		return x.Os
	}
	return ""
}

func (x *Meta) GetOsVersion() string {
	if x != nil {
		return x.OsVersion
	}
	return ""
}

func (x *Meta) GetKernelVersion() string {
	if x != nil {
		return x.KernelVersion
	}
	return ""
}

func (x *Meta) GetArchitecture() string {
	if x != nil {
		return x.Architecture
	}
	return ""
}

func (x *Meta) GetCloudProvider() string {
	if x != nil {
		return x.CloudProvider
	}
	return ""
}

func (x *Meta) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *Meta) GetAvailabilityZone() string {
	if x != nil {
		return x.AvailabilityZone
	}
	return ""
}

func (x *Meta) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

func (x *Meta) GetInstanceType() string {
	if x != nil {
		return x.InstanceType
	}
	return ""
}

func (x *Meta) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *Meta) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *Meta) GetResourceGroup() string {
	if x != nil {
		return x.ResourceGroup
	}
	return ""
}

func (x *Meta) GetVpcId() string {
	if x != nil {
		return x.VpcId
	}
	return ""
}

func (x *Meta) GetSubnetId() string {
	if x != nil {
		return x.SubnetId
	}
	return ""
}

func (x *Meta) GetImageId() string {
	if x != nil {
		return x.ImageId
	}
	return ""
}

func (x *Meta) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *Meta) GetContainerId() string {
	if x != nil {
		return x.ContainerId
	}
	return ""
}

func (x *Meta) GetContainerName() string {
	if x != nil {
		return x.ContainerName
	}
	return ""
}

func (x *Meta) GetPodName() string {
	if x != nil {
		return x.PodName
	}
	return ""
}

func (x *Meta) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *Meta) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *Meta) GetNodeName() string {
	if x != nil {
		return x.NodeName
	}
	return ""
}

func (x *Meta) GetApplication() string {
	if x != nil {
		return x.Application
	}
	return ""
}

func (x *Meta) GetEnvironment() string {
	if x != nil {
		return x.Environment
	}
	return ""
}

func (x *Meta) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *Meta) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Meta) GetDeploymentId() string {
	if x != nil {
		return x.DeploymentId
	}
	return ""
}

func (x *Meta) GetPublicIp() string {
	if x != nil {
		return x.PublicIp
	}
	return ""
}

func (x *Meta) GetPrivateIp() string {
	if x != nil {
		return x.PrivateIp
	}
	return ""
}

func (x *Meta) GetMacAddress() string {
	if x != nil {
		return x.MacAddress
	}
	return ""
}

func (x *Meta) GetNetworkInterface() string {
	if x != nil {
		return x.NetworkInterface
	}
	return ""
}

func (x *Meta) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

type MetricPayload struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Host          string                 `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Timestamp     *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Metrics       []*Metric              `protobuf:"bytes,3,rep,name=metrics,proto3" json:"metrics,omitempty"`
	Meta          *Meta                  `protobuf:"bytes,4,opt,name=meta,proto3" json:"meta,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MetricPayload) Reset() {
	*x = MetricPayload{}
	mi := &file_metrics_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MetricPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricPayload) ProtoMessage() {}

func (x *MetricPayload) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricPayload.ProtoReflect.Descriptor instead.
func (*MetricPayload) Descriptor() ([]byte, []int) {
	return file_metrics_proto_rawDescGZIP(), []int{3}
}

func (x *MetricPayload) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *MetricPayload) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *MetricPayload) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

func (x *MetricPayload) GetMeta() *Meta {
	if x != nil {
		return x.Meta
	}
	return nil
}

type MetricResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	StatusCode    int32                  `protobuf:"varint,2,opt,name=status_code,json=statusCode,proto3" json:"status_code,omitempty"` // 0=OK, 1=validation error, etc.
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MetricResponse) Reset() {
	*x = MetricResponse{}
	mi := &file_metrics_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricResponse) ProtoMessage() {}

func (x *MetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricResponse.ProtoReflect.Descriptor instead.
func (*MetricResponse) Descriptor() ([]byte, []int) {
	return file_metrics_proto_rawDescGZIP(), []int{4}
}

func (x *MetricResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *MetricResponse) GetStatusCode() int32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

var File_metrics_proto protoreflect.FileDescriptor

const file_metrics_proto_rawDesc = "" +
	"\n" +
	"\rmetrics.proto\x12\x05proto\x1a\x1fgoogle/protobuf/timestamp.proto\"z\n" +
	"\x0fStatisticValues\x12\x18\n" +
	"\aminimum\x18\x01 \x01(\x01R\aminimum\x12\x18\n" +
	"\amaximum\x18\x02 \x01(\x01R\amaximum\x12!\n" +
	"\fsample_count\x18\x03 \x01(\x05R\vsampleCount\x12\x10\n" +
	"\x03sum\x18\x04 \x01(\x01R\x03sum\"\xc6\x03\n" +
	"\x06Metric\x12\x1c\n" +
	"\tnamespace\x18\x01 \x01(\tR\tnamespace\x12\"\n" +
	"\fsubnamespace\x18\x02 \x01(\tR\fsubnamespace\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x128\n" +
	"\ttimestamp\x18\x04 \x01(\v2\x1a.google.protobuf.TimestampR\ttimestamp\x12\x14\n" +
	"\x05value\x18\x05 \x01(\x01R\x05value\x12A\n" +
	"\x10statistic_values\x18\x06 \x01(\v2\x16.proto.StatisticValuesR\x0fstatisticValues\x12\x12\n" +
	"\x04unit\x18\a \x01(\tR\x04unit\x12=\n" +
	"\n" +
	"dimensions\x18\b \x03(\v2\x1d.proto.Metric.DimensionsEntryR\n" +
	"dimensions\x12-\n" +
	"\x12storage_resolution\x18\t \x01(\x05R\x11storageResolution\x12\x12\n" +
	"\x04type\x18\n" +
	" \x01(\tR\x04type\x1a=\n" +
	"\x0fDimensionsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x8e\t\n" +
	"\x04Meta\x12\x1a\n" +
	"\bhostname\x18\x01 \x01(\tR\bhostname\x12\x1d\n" +
	"\n" +
	"ip_address\x18\x02 \x01(\tR\tipAddress\x12\x0e\n" +
	"\x02os\x18\x03 \x01(\tR\x02os\x12\x1d\n" +
	"\n" +
	"os_version\x18\x04 \x01(\tR\tosVersion\x12%\n" +
	"\x0ekernel_version\x18\x05 \x01(\tR\rkernelVersion\x12\"\n" +
	"\farchitecture\x18\x06 \x01(\tR\farchitecture\x12%\n" +
	"\x0ecloud_provider\x18\a \x01(\tR\rcloudProvider\x12\x16\n" +
	"\x06region\x18\b \x01(\tR\x06region\x12+\n" +
	"\x11availability_zone\x18\t \x01(\tR\x10availabilityZone\x12\x1f\n" +
	"\vinstance_id\x18\n" +
	" \x01(\tR\n" +
	"instanceId\x12#\n" +
	"\rinstance_type\x18\v \x01(\tR\finstanceType\x12\x1d\n" +
	"\n" +
	"account_id\x18\f \x01(\tR\taccountId\x12\x1d\n" +
	"\n" +
	"project_id\x18\r \x01(\tR\tprojectId\x12%\n" +
	"\x0eresource_group\x18\x0e \x01(\tR\rresourceGroup\x12\x15\n" +
	"\x06vpc_id\x18\x0f \x01(\tR\x05vpcId\x12\x1b\n" +
	"\tsubnet_id\x18\x10 \x01(\tR\bsubnetId\x12\x19\n" +
	"\bimage_id\x18\x11 \x01(\tR\aimageId\x12\x1d\n" +
	"\n" +
	"service_id\x18\x12 \x01(\tR\tserviceId\x12!\n" +
	"\fcontainer_id\x18\x13 \x01(\tR\vcontainerId\x12%\n" +
	"\x0econtainer_name\x18\x14 \x01(\tR\rcontainerName\x12\x19\n" +
	"\bpod_name\x18\x15 \x01(\tR\apodName\x12\x1c\n" +
	"\tnamespace\x18\x16 \x01(\tR\tnamespace\x12!\n" +
	"\fcluster_name\x18\x17 \x01(\tR\vclusterName\x12\x1b\n" +
	"\tnode_name\x18\x18 \x01(\tR\bnodeName\x12 \n" +
	"\vapplication\x18\x19 \x01(\tR\vapplication\x12 \n" +
	"\venvironment\x18\x1a \x01(\tR\venvironment\x12\x18\n" +
	"\aservice\x18\x1b \x01(\tR\aservice\x12\x18\n" +
	"\aversion\x18\x1c \x01(\tR\aversion\x12#\n" +
	"\rdeployment_id\x18\x1d \x01(\tR\fdeploymentId\x12\x1b\n" +
	"\tpublic_ip\x18\x1e \x01(\tR\bpublicIp\x12\x1d\n" +
	"\n" +
	"private_ip\x18\x1f \x01(\tR\tprivateIp\x12\x1f\n" +
	"\vmac_address\x18  \x01(\tR\n" +
	"macAddress\x12+\n" +
	"\x11network_interface\x18! \x01(\tR\x10networkInterface\x12)\n" +
	"\x04tags\x18\" \x03(\v2\x15.proto.Meta.TagsEntryR\x04tags\x1a7\n" +
	"\tTagsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\xa7\x01\n" +
	"\rMetricPayload\x12\x12\n" +
	"\x04host\x18\x01 \x01(\tR\x04host\x128\n" +
	"\ttimestamp\x18\x02 \x01(\v2\x1a.google.protobuf.TimestampR\ttimestamp\x12'\n" +
	"\ametrics\x18\x03 \x03(\v2\r.proto.MetricR\ametrics\x12\x1f\n" +
	"\x04meta\x18\x04 \x01(\v2\v.proto.MetaR\x04meta\"I\n" +
	"\x0eMetricResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status\x12\x1f\n" +
	"\vstatus_code\x18\x02 \x01(\x05R\n" +
	"statusCode2\x8d\x01\n" +
	"\x0eMetricsService\x12<\n" +
	"\rSubmitMetrics\x12\x14.proto.MetricPayload\x1a\x15.proto.MetricResponse\x12=\n" +
	"\fSubmitStream\x12\x14.proto.MetricPayload\x1a\x15.proto.MetricResponse(\x01B.Z,github.com/aaronlmathis/gosight/shared/protob\x06proto3"

var (
	file_metrics_proto_rawDescOnce sync.Once
	file_metrics_proto_rawDescData []byte
)

func file_metrics_proto_rawDescGZIP() []byte {
	file_metrics_proto_rawDescOnce.Do(func() {
		file_metrics_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_metrics_proto_rawDesc), len(file_metrics_proto_rawDesc)))
	})
	return file_metrics_proto_rawDescData
}

var file_metrics_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_metrics_proto_goTypes = []any{
	(*StatisticValues)(nil),       // 0: proto.StatisticValues
	(*Metric)(nil),                // 1: proto.Metric
	(*Meta)(nil),                  // 2: proto.Meta
	(*MetricPayload)(nil),         // 3: proto.MetricPayload
	(*MetricResponse)(nil),        // 4: proto.MetricResponse
	nil,                           // 5: proto.Metric.DimensionsEntry
	nil,                           // 6: proto.Meta.TagsEntry
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
}
var file_metrics_proto_depIdxs = []int32{
	7, // 0: proto.Metric.timestamp:type_name -> google.protobuf.Timestamp
	0, // 1: proto.Metric.statistic_values:type_name -> proto.StatisticValues
	5, // 2: proto.Metric.dimensions:type_name -> proto.Metric.DimensionsEntry
	6, // 3: proto.Meta.tags:type_name -> proto.Meta.TagsEntry
	7, // 4: proto.MetricPayload.timestamp:type_name -> google.protobuf.Timestamp
	1, // 5: proto.MetricPayload.metrics:type_name -> proto.Metric
	2, // 6: proto.MetricPayload.meta:type_name -> proto.Meta
	3, // 7: proto.MetricsService.SubmitMetrics:input_type -> proto.MetricPayload
	3, // 8: proto.MetricsService.SubmitStream:input_type -> proto.MetricPayload
	4, // 9: proto.MetricsService.SubmitMetrics:output_type -> proto.MetricResponse
	4, // 10: proto.MetricsService.SubmitStream:output_type -> proto.MetricResponse
	9, // [9:11] is the sub-list for method output_type
	7, // [7:9] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_metrics_proto_init() }
func file_metrics_proto_init() {
	if File_metrics_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_metrics_proto_rawDesc), len(file_metrics_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_metrics_proto_goTypes,
		DependencyIndexes: file_metrics_proto_depIdxs,
		MessageInfos:      file_metrics_proto_msgTypes,
	}.Build()
	File_metrics_proto = out.File
	file_metrics_proto_goTypes = nil
	file_metrics_proto_depIdxs = nil
}
