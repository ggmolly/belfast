// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: CS_40001.proto

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

type CS_40001 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	System          *uint32        `protobuf:"varint,1,req,name=system" json:"system,omitempty"`
	ShipIdList      []uint32       `protobuf:"varint,2,rep,name=ship_id_list,json=shipIdList" json:"ship_id_list,omitempty"`
	Data            *uint32        `protobuf:"varint,3,req,name=data" json:"data,omitempty"`
	Data2           []uint32       `protobuf:"varint,4,rep,name=data2" json:"data2,omitempty"`
	OtherShipIdList []*OTHERSHIPID `protobuf:"bytes,5,rep,name=other_ship_id_list,json=otherShipIdList" json:"other_ship_id_list,omitempty"`
}

func (x *CS_40001) Reset() {
	*x = CS_40001{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CS_40001_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CS_40001) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CS_40001) ProtoMessage() {}

func (x *CS_40001) ProtoReflect() protoreflect.Message {
	mi := &file_CS_40001_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CS_40001.ProtoReflect.Descriptor instead.
func (*CS_40001) Descriptor() ([]byte, []int) {
	return file_CS_40001_proto_rawDescGZIP(), []int{0}
}

func (x *CS_40001) GetSystem() uint32 {
	if x != nil && x.System != nil {
		return *x.System
	}
	return 0
}

func (x *CS_40001) GetShipIdList() []uint32 {
	if x != nil {
		return x.ShipIdList
	}
	return nil
}

func (x *CS_40001) GetData() uint32 {
	if x != nil && x.Data != nil {
		return *x.Data
	}
	return 0
}

func (x *CS_40001) GetData2() []uint32 {
	if x != nil {
		return x.Data2
	}
	return nil
}

func (x *CS_40001) GetOtherShipIdList() []*OTHERSHIPID {
	if x != nil {
		return x.OtherShipIdList
	}
	return nil
}

var File_CS_40001_proto protoreflect.FileDescriptor

var file_CS_40001_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x43, 0x53, 0x5f, 0x34, 0x30, 0x30, 0x30, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x11, 0x4f, 0x54, 0x48, 0x45, 0x52,
	0x53, 0x48, 0x49, 0x50, 0x49, 0x44, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x01, 0x0a,
	0x08, 0x43, 0x53, 0x5f, 0x34, 0x30, 0x30, 0x30, 0x31, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x12, 0x20, 0x0a, 0x0c, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x69, 0x64, 0x5f, 0x6c, 0x69, 0x73,
	0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x0a, 0x73, 0x68, 0x69, 0x70, 0x49, 0x64, 0x4c,
	0x69, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x61, 0x74, 0x61, 0x32,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x05, 0x64, 0x61, 0x74, 0x61, 0x32, 0x12, 0x41, 0x0a,
	0x12, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x5f, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x69, 0x64, 0x5f, 0x6c,
	0x69, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x62, 0x65, 0x6c, 0x66,
	0x61, 0x73, 0x74, 0x2e, 0x4f, 0x54, 0x48, 0x45, 0x52, 0x53, 0x48, 0x49, 0x50, 0x49, 0x44, 0x52,
	0x0f, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x53, 0x68, 0x69, 0x70, 0x49, 0x64, 0x4c, 0x69, 0x73, 0x74,
	0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CS_40001_proto_rawDescOnce sync.Once
	file_CS_40001_proto_rawDescData = file_CS_40001_proto_rawDesc
)

func file_CS_40001_proto_rawDescGZIP() []byte {
	file_CS_40001_proto_rawDescOnce.Do(func() {
		file_CS_40001_proto_rawDescData = protoimpl.X.CompressGZIP(file_CS_40001_proto_rawDescData)
	})
	return file_CS_40001_proto_rawDescData
}

var file_CS_40001_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CS_40001_proto_goTypes = []any{
	(*CS_40001)(nil),    // 0: belfast.CS_40001
	(*OTHERSHIPID)(nil), // 1: belfast.OTHERSHIPID
}
var file_CS_40001_proto_depIdxs = []int32{
	1, // 0: belfast.CS_40001.other_ship_id_list:type_name -> belfast.OTHERSHIPID
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_CS_40001_proto_init() }
func file_CS_40001_proto_init() {
	if File_CS_40001_proto != nil {
		return
	}
	file_OTHERSHIPID_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_CS_40001_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CS_40001); i {
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
			RawDescriptor: file_CS_40001_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CS_40001_proto_goTypes,
		DependencyIndexes: file_CS_40001_proto_depIdxs,
		MessageInfos:      file_CS_40001_proto_msgTypes,
	}.Build()
	File_CS_40001_proto = out.File
	file_CS_40001_proto_rawDesc = nil
	file_CS_40001_proto_goTypes = nil
	file_CS_40001_proto_depIdxs = nil
}
