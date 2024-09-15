// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_22010.proto

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

type SC_22010 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result    *uint32 `protobuf:"varint,1,req,name=result" json:"result,omitempty"`
	ExpInWell *uint32 `protobuf:"varint,2,req,name=exp_in_well,json=expInWell" json:"exp_in_well,omitempty"`
}

func (x *SC_22010) Reset() {
	*x = SC_22010{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_22010_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_22010) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_22010) ProtoMessage() {}

func (x *SC_22010) ProtoReflect() protoreflect.Message {
	mi := &file_SC_22010_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_22010.ProtoReflect.Descriptor instead.
func (*SC_22010) Descriptor() ([]byte, []int) {
	return file_SC_22010_proto_rawDescGZIP(), []int{0}
}

func (x *SC_22010) GetResult() uint32 {
	if x != nil && x.Result != nil {
		return *x.Result
	}
	return 0
}

func (x *SC_22010) GetExpInWell() uint32 {
	if x != nil && x.ExpInWell != nil {
		return *x.ExpInWell
	}
	return 0
}

var File_SC_22010_proto protoreflect.FileDescriptor

var file_SC_22010_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x32, 0x32, 0x30, 0x31, 0x30, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x42, 0x0a, 0x08, 0x53, 0x43, 0x5f,
	0x32, 0x32, 0x30, 0x31, 0x30, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18,
	0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1e, 0x0a,
	0x0b, 0x65, 0x78, 0x70, 0x5f, 0x69, 0x6e, 0x5f, 0x77, 0x65, 0x6c, 0x6c, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x09, 0x65, 0x78, 0x70, 0x49, 0x6e, 0x57, 0x65, 0x6c, 0x6c, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_SC_22010_proto_rawDescOnce sync.Once
	file_SC_22010_proto_rawDescData = file_SC_22010_proto_rawDesc
)

func file_SC_22010_proto_rawDescGZIP() []byte {
	file_SC_22010_proto_rawDescOnce.Do(func() {
		file_SC_22010_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_22010_proto_rawDescData)
	})
	return file_SC_22010_proto_rawDescData
}

var file_SC_22010_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_22010_proto_goTypes = []any{
	(*SC_22010)(nil), // 0: belfast.SC_22010
}
var file_SC_22010_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SC_22010_proto_init() }
func file_SC_22010_proto_init() {
	if File_SC_22010_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SC_22010_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_22010); i {
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
			RawDescriptor: file_SC_22010_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_22010_proto_goTypes,
		DependencyIndexes: file_SC_22010_proto_depIdxs,
		MessageInfos:      file_SC_22010_proto_msgTypes,
	}.Build()
	File_SC_22010_proto = out.File
	file_SC_22010_proto_rawDesc = nil
	file_SC_22010_proto_goTypes = nil
	file_SC_22010_proto_depIdxs = nil
}
