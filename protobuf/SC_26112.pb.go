// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_26112.proto

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

type SC_26112 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ranks []*FRIENDSCORE `protobuf:"bytes,1,rep,name=ranks" json:"ranks,omitempty"`
}

func (x *SC_26112) Reset() {
	*x = SC_26112{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_26112_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_26112) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_26112) ProtoMessage() {}

func (x *SC_26112) ProtoReflect() protoreflect.Message {
	mi := &file_SC_26112_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_26112.ProtoReflect.Descriptor instead.
func (*SC_26112) Descriptor() ([]byte, []int) {
	return file_SC_26112_proto_rawDescGZIP(), []int{0}
}

func (x *SC_26112) GetRanks() []*FRIENDSCORE {
	if x != nil {
		return x.Ranks
	}
	return nil
}

var File_SC_26112_proto protoreflect.FileDescriptor

var file_SC_26112_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x32, 0x36, 0x31, 0x31, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x11, 0x46, 0x52, 0x49, 0x45, 0x4e,
	0x44, 0x53, 0x43, 0x4f, 0x52, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x36, 0x0a, 0x08,
	0x53, 0x43, 0x5f, 0x32, 0x36, 0x31, 0x31, 0x32, 0x12, 0x2a, 0x0a, 0x05, 0x72, 0x61, 0x6e, 0x6b,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x2e, 0x46, 0x52, 0x49, 0x45, 0x4e, 0x44, 0x53, 0x43, 0x4f, 0x52, 0x45, 0x52, 0x05, 0x72,
	0x61, 0x6e, 0x6b, 0x73, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66,
}

var (
	file_SC_26112_proto_rawDescOnce sync.Once
	file_SC_26112_proto_rawDescData = file_SC_26112_proto_rawDesc
)

func file_SC_26112_proto_rawDescGZIP() []byte {
	file_SC_26112_proto_rawDescOnce.Do(func() {
		file_SC_26112_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_26112_proto_rawDescData)
	})
	return file_SC_26112_proto_rawDescData
}

var file_SC_26112_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_26112_proto_goTypes = []any{
	(*SC_26112)(nil),    // 0: belfast.SC_26112
	(*FRIENDSCORE)(nil), // 1: belfast.FRIENDSCORE
}
var file_SC_26112_proto_depIdxs = []int32{
	1, // 0: belfast.SC_26112.ranks:type_name -> belfast.FRIENDSCORE
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_SC_26112_proto_init() }
func file_SC_26112_proto_init() {
	if File_SC_26112_proto != nil {
		return
	}
	file_FRIENDSCORE_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SC_26112_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_26112); i {
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
			RawDescriptor: file_SC_26112_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_26112_proto_goTypes,
		DependencyIndexes: file_SC_26112_proto_depIdxs,
		MessageInfos:      file_SC_26112_proto_msgTypes,
	}.Build()
	File_SC_26112_proto = out.File
	file_SC_26112_proto_rawDesc = nil
	file_SC_26112_proto_goTypes = nil
	file_SC_26112_proto_depIdxs = nil
}
