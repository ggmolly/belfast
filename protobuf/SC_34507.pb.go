// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_34507.proto

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

type SC_34507 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BossInfo *WORLDBOSS_INFO `protobuf:"bytes,1,req,name=boss_info,json=bossInfo" json:"boss_info,omitempty"`
	UserInfo *USERSIMPLEINFO `protobuf:"bytes,2,req,name=user_info,json=userInfo" json:"user_info,omitempty"`
	Type     *uint32         `protobuf:"varint,3,req,name=type" json:"type,omitempty"`
}

func (x *SC_34507) Reset() {
	*x = SC_34507{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_34507_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_34507) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_34507) ProtoMessage() {}

func (x *SC_34507) ProtoReflect() protoreflect.Message {
	mi := &file_SC_34507_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_34507.ProtoReflect.Descriptor instead.
func (*SC_34507) Descriptor() ([]byte, []int) {
	return file_SC_34507_proto_rawDescGZIP(), []int{0}
}

func (x *SC_34507) GetBossInfo() *WORLDBOSS_INFO {
	if x != nil {
		return x.BossInfo
	}
	return nil
}

func (x *SC_34507) GetUserInfo() *USERSIMPLEINFO {
	if x != nil {
		return x.UserInfo
	}
	return nil
}

func (x *SC_34507) GetType() uint32 {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return 0
}

var File_SC_34507_proto protoreflect.FileDescriptor

var file_SC_34507_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x33, 0x34, 0x35, 0x30, 0x37, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x14, 0x57, 0x4f, 0x52, 0x4c, 0x44,
	0x42, 0x4f, 0x53, 0x53, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x14, 0x55, 0x53, 0x45, 0x52, 0x53, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x49, 0x4e, 0x46, 0x4f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x01, 0x0a, 0x08, 0x53, 0x43, 0x5f, 0x33, 0x34, 0x35,
	0x30, 0x37, 0x12, 0x34, 0x0a, 0x09, 0x62, 0x6f, 0x73, 0x73, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18,
	0x01, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e,
	0x57, 0x4f, 0x52, 0x4c, 0x44, 0x42, 0x4f, 0x53, 0x53, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x08,
	0x62, 0x6f, 0x73, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x34, 0x0a, 0x09, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65,
	0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x55, 0x53, 0x45, 0x52, 0x53, 0x49, 0x4d, 0x50, 0x4c, 0x45,
	0x49, 0x4e, 0x46, 0x4f, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_SC_34507_proto_rawDescOnce sync.Once
	file_SC_34507_proto_rawDescData = file_SC_34507_proto_rawDesc
)

func file_SC_34507_proto_rawDescGZIP() []byte {
	file_SC_34507_proto_rawDescOnce.Do(func() {
		file_SC_34507_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_34507_proto_rawDescData)
	})
	return file_SC_34507_proto_rawDescData
}

var file_SC_34507_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_34507_proto_goTypes = []any{
	(*SC_34507)(nil),       // 0: belfast.SC_34507
	(*WORLDBOSS_INFO)(nil), // 1: belfast.WORLDBOSS_INFO
	(*USERSIMPLEINFO)(nil), // 2: belfast.USERSIMPLEINFO
}
var file_SC_34507_proto_depIdxs = []int32{
	1, // 0: belfast.SC_34507.boss_info:type_name -> belfast.WORLDBOSS_INFO
	2, // 1: belfast.SC_34507.user_info:type_name -> belfast.USERSIMPLEINFO
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_SC_34507_proto_init() }
func file_SC_34507_proto_init() {
	if File_SC_34507_proto != nil {
		return
	}
	file_WORLDBOSS_INFO_proto_init()
	file_USERSIMPLEINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SC_34507_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_34507); i {
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
			RawDescriptor: file_SC_34507_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_34507_proto_goTypes,
		DependencyIndexes: file_SC_34507_proto_depIdxs,
		MessageInfos:      file_SC_34507_proto_msgTypes,
	}.Build()
	File_SC_34507_proto = out.File
	file_SC_34507_proto_rawDesc = nil
	file_SC_34507_proto_goTypes = nil
	file_SC_34507_proto_depIdxs = nil
}
