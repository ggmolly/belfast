// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.25.3
// source: EVENT_PERFORMANCE.proto

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

type EVENT_PERFORMANCE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId *uint32 `protobuf:"varint,1,req,name=event_id,json=eventId" json:"event_id,omitempty"`
	Index   *uint32 `protobuf:"varint,2,req,name=index" json:"index,omitempty"`
}

func (x *EVENT_PERFORMANCE) Reset() {
	*x = EVENT_PERFORMANCE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_EVENT_PERFORMANCE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVENT_PERFORMANCE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVENT_PERFORMANCE) ProtoMessage() {}

func (x *EVENT_PERFORMANCE) ProtoReflect() protoreflect.Message {
	mi := &file_EVENT_PERFORMANCE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVENT_PERFORMANCE.ProtoReflect.Descriptor instead.
func (*EVENT_PERFORMANCE) Descriptor() ([]byte, []int) {
	return file_EVENT_PERFORMANCE_proto_rawDescGZIP(), []int{0}
}

func (x *EVENT_PERFORMANCE) GetEventId() uint32 {
	if x != nil && x.EventId != nil {
		return *x.EventId
	}
	return 0
}

func (x *EVENT_PERFORMANCE) GetIndex() uint32 {
	if x != nil && x.Index != nil {
		return *x.Index
	}
	return 0
}

var File_EVENT_PERFORMANCE_proto protoreflect.FileDescriptor

var file_EVENT_PERFORMANCE_proto_rawDesc = []byte{
	0x0a, 0x17, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x50, 0x45, 0x52, 0x46, 0x4f, 0x52, 0x4d, 0x41,
	0x4e, 0x43, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61,
	0x73, 0x74, 0x22, 0x44, 0x0a, 0x11, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x50, 0x45, 0x52, 0x46,
	0x4f, 0x52, 0x4d, 0x41, 0x4e, 0x43, 0x45, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_EVENT_PERFORMANCE_proto_rawDescOnce sync.Once
	file_EVENT_PERFORMANCE_proto_rawDescData = file_EVENT_PERFORMANCE_proto_rawDesc
)

func file_EVENT_PERFORMANCE_proto_rawDescGZIP() []byte {
	file_EVENT_PERFORMANCE_proto_rawDescOnce.Do(func() {
		file_EVENT_PERFORMANCE_proto_rawDescData = protoimpl.X.CompressGZIP(file_EVENT_PERFORMANCE_proto_rawDescData)
	})
	return file_EVENT_PERFORMANCE_proto_rawDescData
}

var file_EVENT_PERFORMANCE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_EVENT_PERFORMANCE_proto_goTypes = []interface{}{
	(*EVENT_PERFORMANCE)(nil), // 0: belfast.EVENT_PERFORMANCE
}
var file_EVENT_PERFORMANCE_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_EVENT_PERFORMANCE_proto_init() }
func file_EVENT_PERFORMANCE_proto_init() {
	if File_EVENT_PERFORMANCE_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_EVENT_PERFORMANCE_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EVENT_PERFORMANCE); i {
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
			RawDescriptor: file_EVENT_PERFORMANCE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_EVENT_PERFORMANCE_proto_goTypes,
		DependencyIndexes: file_EVENT_PERFORMANCE_proto_depIdxs,
		MessageInfos:      file_EVENT_PERFORMANCE_proto_msgTypes,
	}.Build()
	File_EVENT_PERFORMANCE_proto = out.File
	file_EVENT_PERFORMANCE_proto_rawDesc = nil
	file_EVENT_PERFORMANCE_proto_goTypes = nil
	file_EVENT_PERFORMANCE_proto_depIdxs = nil
}