// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: BUFF_INFO.proto

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

type BUFF_INFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        *uint32 `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Stack     *uint32 `protobuf:"varint,2,req,name=stack" json:"stack,omitempty"`
	Timestamp *uint32 `protobuf:"varint,3,req,name=timestamp" json:"timestamp,omitempty"`
	Round     *uint32 `protobuf:"varint,4,req,name=round" json:"round,omitempty"`
	Step      *uint32 `protobuf:"varint,5,req,name=step" json:"step,omitempty"`
}

func (x *BUFF_INFO) Reset() {
	*x = BUFF_INFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_BUFF_INFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BUFF_INFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BUFF_INFO) ProtoMessage() {}

func (x *BUFF_INFO) ProtoReflect() protoreflect.Message {
	mi := &file_BUFF_INFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BUFF_INFO.ProtoReflect.Descriptor instead.
func (*BUFF_INFO) Descriptor() ([]byte, []int) {
	return file_BUFF_INFO_proto_rawDescGZIP(), []int{0}
}

func (x *BUFF_INFO) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *BUFF_INFO) GetStack() uint32 {
	if x != nil && x.Stack != nil {
		return *x.Stack
	}
	return 0
}

func (x *BUFF_INFO) GetTimestamp() uint32 {
	if x != nil && x.Timestamp != nil {
		return *x.Timestamp
	}
	return 0
}

func (x *BUFF_INFO) GetRound() uint32 {
	if x != nil && x.Round != nil {
		return *x.Round
	}
	return 0
}

func (x *BUFF_INFO) GetStep() uint32 {
	if x != nil && x.Step != nil {
		return *x.Step
	}
	return 0
}

var File_BUFF_INFO_proto protoreflect.FileDescriptor

var file_BUFF_INFO_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x42, 0x55, 0x46, 0x46, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x79, 0x0a, 0x09, 0x42, 0x55,
	0x46, 0x46, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x12, 0x1c, 0x0a,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x72,
	0x6f, 0x75, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x74, 0x65, 0x70, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x04, 0x73, 0x74, 0x65, 0x70, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_BUFF_INFO_proto_rawDescOnce sync.Once
	file_BUFF_INFO_proto_rawDescData = file_BUFF_INFO_proto_rawDesc
)

func file_BUFF_INFO_proto_rawDescGZIP() []byte {
	file_BUFF_INFO_proto_rawDescOnce.Do(func() {
		file_BUFF_INFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_BUFF_INFO_proto_rawDescData)
	})
	return file_BUFF_INFO_proto_rawDescData
}

var file_BUFF_INFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_BUFF_INFO_proto_goTypes = []any{
	(*BUFF_INFO)(nil), // 0: belfast.BUFF_INFO
}
var file_BUFF_INFO_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_BUFF_INFO_proto_init() }
func file_BUFF_INFO_proto_init() {
	if File_BUFF_INFO_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_BUFF_INFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*BUFF_INFO); i {
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
			RawDescriptor: file_BUFF_INFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_BUFF_INFO_proto_goTypes,
		DependencyIndexes: file_BUFF_INFO_proto_depIdxs,
		MessageInfos:      file_BUFF_INFO_proto_msgTypes,
	}.Build()
	File_BUFF_INFO_proto = out.File
	file_BUFF_INFO_proto_rawDesc = nil
	file_BUFF_INFO_proto_goTypes = nil
	file_BUFF_INFO_proto_depIdxs = nil
}
