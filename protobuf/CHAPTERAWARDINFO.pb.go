// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: CHAPTERAWARDINFO.proto

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

type CHAPTERAWARDINFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    *uint32 `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Award *uint32 `protobuf:"varint,2,req,name=award" json:"award,omitempty"`
	Flag  *uint32 `protobuf:"varint,3,req,name=flag" json:"flag,omitempty"`
}

func (x *CHAPTERAWARDINFO) Reset() {
	*x = CHAPTERAWARDINFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CHAPTERAWARDINFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CHAPTERAWARDINFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CHAPTERAWARDINFO) ProtoMessage() {}

func (x *CHAPTERAWARDINFO) ProtoReflect() protoreflect.Message {
	mi := &file_CHAPTERAWARDINFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CHAPTERAWARDINFO.ProtoReflect.Descriptor instead.
func (*CHAPTERAWARDINFO) Descriptor() ([]byte, []int) {
	return file_CHAPTERAWARDINFO_proto_rawDescGZIP(), []int{0}
}

func (x *CHAPTERAWARDINFO) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *CHAPTERAWARDINFO) GetAward() uint32 {
	if x != nil && x.Award != nil {
		return *x.Award
	}
	return 0
}

func (x *CHAPTERAWARDINFO) GetFlag() uint32 {
	if x != nil && x.Flag != nil {
		return *x.Flag
	}
	return 0
}

var File_CHAPTERAWARDINFO_proto protoreflect.FileDescriptor

var file_CHAPTERAWARDINFO_proto_rawDesc = []byte{
	0x0a, 0x16, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x41, 0x57, 0x41, 0x52, 0x44, 0x49, 0x4e,
	0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x22, 0x4c, 0x0a, 0x10, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x41, 0x57, 0x41, 0x52,
	0x44, 0x49, 0x4e, 0x46, 0x4f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x77, 0x61, 0x72, 0x64, 0x18, 0x02,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x61, 0x77, 0x61, 0x72, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x66,
	0x6c, 0x61, 0x67, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x42,
	0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CHAPTERAWARDINFO_proto_rawDescOnce sync.Once
	file_CHAPTERAWARDINFO_proto_rawDescData = file_CHAPTERAWARDINFO_proto_rawDesc
)

func file_CHAPTERAWARDINFO_proto_rawDescGZIP() []byte {
	file_CHAPTERAWARDINFO_proto_rawDescOnce.Do(func() {
		file_CHAPTERAWARDINFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_CHAPTERAWARDINFO_proto_rawDescData)
	})
	return file_CHAPTERAWARDINFO_proto_rawDescData
}

var file_CHAPTERAWARDINFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CHAPTERAWARDINFO_proto_goTypes = []any{
	(*CHAPTERAWARDINFO)(nil), // 0: belfast.CHAPTERAWARDINFO
}
var file_CHAPTERAWARDINFO_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CHAPTERAWARDINFO_proto_init() }
func file_CHAPTERAWARDINFO_proto_init() {
	if File_CHAPTERAWARDINFO_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_CHAPTERAWARDINFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CHAPTERAWARDINFO); i {
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
			RawDescriptor: file_CHAPTERAWARDINFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CHAPTERAWARDINFO_proto_goTypes,
		DependencyIndexes: file_CHAPTERAWARDINFO_proto_depIdxs,
		MessageInfos:      file_CHAPTERAWARDINFO_proto_msgTypes,
	}.Build()
	File_CHAPTERAWARDINFO_proto = out.File
	file_CHAPTERAWARDINFO_proto_rawDesc = nil
	file_CHAPTERAWARDINFO_proto_goTypes = nil
	file_CHAPTERAWARDINFO_proto_depIdxs = nil
}
