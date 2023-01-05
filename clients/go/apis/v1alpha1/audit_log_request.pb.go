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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: audit_log_request.proto

package v1alpha1

import (
	audit "google.golang.org/genproto/googleapis/cloud/audit"
	v2 "google.golang.org/genproto/googleapis/logging/v2"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The log type where this audit log entry goes. Our client converts
// the LogType enum to a Cloud Logging log name using the `log_name`
// option.
type AuditLogRequest_LogType int32

const (
	AuditLogRequest_UNSPECIFIED AuditLogRequest_LogType = 0
	// Administrative actions or changes to configuration through public APIs.
	AuditLogRequest_ADMIN_ACTIVITY AuditLogRequest_LogType = 1
	// Reads of configuration data and all access to user data through public
	// APIs.
	AuditLogRequest_DATA_ACCESS AuditLogRequest_LogType = 2
)

// Enum value maps for AuditLogRequest_LogType.
var (
	AuditLogRequest_LogType_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "ADMIN_ACTIVITY",
		2: "DATA_ACCESS",
	}
	AuditLogRequest_LogType_value = map[string]int32{
		"UNSPECIFIED":    0,
		"ADMIN_ACTIVITY": 1,
		"DATA_ACCESS":    2,
	}
)

func (x AuditLogRequest_LogType) Enum() *AuditLogRequest_LogType {
	p := new(AuditLogRequest_LogType)
	*p = x
	return p
}

func (x AuditLogRequest_LogType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuditLogRequest_LogType) Descriptor() protoreflect.EnumDescriptor {
	return file_audit_log_request_proto_enumTypes[0].Descriptor()
}

func (AuditLogRequest_LogType) Type() protoreflect.EnumType {
	return &file_audit_log_request_proto_enumTypes[0]
}

func (x AuditLogRequest_LogType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuditLogRequest_LogType.Descriptor instead.
func (AuditLogRequest_LogType) EnumDescriptor() ([]byte, []int) {
	return file_audit_log_request_proto_rawDescGZIP(), []int{0, 0}
}

// LogMode specifies the logging mode for the individual log request.
type AuditLogRequest_LogMode int32

const (
	// If unspecified, it's up to the audit client to decide what log
	// mode to use.
	AuditLogRequest_LOG_MODE_UNSPECIFIED AuditLogRequest_LogMode = 0
	// In FAIL_CLOSE mode, the log request must be persisted in the system
	// before return; in case of persistence failure, an error must be returned.
	AuditLogRequest_FAIL_CLOSE AuditLogRequest_LogMode = 1
	// In BEST_EFFORT mode, the log request will never return error; the
	// log request will be persisted with best effort.
	AuditLogRequest_BEST_EFFORT AuditLogRequest_LogMode = 2
)

// Enum value maps for AuditLogRequest_LogMode.
var (
	AuditLogRequest_LogMode_name = map[int32]string{
		0: "LOG_MODE_UNSPECIFIED",
		1: "FAIL_CLOSE",
		2: "BEST_EFFORT",
	}
	AuditLogRequest_LogMode_value = map[string]int32{
		"LOG_MODE_UNSPECIFIED": 0,
		"FAIL_CLOSE":           1,
		"BEST_EFFORT":          2,
	}
)

func (x AuditLogRequest_LogMode) Enum() *AuditLogRequest_LogMode {
	p := new(AuditLogRequest_LogMode)
	*p = x
	return p
}

func (x AuditLogRequest_LogMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuditLogRequest_LogMode) Descriptor() protoreflect.EnumDescriptor {
	return file_audit_log_request_proto_enumTypes[1].Descriptor()
}

func (AuditLogRequest_LogMode) Type() protoreflect.EnumType {
	return &file_audit_log_request_proto_enumTypes[1]
}

func (x AuditLogRequest_LogMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AuditLogRequest_LogMode.Descriptor instead.
func (AuditLogRequest_LogMode) EnumDescriptor() ([]byte, []int) {
	return file_audit_log_request_proto_rawDescGZIP(), []int{0, 1}
}

// Audit logging data pertaining to an operation, for use in-process.
//
// Our cloud logging client converts from this form to one or more
// google.logging.v2.LogEntry messages for transmission to Cloud Logging.
type AuditLogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type AuditLogRequest_LogType `protobuf:"varint,1,opt,name=type,proto3,enum=abcxyz.lumberjack.AuditLogRequest_LogType" json:"type,omitempty"`
	// The Cloud audit log payload.
	Payload *audit.AuditLog `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	// A map of key, value pairs that provides additional information about the
	// log entry. For example, an integration test can store a UUID in this field
	// to track a test log. Later, the integration test can query the UUID from a
	// BigQuery sink to ensure that a logging request completed successfully.
	Labels map[string]string       `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Mode   AuditLogRequest_LogMode `protobuf:"varint,4,opt,name=mode,proto3,enum=abcxyz.lumberjack.AuditLogRequest_LogMode" json:"mode,omitempty"`
	// Standard object used for correlating logs across streams
	Operation *v2.LogEntryOperation  `protobuf:"bytes,5,opt,name=operation,proto3" json:"operation,omitempty"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Contextual info associated with the log request.
	// Such info is meant for log processors and may not be directly
	// reflected in the final audit log.
	Context *structpb.Struct `protobuf:"bytes,7,opt,name=context,proto3" json:"context,omitempty"`
}

func (x *AuditLogRequest) Reset() {
	*x = AuditLogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_audit_log_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuditLogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuditLogRequest) ProtoMessage() {}

func (x *AuditLogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_audit_log_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuditLogRequest.ProtoReflect.Descriptor instead.
func (*AuditLogRequest) Descriptor() ([]byte, []int) {
	return file_audit_log_request_proto_rawDescGZIP(), []int{0}
}

func (x *AuditLogRequest) GetType() AuditLogRequest_LogType {
	if x != nil {
		return x.Type
	}
	return AuditLogRequest_UNSPECIFIED
}

func (x *AuditLogRequest) GetPayload() *audit.AuditLog {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *AuditLogRequest) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *AuditLogRequest) GetMode() AuditLogRequest_LogMode {
	if x != nil {
		return x.Mode
	}
	return AuditLogRequest_LOG_MODE_UNSPECIFIED
}

func (x *AuditLogRequest) GetOperation() *v2.LogEntryOperation {
	if x != nil {
		return x.Operation
	}
	return nil
}

func (x *AuditLogRequest) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *AuditLogRequest) GetContext() *structpb.Struct {
	if x != nil {
		return x.Context
	}
	return nil
}

var file_audit_log_request_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         390161750,
		Name:          "abcxyz.lumberjack.log_name",
		Tag:           "bytes,390161750,opt,name=log_name",
		Filename:      "audit_log_request.proto",
	},
}

// Extension fields to descriptorpb.EnumValueOptions.
var (
	// optional string log_name = 390161750;
	E_LogName = &file_audit_log_request_proto_extTypes[0]
)

var File_audit_log_request_proto protoreflect.FileDescriptor

var file_audit_log_request_proto_rawDesc = []byte{
	0x0a, 0x17, 0x61, 0x75, 0x64, 0x69, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x61, 0x62, 0x63, 0x78, 0x79,
	0x7a, 0x2e, 0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x6a, 0x61, 0x63, 0x6b, 0x1a, 0x22, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61, 0x75, 0x64, 0x69, 0x74,
	0x2f, 0x61, 0x75, 0x64, 0x69, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x21, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x67, 0x69, 0x6e, 0x67,
	0x2f, 0x76, 0x32, 0x2f, 0x6c, 0x6f, 0x67, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe2, 0x05, 0x0a, 0x0f, 0x41, 0x75, 0x64, 0x69, 0x74, 0x4c, 0x6f,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3e, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2e,
	0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x6a, 0x61, 0x63, 0x6b, 0x2e, 0x41, 0x75, 0x64, 0x69, 0x74,
	0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4c, 0x6f, 0x67, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x36, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x75, 0x64, 0x69, 0x74, 0x2e, 0x41,
	0x75, 0x64, 0x69, 0x74, 0x4c, 0x6f, 0x67, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x12, 0x46, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x2e, 0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2e, 0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x6a, 0x61, 0x63, 0x6b, 0x2e, 0x41, 0x75, 0x64, 0x69, 0x74, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x3e, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2e,
	0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x6a, 0x61, 0x63, 0x6b, 0x2e, 0x41, 0x75, 0x64, 0x69, 0x74,
	0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4c, 0x6f, 0x67, 0x4d, 0x6f,
	0x64, 0x65, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x42, 0x0a, 0x09, 0x6f, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x32, 0x2e,
	0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x38, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x31, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x1a, 0x39, 0x0a, 0x0b, 0x4c, 0x61, 0x62,
	0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x9c, 0x01, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x2f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x1a, 0x1e, 0xb2, 0xd5, 0xac, 0xd0, 0x0b, 0x18, 0x61, 0x75, 0x64, 0x69, 0x74, 0x2e, 0x61,
	0x62, 0x63, 0x78, 0x79, 0x7a, 0x2f, 0x75, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x12, 0x2f, 0x0a, 0x0e, 0x41, 0x44, 0x4d, 0x49, 0x4e, 0x5f, 0x41, 0x43, 0x54, 0x49, 0x56,
	0x49, 0x54, 0x59, 0x10, 0x01, 0x1a, 0x1b, 0xb2, 0xd5, 0xac, 0xd0, 0x0b, 0x15, 0x61, 0x75, 0x64,
	0x69, 0x74, 0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x69,
	0x74, 0x79, 0x12, 0x2f, 0x0a, 0x0b, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x41, 0x43, 0x43, 0x45, 0x53,
	0x53, 0x10, 0x02, 0x1a, 0x1e, 0xb2, 0xd5, 0xac, 0xd0, 0x0b, 0x18, 0x61, 0x75, 0x64, 0x69, 0x74,
	0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x61, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x22, 0x44, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x18,
	0x0a, 0x14, 0x4c, 0x4f, 0x47, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x46, 0x41, 0x49, 0x4c,
	0x5f, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x42, 0x45, 0x53, 0x54,
	0x5f, 0x45, 0x46, 0x46, 0x4f, 0x52, 0x54, 0x10, 0x02, 0x3a, 0x40, 0x0a, 0x08, 0x6c, 0x6f, 0x67,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd6, 0xca, 0x85, 0xba, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x6f, 0x0a, 0x1e, 0x63,
	0x6f, 0x6d, 0x2e, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2e, 0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x6a, 0x61, 0x63, 0x6b, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x14, 0x41,
	0x75, 0x64, 0x69, 0x74, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x61, 0x62, 0x63, 0x78, 0x79, 0x7a, 0x2f, 0x6c, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x6a,
	0x61, 0x63, 0x6b, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67, 0x6f, 0x2f, 0x61,
	0x70, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_audit_log_request_proto_rawDescOnce sync.Once
	file_audit_log_request_proto_rawDescData = file_audit_log_request_proto_rawDesc
)

func file_audit_log_request_proto_rawDescGZIP() []byte {
	file_audit_log_request_proto_rawDescOnce.Do(func() {
		file_audit_log_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_audit_log_request_proto_rawDescData)
	})
	return file_audit_log_request_proto_rawDescData
}

var file_audit_log_request_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_audit_log_request_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_audit_log_request_proto_goTypes = []interface{}{
	(AuditLogRequest_LogType)(0),          // 0: abcxyz.lumberjack.AuditLogRequest.LogType
	(AuditLogRequest_LogMode)(0),          // 1: abcxyz.lumberjack.AuditLogRequest.LogMode
	(*AuditLogRequest)(nil),               // 2: abcxyz.lumberjack.AuditLogRequest
	nil,                                   // 3: abcxyz.lumberjack.AuditLogRequest.LabelsEntry
	(*audit.AuditLog)(nil),                // 4: google.cloud.audit.AuditLog
	(*v2.LogEntryOperation)(nil),          // 5: google.logging.v2.LogEntryOperation
	(*timestamppb.Timestamp)(nil),         // 6: google.protobuf.Timestamp
	(*structpb.Struct)(nil),               // 7: google.protobuf.Struct
	(*descriptorpb.EnumValueOptions)(nil), // 8: google.protobuf.EnumValueOptions
}
var file_audit_log_request_proto_depIdxs = []int32{
	0, // 0: abcxyz.lumberjack.AuditLogRequest.type:type_name -> abcxyz.lumberjack.AuditLogRequest.LogType
	4, // 1: abcxyz.lumberjack.AuditLogRequest.payload:type_name -> google.cloud.audit.AuditLog
	3, // 2: abcxyz.lumberjack.AuditLogRequest.labels:type_name -> abcxyz.lumberjack.AuditLogRequest.LabelsEntry
	1, // 3: abcxyz.lumberjack.AuditLogRequest.mode:type_name -> abcxyz.lumberjack.AuditLogRequest.LogMode
	5, // 4: abcxyz.lumberjack.AuditLogRequest.operation:type_name -> google.logging.v2.LogEntryOperation
	6, // 5: abcxyz.lumberjack.AuditLogRequest.timestamp:type_name -> google.protobuf.Timestamp
	7, // 6: abcxyz.lumberjack.AuditLogRequest.context:type_name -> google.protobuf.Struct
	8, // 7: abcxyz.lumberjack.log_name:extendee -> google.protobuf.EnumValueOptions
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	7, // [7:8] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_audit_log_request_proto_init() }
func file_audit_log_request_proto_init() {
	if File_audit_log_request_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_audit_log_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuditLogRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_audit_log_request_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_audit_log_request_proto_goTypes,
		DependencyIndexes: file_audit_log_request_proto_depIdxs,
		EnumInfos:         file_audit_log_request_proto_enumTypes,
		MessageInfos:      file_audit_log_request_proto_msgTypes,
		ExtensionInfos:    file_audit_log_request_proto_extTypes,
	}.Build()
	File_audit_log_request_proto = out.File
	file_audit_log_request_proto_rawDesc = nil
	file_audit_log_request_proto_goTypes = nil
	file_audit_log_request_proto_depIdxs = nil
}
