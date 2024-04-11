// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.25.3
// source: CHILD_PLAN_VAL.proto

package protobuf

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CHILD_PLAN_VAL struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlanId      *uint32 `protobuf:"varint,1,opt,name=plan_id,json=planId" json:"plan_id,omitempty"`
	EventId     *uint32 `protobuf:"varint,2,opt,name=event_id,json=eventId" json:"event_id,omitempty"`
	SpecEventId *uint32 `protobuf:"varint,3,opt,name=spec_event_id,json=specEventId" json:"spec_event_id,omitempty"`
}

func (x *CHILD_PLAN_VAL) Reset() {
	*x = CHILD_PLAN_VAL{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CHILD_PLAN_VAL_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CHILD_PLAN_VAL) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CHILD_PLAN_VAL) ProtoMessage() {}

func (x *CHILD_PLAN_VAL) ProtoReflect() protoreflect.Message {
	mi := &file_CHILD_PLAN_VAL_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CHILD_PLAN_VAL.ProtoReflect.Descriptor instead.
func (*CHILD_PLAN_VAL) Descriptor() ([]byte, []int) {
	return file_CHILD_PLAN_VAL_proto_rawDescGZIP(), []int{0}
}

func (x *CHILD_PLAN_VAL) GetPlanId() uint32 {
	if x != nil && x.PlanId != nil {
		return *x.PlanId
	}
	return 0
}

func (x *CHILD_PLAN_VAL) GetEventId() uint32 {
	if x != nil && x.EventId != nil {
		return *x.EventId
	}
	return 0
}

func (x *CHILD_PLAN_VAL) GetSpecEventId() uint32 {
	if x != nil && x.SpecEventId != nil {
		return *x.SpecEventId
	}
	return 0
}

var File_CHILD_PLAN_VAL_proto protoreflect.FileDescriptor

var file_CHILD_PLAN_VAL_proto_rawDesc = []byte{
	0x0a, 0x14, 0x43, 0x48, 0x49, 0x4c, 0x44, 0x5f, 0x50, 0x4c, 0x41, 0x4e, 0x5f, 0x56, 0x41, 0x4c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22,
	0x68, 0x0a, 0x0e, 0x43, 0x48, 0x49, 0x4c, 0x44, 0x5f, 0x50, 0x4c, 0x41, 0x4e, 0x5f, 0x56, 0x41,
	0x4c, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x06, 0x70, 0x6c, 0x61, 0x6e, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0d, 0x73, 0x70, 0x65, 0x63, 0x5f, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x73, 0x70,
	0x65, 0x63, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CHILD_PLAN_VAL_proto_rawDescOnce sync.Once
	file_CHILD_PLAN_VAL_proto_rawDescData = file_CHILD_PLAN_VAL_proto_rawDesc
)

func file_CHILD_PLAN_VAL_proto_rawDescGZIP() []byte {
	file_CHILD_PLAN_VAL_proto_rawDescOnce.Do(func() {
		file_CHILD_PLAN_VAL_proto_rawDescData = protoimpl.X.CompressGZIP(file_CHILD_PLAN_VAL_proto_rawDescData)
	})
	return file_CHILD_PLAN_VAL_proto_rawDescData
}

var file_CHILD_PLAN_VAL_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CHILD_PLAN_VAL_proto_goTypes = []interface{}{
	(*CHILD_PLAN_VAL)(nil), // 0: belfast.CHILD_PLAN_VAL
}
var file_CHILD_PLAN_VAL_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CHILD_PLAN_VAL_proto_init() }
func file_CHILD_PLAN_VAL_proto_init() {
	if File_CHILD_PLAN_VAL_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_CHILD_PLAN_VAL_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CHILD_PLAN_VAL); i {
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
			RawDescriptor: file_CHILD_PLAN_VAL_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CHILD_PLAN_VAL_proto_goTypes,
		DependencyIndexes: file_CHILD_PLAN_VAL_proto_depIdxs,
		MessageInfos:      file_CHILD_PLAN_VAL_proto_msgTypes,
	}.Build()
	File_CHILD_PLAN_VAL_proto = out.File
	file_CHILD_PLAN_VAL_proto_rawDesc = nil
	file_CHILD_PLAN_VAL_proto_goTypes = nil
	file_CHILD_PLAN_VAL_proto_depIdxs = nil
}