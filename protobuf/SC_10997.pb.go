// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_10997.proto

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

type SC_10997 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version1    *uint32 `protobuf:"varint,1,req,name=version1" json:"version1,omitempty"`
	Version2    *uint32 `protobuf:"varint,2,req,name=version2" json:"version2,omitempty"`
	Version3    *uint32 `protobuf:"varint,3,req,name=version3" json:"version3,omitempty"`
	Version4    *uint32 `protobuf:"varint,4,req,name=version4" json:"version4,omitempty"`
	GatewayIp   *string `protobuf:"bytes,5,req,name=gateway_ip,json=gatewayIp" json:"gateway_ip,omitempty"`
	GatewayPort *uint32 `protobuf:"varint,6,req,name=gateway_port,json=gatewayPort" json:"gateway_port,omitempty"`
	Url         *string `protobuf:"bytes,7,req,name=url" json:"url,omitempty"`
}

func (x *SC_10997) Reset() {
	*x = SC_10997{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_10997_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_10997) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_10997) ProtoMessage() {}

func (x *SC_10997) ProtoReflect() protoreflect.Message {
	mi := &file_SC_10997_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_10997.ProtoReflect.Descriptor instead.
func (*SC_10997) Descriptor() ([]byte, []int) {
	return file_SC_10997_proto_rawDescGZIP(), []int{0}
}

func (x *SC_10997) GetVersion1() uint32 {
	if x != nil && x.Version1 != nil {
		return *x.Version1
	}
	return 0
}

func (x *SC_10997) GetVersion2() uint32 {
	if x != nil && x.Version2 != nil {
		return *x.Version2
	}
	return 0
}

func (x *SC_10997) GetVersion3() uint32 {
	if x != nil && x.Version3 != nil {
		return *x.Version3
	}
	return 0
}

func (x *SC_10997) GetVersion4() uint32 {
	if x != nil && x.Version4 != nil {
		return *x.Version4
	}
	return 0
}

func (x *SC_10997) GetGatewayIp() string {
	if x != nil && x.GatewayIp != nil {
		return *x.GatewayIp
	}
	return ""
}

func (x *SC_10997) GetGatewayPort() uint32 {
	if x != nil && x.GatewayPort != nil {
		return *x.GatewayPort
	}
	return 0
}

func (x *SC_10997) GetUrl() string {
	if x != nil && x.Url != nil {
		return *x.Url
	}
	return ""
}

var File_SC_10997_proto protoreflect.FileDescriptor

var file_SC_10997_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x31, 0x30, 0x39, 0x39, 0x37, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0xce, 0x01, 0x0a, 0x08, 0x53, 0x43,
	0x5f, 0x31, 0x30, 0x39, 0x39, 0x37, 0x12, 0x1a, 0x0a, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x31, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x31, 0x12, 0x1a, 0x0a, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x32, 0x18, 0x02,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x32, 0x12, 0x1a,
	0x0a, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x33, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x33, 0x12, 0x1a, 0x0a, 0x08, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x34, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x34, 0x12, 0x1d, 0x0a, 0x0a, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x5f, 0x69, 0x70, 0x18, 0x05, 0x20, 0x02, 0x28, 0x09, 0x52, 0x09, 0x67, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x49, 0x70, 0x12, 0x21, 0x0a, 0x0c, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0b, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18,
	0x07, 0x20, 0x02, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_SC_10997_proto_rawDescOnce sync.Once
	file_SC_10997_proto_rawDescData = file_SC_10997_proto_rawDesc
)

func file_SC_10997_proto_rawDescGZIP() []byte {
	file_SC_10997_proto_rawDescOnce.Do(func() {
		file_SC_10997_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_10997_proto_rawDescData)
	})
	return file_SC_10997_proto_rawDescData
}

var file_SC_10997_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_10997_proto_goTypes = []any{
	(*SC_10997)(nil), // 0: belfast.SC_10997
}
var file_SC_10997_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SC_10997_proto_init() }
func file_SC_10997_proto_init() {
	if File_SC_10997_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SC_10997_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_10997); i {
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
			RawDescriptor: file_SC_10997_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_10997_proto_goTypes,
		DependencyIndexes: file_SC_10997_proto_depIdxs,
		MessageInfos:      file_SC_10997_proto_msgTypes,
	}.Build()
	File_SC_10997_proto = out.File
	file_SC_10997_proto_rawDesc = nil
	file_SC_10997_proto_goTypes = nil
	file_SC_10997_proto_depIdxs = nil
}
