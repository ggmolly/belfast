// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: MAIL_TITLE.proto

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

type MAIL_TITLE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    *uint32 `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Title *string `protobuf:"bytes,2,req,name=title" json:"title,omitempty"`
}

func (x *MAIL_TITLE) Reset() {
	*x = MAIL_TITLE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_MAIL_TITLE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MAIL_TITLE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MAIL_TITLE) ProtoMessage() {}

func (x *MAIL_TITLE) ProtoReflect() protoreflect.Message {
	mi := &file_MAIL_TITLE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MAIL_TITLE.ProtoReflect.Descriptor instead.
func (*MAIL_TITLE) Descriptor() ([]byte, []int) {
	return file_MAIL_TITLE_proto_rawDescGZIP(), []int{0}
}

func (x *MAIL_TITLE) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *MAIL_TITLE) GetTitle() string {
	if x != nil && x.Title != nil {
		return *x.Title
	}
	return ""
}

var File_MAIL_TITLE_proto protoreflect.FileDescriptor

var file_MAIL_TITLE_proto_rawDesc = []byte{
	0x0a, 0x10, 0x4d, 0x41, 0x49, 0x4c, 0x5f, 0x54, 0x49, 0x54, 0x4c, 0x45, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x32, 0x0a, 0x0a, 0x4d,
	0x41, 0x49, 0x4c, 0x5f, 0x54, 0x49, 0x54, 0x4c, 0x45, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x42,
	0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_MAIL_TITLE_proto_rawDescOnce sync.Once
	file_MAIL_TITLE_proto_rawDescData = file_MAIL_TITLE_proto_rawDesc
)

func file_MAIL_TITLE_proto_rawDescGZIP() []byte {
	file_MAIL_TITLE_proto_rawDescOnce.Do(func() {
		file_MAIL_TITLE_proto_rawDescData = protoimpl.X.CompressGZIP(file_MAIL_TITLE_proto_rawDescData)
	})
	return file_MAIL_TITLE_proto_rawDescData
}

var file_MAIL_TITLE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_MAIL_TITLE_proto_goTypes = []any{
	(*MAIL_TITLE)(nil), // 0: belfast.MAIL_TITLE
}
var file_MAIL_TITLE_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_MAIL_TITLE_proto_init() }
func file_MAIL_TITLE_proto_init() {
	if File_MAIL_TITLE_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_MAIL_TITLE_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MAIL_TITLE); i {
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
			RawDescriptor: file_MAIL_TITLE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_MAIL_TITLE_proto_goTypes,
		DependencyIndexes: file_MAIL_TITLE_proto_depIdxs,
		MessageInfos:      file_MAIL_TITLE_proto_msgTypes,
	}.Build()
	File_MAIL_TITLE_proto = out.File
	file_MAIL_TITLE_proto_rawDesc = nil
	file_MAIL_TITLE_proto_goTypes = nil
	file_MAIL_TITLE_proto_depIdxs = nil
}
