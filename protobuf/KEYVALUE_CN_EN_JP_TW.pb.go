// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: KEYVALUE_CN_EN_JP_TW.proto

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

type KEYVALUE_CN_EN_JP_TW struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    *uint32 `protobuf:"varint,1,req,name=key" json:"key,omitempty"`
	Value  *uint32 `protobuf:"varint,2,req,name=value" json:"value,omitempty"`
	Value2 *uint32 `protobuf:"varint,3,opt,name=value2" json:"value2,omitempty"`
}

func (x *KEYVALUE_CN_EN_JP_TW) Reset() {
	*x = KEYVALUE_CN_EN_JP_TW{}
	if protoimpl.UnsafeEnabled {
		mi := &file_KEYVALUE_CN_EN_JP_TW_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KEYVALUE_CN_EN_JP_TW) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KEYVALUE_CN_EN_JP_TW) ProtoMessage() {}

func (x *KEYVALUE_CN_EN_JP_TW) ProtoReflect() protoreflect.Message {
	mi := &file_KEYVALUE_CN_EN_JP_TW_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KEYVALUE_CN_EN_JP_TW.ProtoReflect.Descriptor instead.
func (*KEYVALUE_CN_EN_JP_TW) Descriptor() ([]byte, []int) {
	return file_KEYVALUE_CN_EN_JP_TW_proto_rawDescGZIP(), []int{0}
}

func (x *KEYVALUE_CN_EN_JP_TW) GetKey() uint32 {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return 0
}

func (x *KEYVALUE_CN_EN_JP_TW) GetValue() uint32 {
	if x != nil && x.Value != nil {
		return *x.Value
	}
	return 0
}

func (x *KEYVALUE_CN_EN_JP_TW) GetValue2() uint32 {
	if x != nil && x.Value2 != nil {
		return *x.Value2
	}
	return 0
}

var File_KEYVALUE_CN_EN_JP_TW_proto protoreflect.FileDescriptor

var file_KEYVALUE_CN_EN_JP_TW_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x4b, 0x45, 0x59, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x5f, 0x43, 0x4e, 0x5f, 0x45, 0x4e,
	0x5f, 0x4a, 0x50, 0x5f, 0x54, 0x57, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65,
	0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x56, 0x0a, 0x14, 0x4b, 0x45, 0x59, 0x56, 0x41, 0x4c, 0x55,
	0x45, 0x5f, 0x43, 0x4e, 0x5f, 0x45, 0x4e, 0x5f, 0x4a, 0x50, 0x5f, 0x54, 0x57, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_KEYVALUE_CN_EN_JP_TW_proto_rawDescOnce sync.Once
	file_KEYVALUE_CN_EN_JP_TW_proto_rawDescData = file_KEYVALUE_CN_EN_JP_TW_proto_rawDesc
)

func file_KEYVALUE_CN_EN_JP_TW_proto_rawDescGZIP() []byte {
	file_KEYVALUE_CN_EN_JP_TW_proto_rawDescOnce.Do(func() {
		file_KEYVALUE_CN_EN_JP_TW_proto_rawDescData = protoimpl.X.CompressGZIP(file_KEYVALUE_CN_EN_JP_TW_proto_rawDescData)
	})
	return file_KEYVALUE_CN_EN_JP_TW_proto_rawDescData
}

var file_KEYVALUE_CN_EN_JP_TW_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_KEYVALUE_CN_EN_JP_TW_proto_goTypes = []any{
	(*KEYVALUE_CN_EN_JP_TW)(nil), // 0: belfast.KEYVALUE_CN_EN_JP_TW
}
var file_KEYVALUE_CN_EN_JP_TW_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_KEYVALUE_CN_EN_JP_TW_proto_init() }
func file_KEYVALUE_CN_EN_JP_TW_proto_init() {
	if File_KEYVALUE_CN_EN_JP_TW_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_KEYVALUE_CN_EN_JP_TW_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*KEYVALUE_CN_EN_JP_TW); i {
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
			RawDescriptor: file_KEYVALUE_CN_EN_JP_TW_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_KEYVALUE_CN_EN_JP_TW_proto_goTypes,
		DependencyIndexes: file_KEYVALUE_CN_EN_JP_TW_proto_depIdxs,
		MessageInfos:      file_KEYVALUE_CN_EN_JP_TW_proto_msgTypes,
	}.Build()
	File_KEYVALUE_CN_EN_JP_TW_proto = out.File
	file_KEYVALUE_CN_EN_JP_TW_proto_rawDesc = nil
	file_KEYVALUE_CN_EN_JP_TW_proto_goTypes = nil
	file_KEYVALUE_CN_EN_JP_TW_proto_depIdxs = nil
}
