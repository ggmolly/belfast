// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: CS_10001.proto

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

type CS_10001 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account  *string `protobuf:"bytes,1,req,name=account" json:"account,omitempty"`
	Password *string `protobuf:"bytes,2,req,name=password" json:"password,omitempty"`
	MailBox  *string `protobuf:"bytes,3,req,name=mail_box,json=mailBox" json:"mail_box,omitempty"`
}

func (x *CS_10001) Reset() {
	*x = CS_10001{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CS_10001_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CS_10001) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CS_10001) ProtoMessage() {}

func (x *CS_10001) ProtoReflect() protoreflect.Message {
	mi := &file_CS_10001_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CS_10001.ProtoReflect.Descriptor instead.
func (*CS_10001) Descriptor() ([]byte, []int) {
	return file_CS_10001_proto_rawDescGZIP(), []int{0}
}

func (x *CS_10001) GetAccount() string {
	if x != nil && x.Account != nil {
		return *x.Account
	}
	return ""
}

func (x *CS_10001) GetPassword() string {
	if x != nil && x.Password != nil {
		return *x.Password
	}
	return ""
}

func (x *CS_10001) GetMailBox() string {
	if x != nil && x.MailBox != nil {
		return *x.MailBox
	}
	return ""
}

var File_CS_10001_proto protoreflect.FileDescriptor

var file_CS_10001_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x43, 0x53, 0x5f, 0x31, 0x30, 0x30, 0x30, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x5b, 0x0a, 0x08, 0x43, 0x53, 0x5f,
	0x31, 0x30, 0x30, 0x30, 0x31, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28,
	0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x6d,
	0x61, 0x69, 0x6c, 0x5f, 0x62, 0x6f, 0x78, 0x18, 0x03, 0x20, 0x02, 0x28, 0x09, 0x52, 0x07, 0x6d,
	0x61, 0x69, 0x6c, 0x42, 0x6f, 0x78, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CS_10001_proto_rawDescOnce sync.Once
	file_CS_10001_proto_rawDescData = file_CS_10001_proto_rawDesc
)

func file_CS_10001_proto_rawDescGZIP() []byte {
	file_CS_10001_proto_rawDescOnce.Do(func() {
		file_CS_10001_proto_rawDescData = protoimpl.X.CompressGZIP(file_CS_10001_proto_rawDescData)
	})
	return file_CS_10001_proto_rawDescData
}

var file_CS_10001_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CS_10001_proto_goTypes = []any{
	(*CS_10001)(nil), // 0: belfast.CS_10001
}
var file_CS_10001_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CS_10001_proto_init() }
func file_CS_10001_proto_init() {
	if File_CS_10001_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_CS_10001_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CS_10001); i {
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
			RawDescriptor: file_CS_10001_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CS_10001_proto_goTypes,
		DependencyIndexes: file_CS_10001_proto_depIdxs,
		MessageInfos:      file_CS_10001_proto_msgTypes,
	}.Build()
	File_CS_10001_proto = out.File
	file_CS_10001_proto_rawDesc = nil
	file_CS_10001_proto_goTypes = nil
	file_CS_10001_proto_depIdxs = nil
}
