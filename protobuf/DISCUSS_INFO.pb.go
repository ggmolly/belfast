// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: DISCUSS_INFO.proto

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

type DISCUSS_INFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShipGroupId       *uint32         `protobuf:"varint,1,req,name=ship_group_id,json=shipGroupId" json:"ship_group_id,omitempty"`
	DiscussCount      *uint32         `protobuf:"varint,2,req,name=discuss_count,json=discussCount" json:"discuss_count,omitempty"`
	HeartCount        *uint32         `protobuf:"varint,3,req,name=heart_count,json=heartCount" json:"heart_count,omitempty"`
	DiscussList       []*DISCUSS_INFO `protobuf:"bytes,4,rep,name=discuss_list,json=discussList" json:"discuss_list,omitempty"`
	DailyDiscussCount *uint32         `protobuf:"varint,5,req,name=daily_discuss_count,json=dailyDiscussCount" json:"daily_discuss_count,omitempty"`
}

func (x *DISCUSS_INFO) Reset() {
	*x = DISCUSS_INFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DISCUSS_INFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DISCUSS_INFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DISCUSS_INFO) ProtoMessage() {}

func (x *DISCUSS_INFO) ProtoReflect() protoreflect.Message {
	mi := &file_DISCUSS_INFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DISCUSS_INFO.ProtoReflect.Descriptor instead.
func (*DISCUSS_INFO) Descriptor() ([]byte, []int) {
	return file_DISCUSS_INFO_proto_rawDescGZIP(), []int{0}
}

func (x *DISCUSS_INFO) GetShipGroupId() uint32 {
	if x != nil && x.ShipGroupId != nil {
		return *x.ShipGroupId
	}
	return 0
}

func (x *DISCUSS_INFO) GetDiscussCount() uint32 {
	if x != nil && x.DiscussCount != nil {
		return *x.DiscussCount
	}
	return 0
}

func (x *DISCUSS_INFO) GetHeartCount() uint32 {
	if x != nil && x.HeartCount != nil {
		return *x.HeartCount
	}
	return 0
}

func (x *DISCUSS_INFO) GetDiscussList() []*DISCUSS_INFO {
	if x != nil {
		return x.DiscussList
	}
	return nil
}

func (x *DISCUSS_INFO) GetDailyDiscussCount() uint32 {
	if x != nil && x.DailyDiscussCount != nil {
		return *x.DailyDiscussCount
	}
	return 0
}

var File_DISCUSS_INFO_proto protoreflect.FileDescriptor

var file_DISCUSS_INFO_proto_rawDesc = []byte{
	0x0a, 0x12, 0x44, 0x49, 0x53, 0x43, 0x55, 0x53, 0x53, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0xe2, 0x01,
	0x0a, 0x0c, 0x44, 0x49, 0x53, 0x43, 0x55, 0x53, 0x53, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x12, 0x22,
	0x0a, 0x0d, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0b, 0x73, 0x68, 0x69, 0x70, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x69, 0x73, 0x63, 0x75, 0x73, 0x73, 0x5f, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0c, 0x64, 0x69, 0x73, 0x63, 0x75,
	0x73, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x68, 0x65, 0x61, 0x72, 0x74,
	0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x68, 0x65,
	0x61, 0x72, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x63,
	0x75, 0x73, 0x73, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x44, 0x49, 0x53, 0x43, 0x55, 0x53, 0x53,
	0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x63, 0x75, 0x73, 0x73, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x2e, 0x0a, 0x13, 0x64, 0x61, 0x69, 0x6c, 0x79, 0x5f, 0x64, 0x69, 0x73, 0x63,
	0x75, 0x73, 0x73, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x11, 0x64, 0x61, 0x69, 0x6c, 0x79, 0x44, 0x69, 0x73, 0x63, 0x75, 0x73, 0x73, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_DISCUSS_INFO_proto_rawDescOnce sync.Once
	file_DISCUSS_INFO_proto_rawDescData = file_DISCUSS_INFO_proto_rawDesc
)

func file_DISCUSS_INFO_proto_rawDescGZIP() []byte {
	file_DISCUSS_INFO_proto_rawDescOnce.Do(func() {
		file_DISCUSS_INFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_DISCUSS_INFO_proto_rawDescData)
	})
	return file_DISCUSS_INFO_proto_rawDescData
}

var file_DISCUSS_INFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_DISCUSS_INFO_proto_goTypes = []any{
	(*DISCUSS_INFO)(nil), // 0: belfast.DISCUSS_INFO
}
var file_DISCUSS_INFO_proto_depIdxs = []int32{
	0, // 0: belfast.DISCUSS_INFO.discuss_list:type_name -> belfast.DISCUSS_INFO
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_DISCUSS_INFO_proto_init() }
func file_DISCUSS_INFO_proto_init() {
	if File_DISCUSS_INFO_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_DISCUSS_INFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*DISCUSS_INFO); i {
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
			RawDescriptor: file_DISCUSS_INFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_DISCUSS_INFO_proto_goTypes,
		DependencyIndexes: file_DISCUSS_INFO_proto_depIdxs,
		MessageInfos:      file_DISCUSS_INFO_proto_msgTypes,
	}.Build()
	File_DISCUSS_INFO_proto = out.File
	file_DISCUSS_INFO_proto_rawDesc = nil
	file_DISCUSS_INFO_proto_goTypes = nil
	file_DISCUSS_INFO_proto_depIdxs = nil
}
