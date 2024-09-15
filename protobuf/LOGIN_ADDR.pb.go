// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: LOGIN_ADDR.proto

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

type LOGIN_ADDR struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Desc      *string `protobuf:"bytes,1,req,name=desc" json:"desc,omitempty"`
	Ip        *string `protobuf:"bytes,2,req,name=ip" json:"ip,omitempty"`
	Port      *uint32 `protobuf:"varint,3,req,name=port" json:"port,omitempty"`
	ProxyIp   *string `protobuf:"bytes,4,opt,name=proxy_ip,json=proxyIp" json:"proxy_ip,omitempty"`
	ProxyPort *uint32 `protobuf:"varint,5,opt,name=proxy_port,json=proxyPort" json:"proxy_port,omitempty"`
	Type      *uint32 `protobuf:"varint,6,req,name=type" json:"type,omitempty"`
}

func (x *LOGIN_ADDR) Reset() {
	*x = LOGIN_ADDR{}
	if protoimpl.UnsafeEnabled {
		mi := &file_LOGIN_ADDR_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LOGIN_ADDR) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LOGIN_ADDR) ProtoMessage() {}

func (x *LOGIN_ADDR) ProtoReflect() protoreflect.Message {
	mi := &file_LOGIN_ADDR_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LOGIN_ADDR.ProtoReflect.Descriptor instead.
func (*LOGIN_ADDR) Descriptor() ([]byte, []int) {
	return file_LOGIN_ADDR_proto_rawDescGZIP(), []int{0}
}

func (x *LOGIN_ADDR) GetDesc() string {
	if x != nil && x.Desc != nil {
		return *x.Desc
	}
	return ""
}

func (x *LOGIN_ADDR) GetIp() string {
	if x != nil && x.Ip != nil {
		return *x.Ip
	}
	return ""
}

func (x *LOGIN_ADDR) GetPort() uint32 {
	if x != nil && x.Port != nil {
		return *x.Port
	}
	return 0
}

func (x *LOGIN_ADDR) GetProxyIp() string {
	if x != nil && x.ProxyIp != nil {
		return *x.ProxyIp
	}
	return ""
}

func (x *LOGIN_ADDR) GetProxyPort() uint32 {
	if x != nil && x.ProxyPort != nil {
		return *x.ProxyPort
	}
	return 0
}

func (x *LOGIN_ADDR) GetType() uint32 {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return 0
}

var File_LOGIN_ADDR_proto protoreflect.FileDescriptor

var file_LOGIN_ADDR_proto_rawDesc = []byte{
	0x0a, 0x10, 0x4c, 0x4f, 0x47, 0x49, 0x4e, 0x5f, 0x41, 0x44, 0x44, 0x52, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x92, 0x01, 0x0a, 0x0a,
	0x4c, 0x4f, 0x47, 0x49, 0x4e, 0x5f, 0x41, 0x44, 0x44, 0x52, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65,
	0x73, 0x63, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x70, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x69, 0x70, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x49, 0x70, 0x12, 0x1d, 0x0a,
	0x0a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_LOGIN_ADDR_proto_rawDescOnce sync.Once
	file_LOGIN_ADDR_proto_rawDescData = file_LOGIN_ADDR_proto_rawDesc
)

func file_LOGIN_ADDR_proto_rawDescGZIP() []byte {
	file_LOGIN_ADDR_proto_rawDescOnce.Do(func() {
		file_LOGIN_ADDR_proto_rawDescData = protoimpl.X.CompressGZIP(file_LOGIN_ADDR_proto_rawDescData)
	})
	return file_LOGIN_ADDR_proto_rawDescData
}

var file_LOGIN_ADDR_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_LOGIN_ADDR_proto_goTypes = []any{
	(*LOGIN_ADDR)(nil), // 0: belfast.LOGIN_ADDR
}
var file_LOGIN_ADDR_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_LOGIN_ADDR_proto_init() }
func file_LOGIN_ADDR_proto_init() {
	if File_LOGIN_ADDR_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_LOGIN_ADDR_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*LOGIN_ADDR); i {
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
			RawDescriptor: file_LOGIN_ADDR_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_LOGIN_ADDR_proto_goTypes,
		DependencyIndexes: file_LOGIN_ADDR_proto_depIdxs,
		MessageInfos:      file_LOGIN_ADDR_proto_msgTypes,
	}.Build()
	File_LOGIN_ADDR_proto = out.File
	file_LOGIN_ADDR_proto_rawDesc = nil
	file_LOGIN_ADDR_proto_goTypes = nil
	file_LOGIN_ADDR_proto_depIdxs = nil
}
