// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: CS_17105.proto

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

type CS_17105 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShipGroupId *uint32 `protobuf:"varint,1,req,name=ship_group_id,json=shipGroupId" json:"ship_group_id,omitempty"`
	DiscussId   *uint32 `protobuf:"varint,2,req,name=discuss_id,json=discussId" json:"discuss_id,omitempty"`
	GoodOrBad   *uint32 `protobuf:"varint,3,req,name=good_or_bad,json=goodOrBad" json:"good_or_bad,omitempty"`
}

func (x *CS_17105) Reset() {
	*x = CS_17105{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CS_17105_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CS_17105) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CS_17105) ProtoMessage() {}

func (x *CS_17105) ProtoReflect() protoreflect.Message {
	mi := &file_CS_17105_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CS_17105.ProtoReflect.Descriptor instead.
func (*CS_17105) Descriptor() ([]byte, []int) {
	return file_CS_17105_proto_rawDescGZIP(), []int{0}
}

func (x *CS_17105) GetShipGroupId() uint32 {
	if x != nil && x.ShipGroupId != nil {
		return *x.ShipGroupId
	}
	return 0
}

func (x *CS_17105) GetDiscussId() uint32 {
	if x != nil && x.DiscussId != nil {
		return *x.DiscussId
	}
	return 0
}

func (x *CS_17105) GetGoodOrBad() uint32 {
	if x != nil && x.GoodOrBad != nil {
		return *x.GoodOrBad
	}
	return 0
}

var File_CS_17105_proto protoreflect.FileDescriptor

var file_CS_17105_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x43, 0x53, 0x5f, 0x31, 0x37, 0x31, 0x30, 0x35, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x6d, 0x0a, 0x08, 0x43, 0x53, 0x5f,
	0x31, 0x37, 0x31, 0x30, 0x35, 0x12, 0x22, 0x0a, 0x0d, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0b, 0x73, 0x68,
	0x69, 0x70, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x64, 0x69, 0x73,
	0x63, 0x75, 0x73, 0x73, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x09, 0x64,
	0x69, 0x73, 0x63, 0x75, 0x73, 0x73, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x0b, 0x67, 0x6f, 0x6f, 0x64,
	0x5f, 0x6f, 0x72, 0x5f, 0x62, 0x61, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x09, 0x67,
	0x6f, 0x6f, 0x64, 0x4f, 0x72, 0x42, 0x61, 0x64, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CS_17105_proto_rawDescOnce sync.Once
	file_CS_17105_proto_rawDescData = file_CS_17105_proto_rawDesc
)

func file_CS_17105_proto_rawDescGZIP() []byte {
	file_CS_17105_proto_rawDescOnce.Do(func() {
		file_CS_17105_proto_rawDescData = protoimpl.X.CompressGZIP(file_CS_17105_proto_rawDescData)
	})
	return file_CS_17105_proto_rawDescData
}

var file_CS_17105_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CS_17105_proto_goTypes = []any{
	(*CS_17105)(nil), // 0: belfast.CS_17105
}
var file_CS_17105_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CS_17105_proto_init() }
func file_CS_17105_proto_init() {
	if File_CS_17105_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_CS_17105_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CS_17105); i {
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
			RawDescriptor: file_CS_17105_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CS_17105_proto_goTypes,
		DependencyIndexes: file_CS_17105_proto_depIdxs,
		MessageInfos:      file_CS_17105_proto_msgTypes,
	}.Build()
	File_CS_17105_proto = out.File
	file_CS_17105_proto_rawDesc = nil
	file_CS_17105_proto_goTypes = nil
	file_CS_17105_proto_depIdxs = nil
}
