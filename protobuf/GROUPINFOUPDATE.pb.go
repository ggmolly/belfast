// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: GROUPINFOUPDATE.proto

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

type GROUPINFOUPDATE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       *uint32      `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	BuffList []*BUFF_INFO `protobuf:"bytes,2,rep,name=buff_list,json=buffList" json:"buff_list,omitempty"`
}

func (x *GROUPINFOUPDATE) Reset() {
	*x = GROUPINFOUPDATE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GROUPINFOUPDATE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GROUPINFOUPDATE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GROUPINFOUPDATE) ProtoMessage() {}

func (x *GROUPINFOUPDATE) ProtoReflect() protoreflect.Message {
	mi := &file_GROUPINFOUPDATE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GROUPINFOUPDATE.ProtoReflect.Descriptor instead.
func (*GROUPINFOUPDATE) Descriptor() ([]byte, []int) {
	return file_GROUPINFOUPDATE_proto_rawDescGZIP(), []int{0}
}

func (x *GROUPINFOUPDATE) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *GROUPINFOUPDATE) GetBuffList() []*BUFF_INFO {
	if x != nil {
		return x.BuffList
	}
	return nil
}

var File_GROUPINFOUPDATE_proto protoreflect.FileDescriptor

var file_GROUPINFOUPDATE_proto_rawDesc = []byte{
	0x0a, 0x15, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x55, 0x50, 0x44, 0x41, 0x54,
	0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74,
	0x1a, 0x0f, 0x42, 0x55, 0x46, 0x46, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x52, 0x0a, 0x0f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x55, 0x50,
	0x44, 0x41, 0x54, 0x45, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x2f, 0x0a, 0x09, 0x62, 0x75, 0x66, 0x66, 0x5f, 0x6c, 0x69, 0x73,
	0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x2e, 0x42, 0x55, 0x46, 0x46, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x08, 0x62, 0x75, 0x66,
	0x66, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_GROUPINFOUPDATE_proto_rawDescOnce sync.Once
	file_GROUPINFOUPDATE_proto_rawDescData = file_GROUPINFOUPDATE_proto_rawDesc
)

func file_GROUPINFOUPDATE_proto_rawDescGZIP() []byte {
	file_GROUPINFOUPDATE_proto_rawDescOnce.Do(func() {
		file_GROUPINFOUPDATE_proto_rawDescData = protoimpl.X.CompressGZIP(file_GROUPINFOUPDATE_proto_rawDescData)
	})
	return file_GROUPINFOUPDATE_proto_rawDescData
}

var file_GROUPINFOUPDATE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_GROUPINFOUPDATE_proto_goTypes = []any{
	(*GROUPINFOUPDATE)(nil), // 0: belfast.GROUPINFOUPDATE
	(*BUFF_INFO)(nil),       // 1: belfast.BUFF_INFO
}
var file_GROUPINFOUPDATE_proto_depIdxs = []int32{
	1, // 0: belfast.GROUPINFOUPDATE.buff_list:type_name -> belfast.BUFF_INFO
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_GROUPINFOUPDATE_proto_init() }
func file_GROUPINFOUPDATE_proto_init() {
	if File_GROUPINFOUPDATE_proto != nil {
		return
	}
	file_BUFF_INFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_GROUPINFOUPDATE_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GROUPINFOUPDATE); i {
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
			RawDescriptor: file_GROUPINFOUPDATE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_GROUPINFOUPDATE_proto_goTypes,
		DependencyIndexes: file_GROUPINFOUPDATE_proto_depIdxs,
		MessageInfos:      file_GROUPINFOUPDATE_proto_msgTypes,
	}.Build()
	File_GROUPINFOUPDATE_proto = out.File
	file_GROUPINFOUPDATE_proto_rawDesc = nil
	file_GROUPINFOUPDATE_proto_goTypes = nil
	file_GROUPINFOUPDATE_proto_depIdxs = nil
}
