// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_12024.proto

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

type SC_12024 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorklistCount *uint32      `protobuf:"varint,1,req,name=worklist_count,json=worklistCount" json:"worklist_count,omitempty"`
	WorklistList  []*BUILDINFO `protobuf:"bytes,2,rep,name=worklist_list,json=worklistList" json:"worklist_list,omitempty"`
	DrawCount_1   *uint32      `protobuf:"varint,3,req,name=draw_count_1,json=drawCount1" json:"draw_count_1,omitempty"`
	DrawCount_10  *uint32      `protobuf:"varint,4,req,name=draw_count_10,json=drawCount10" json:"draw_count_10,omitempty"`
	ExchangeCount *uint32      `protobuf:"varint,5,req,name=exchange_count,json=exchangeCount" json:"exchange_count,omitempty"`
}

func (x *SC_12024) Reset() {
	*x = SC_12024{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_12024_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_12024) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_12024) ProtoMessage() {}

func (x *SC_12024) ProtoReflect() protoreflect.Message {
	mi := &file_SC_12024_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_12024.ProtoReflect.Descriptor instead.
func (*SC_12024) Descriptor() ([]byte, []int) {
	return file_SC_12024_proto_rawDescGZIP(), []int{0}
}

func (x *SC_12024) GetWorklistCount() uint32 {
	if x != nil && x.WorklistCount != nil {
		return *x.WorklistCount
	}
	return 0
}

func (x *SC_12024) GetWorklistList() []*BUILDINFO {
	if x != nil {
		return x.WorklistList
	}
	return nil
}

func (x *SC_12024) GetDrawCount_1() uint32 {
	if x != nil && x.DrawCount_1 != nil {
		return *x.DrawCount_1
	}
	return 0
}

func (x *SC_12024) GetDrawCount_10() uint32 {
	if x != nil && x.DrawCount_10 != nil {
		return *x.DrawCount_10
	}
	return 0
}

func (x *SC_12024) GetExchangeCount() uint32 {
	if x != nil && x.ExchangeCount != nil {
		return *x.ExchangeCount
	}
	return 0
}

var File_SC_12024_proto protoreflect.FileDescriptor

var file_SC_12024_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x31, 0x32, 0x30, 0x32, 0x34, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x0f, 0x42, 0x55, 0x49, 0x4c, 0x44,
	0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd7, 0x01, 0x0a, 0x08, 0x53,
	0x43, 0x5f, 0x31, 0x32, 0x30, 0x32, 0x34, 0x12, 0x25, 0x0a, 0x0e, 0x77, 0x6f, 0x72, 0x6b, 0x6c,
	0x69, 0x73, 0x74, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x0d, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x37,
	0x0a, 0x0d, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e,
	0x42, 0x55, 0x49, 0x4c, 0x44, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x6c,
	0x69, 0x73, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0c, 0x64, 0x72, 0x61, 0x77, 0x5f,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x31, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x64,
	0x72, 0x61, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x31, 0x12, 0x22, 0x0a, 0x0d, 0x64, 0x72, 0x61,
	0x77, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x31, 0x30, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x0b, 0x64, 0x72, 0x61, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x31, 0x30, 0x12, 0x25, 0x0a,
	0x0e, 0x65, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x05, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0d, 0x65, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66,
}

var (
	file_SC_12024_proto_rawDescOnce sync.Once
	file_SC_12024_proto_rawDescData = file_SC_12024_proto_rawDesc
)

func file_SC_12024_proto_rawDescGZIP() []byte {
	file_SC_12024_proto_rawDescOnce.Do(func() {
		file_SC_12024_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_12024_proto_rawDescData)
	})
	return file_SC_12024_proto_rawDescData
}

var file_SC_12024_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_12024_proto_goTypes = []any{
	(*SC_12024)(nil),  // 0: belfast.SC_12024
	(*BUILDINFO)(nil), // 1: belfast.BUILDINFO
}
var file_SC_12024_proto_depIdxs = []int32{
	1, // 0: belfast.SC_12024.worklist_list:type_name -> belfast.BUILDINFO
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_SC_12024_proto_init() }
func file_SC_12024_proto_init() {
	if File_SC_12024_proto != nil {
		return
	}
	file_BUILDINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SC_12024_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_12024); i {
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
			RawDescriptor: file_SC_12024_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_12024_proto_goTypes,
		DependencyIndexes: file_SC_12024_proto_depIdxs,
		MessageInfos:      file_SC_12024_proto_msgTypes,
	}.Build()
	File_SC_12024_proto = out.File
	file_SC_12024_proto_rawDesc = nil
	file_SC_12024_proto_goTypes = nil
	file_SC_12024_proto_depIdxs = nil
}
