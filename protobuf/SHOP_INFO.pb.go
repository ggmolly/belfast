// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SHOP_INFO.proto

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

type SHOP_INFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RefreshCount    *uint32       `protobuf:"varint,1,req,name=refresh_count,json=refreshCount" json:"refresh_count,omitempty"`
	NextRefreshTime *uint32       `protobuf:"varint,2,req,name=next_refresh_time,json=nextRefreshTime" json:"next_refresh_time,omitempty"`
	GoodList        []*GOODS_INFO `protobuf:"bytes,3,rep,name=good_list,json=goodList" json:"good_list,omitempty"`
}

func (x *SHOP_INFO) Reset() {
	*x = SHOP_INFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SHOP_INFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SHOP_INFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SHOP_INFO) ProtoMessage() {}

func (x *SHOP_INFO) ProtoReflect() protoreflect.Message {
	mi := &file_SHOP_INFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SHOP_INFO.ProtoReflect.Descriptor instead.
func (*SHOP_INFO) Descriptor() ([]byte, []int) {
	return file_SHOP_INFO_proto_rawDescGZIP(), []int{0}
}

func (x *SHOP_INFO) GetRefreshCount() uint32 {
	if x != nil && x.RefreshCount != nil {
		return *x.RefreshCount
	}
	return 0
}

func (x *SHOP_INFO) GetNextRefreshTime() uint32 {
	if x != nil && x.NextRefreshTime != nil {
		return *x.NextRefreshTime
	}
	return 0
}

func (x *SHOP_INFO) GetGoodList() []*GOODS_INFO {
	if x != nil {
		return x.GoodList
	}
	return nil
}

var File_SHOP_INFO_proto protoreflect.FileDescriptor

var file_SHOP_INFO_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x53, 0x48, 0x4f, 0x50, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x10, 0x47, 0x4f, 0x4f, 0x44,
	0x53, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x01, 0x0a,
	0x09, 0x53, 0x48, 0x4f, 0x50, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65,
	0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x0c, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x2a, 0x0a, 0x11, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0f, 0x6e, 0x65, 0x78, 0x74,
	0x52, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x30, 0x0a, 0x09, 0x67,
	0x6f, 0x6f, 0x64, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x47, 0x4f, 0x4f, 0x44, 0x53, 0x5f, 0x49,
	0x4e, 0x46, 0x4f, 0x52, 0x08, 0x67, 0x6f, 0x6f, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_SHOP_INFO_proto_rawDescOnce sync.Once
	file_SHOP_INFO_proto_rawDescData = file_SHOP_INFO_proto_rawDesc
)

func file_SHOP_INFO_proto_rawDescGZIP() []byte {
	file_SHOP_INFO_proto_rawDescOnce.Do(func() {
		file_SHOP_INFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_SHOP_INFO_proto_rawDescData)
	})
	return file_SHOP_INFO_proto_rawDescData
}

var file_SHOP_INFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SHOP_INFO_proto_goTypes = []any{
	(*SHOP_INFO)(nil),  // 0: belfast.SHOP_INFO
	(*GOODS_INFO)(nil), // 1: belfast.GOODS_INFO
}
var file_SHOP_INFO_proto_depIdxs = []int32{
	1, // 0: belfast.SHOP_INFO.good_list:type_name -> belfast.GOODS_INFO
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_SHOP_INFO_proto_init() }
func file_SHOP_INFO_proto_init() {
	if File_SHOP_INFO_proto != nil {
		return
	}
	file_GOODS_INFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SHOP_INFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SHOP_INFO); i {
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
			RawDescriptor: file_SHOP_INFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SHOP_INFO_proto_goTypes,
		DependencyIndexes: file_SHOP_INFO_proto_depIdxs,
		MessageInfos:      file_SHOP_INFO_proto_msgTypes,
	}.Build()
	File_SHOP_INFO_proto = out.File
	file_SHOP_INFO_proto_rawDesc = nil
	file_SHOP_INFO_proto_goTypes = nil
	file_SHOP_INFO_proto_depIdxs = nil
}
