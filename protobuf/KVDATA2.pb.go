// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: KVDATA2.proto

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

type KVDATA2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    *uint32 `protobuf:"varint,1,req,name=key" json:"key,omitempty"`
	Value1 *uint32 `protobuf:"varint,2,req,name=value1" json:"value1,omitempty"`
	Value2 *uint32 `protobuf:"varint,3,req,name=value2" json:"value2,omitempty"`
}

func (x *KVDATA2) Reset() {
	*x = KVDATA2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_KVDATA2_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVDATA2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVDATA2) ProtoMessage() {}

func (x *KVDATA2) ProtoReflect() protoreflect.Message {
	mi := &file_KVDATA2_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVDATA2.ProtoReflect.Descriptor instead.
func (*KVDATA2) Descriptor() ([]byte, []int) {
	return file_KVDATA2_proto_rawDescGZIP(), []int{0}
}

func (x *KVDATA2) GetKey() uint32 {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return 0
}

func (x *KVDATA2) GetValue1() uint32 {
	if x != nil && x.Value1 != nil {
		return *x.Value1
	}
	return 0
}

func (x *KVDATA2) GetValue2() uint32 {
	if x != nil && x.Value2 != nil {
		return *x.Value2
	}
	return 0
}

var File_KVDATA2_proto protoreflect.FileDescriptor

var file_KVDATA2_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x4b, 0x56, 0x44, 0x41, 0x54, 0x41, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x4b, 0x0a, 0x07, 0x4b, 0x56, 0x44, 0x41,
	0x54, 0x41, 0x32, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, 0x18,
	0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x31, 0x12, 0x16, 0x0a,
	0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x32, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_KVDATA2_proto_rawDescOnce sync.Once
	file_KVDATA2_proto_rawDescData = file_KVDATA2_proto_rawDesc
)

func file_KVDATA2_proto_rawDescGZIP() []byte {
	file_KVDATA2_proto_rawDescOnce.Do(func() {
		file_KVDATA2_proto_rawDescData = protoimpl.X.CompressGZIP(file_KVDATA2_proto_rawDescData)
	})
	return file_KVDATA2_proto_rawDescData
}

var file_KVDATA2_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_KVDATA2_proto_goTypes = []any{
	(*KVDATA2)(nil), // 0: belfast.KVDATA2
}
var file_KVDATA2_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_KVDATA2_proto_init() }
func file_KVDATA2_proto_init() {
	if File_KVDATA2_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_KVDATA2_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*KVDATA2); i {
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
			RawDescriptor: file_KVDATA2_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_KVDATA2_proto_goTypes,
		DependencyIndexes: file_KVDATA2_proto_depIdxs,
		MessageInfos:      file_KVDATA2_proto_msgTypes,
	}.Build()
	File_KVDATA2_proto = out.File
	file_KVDATA2_proto_rawDesc = nil
	file_KVDATA2_proto_goTypes = nil
	file_KVDATA2_proto_depIdxs = nil
}
