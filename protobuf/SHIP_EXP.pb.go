// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SHIP_EXP.proto

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

type SHIP_EXP struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShipId   *uint32 `protobuf:"varint,1,req,name=ship_id,json=shipId" json:"ship_id,omitempty"`
	Exp      *uint32 `protobuf:"varint,2,req,name=exp" json:"exp,omitempty"`
	Intimacy *uint32 `protobuf:"varint,3,req,name=intimacy" json:"intimacy,omitempty"`
	Energy   *uint32 `protobuf:"varint,4,req,name=energy" json:"energy,omitempty"`
}

func (x *SHIP_EXP) Reset() {
	*x = SHIP_EXP{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SHIP_EXP_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SHIP_EXP) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SHIP_EXP) ProtoMessage() {}

func (x *SHIP_EXP) ProtoReflect() protoreflect.Message {
	mi := &file_SHIP_EXP_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SHIP_EXP.ProtoReflect.Descriptor instead.
func (*SHIP_EXP) Descriptor() ([]byte, []int) {
	return file_SHIP_EXP_proto_rawDescGZIP(), []int{0}
}

func (x *SHIP_EXP) GetShipId() uint32 {
	if x != nil && x.ShipId != nil {
		return *x.ShipId
	}
	return 0
}

func (x *SHIP_EXP) GetExp() uint32 {
	if x != nil && x.Exp != nil {
		return *x.Exp
	}
	return 0
}

func (x *SHIP_EXP) GetIntimacy() uint32 {
	if x != nil && x.Intimacy != nil {
		return *x.Intimacy
	}
	return 0
}

func (x *SHIP_EXP) GetEnergy() uint32 {
	if x != nil && x.Energy != nil {
		return *x.Energy
	}
	return 0
}

var File_SHIP_EXP_proto protoreflect.FileDescriptor

var file_SHIP_EXP_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x48, 0x49, 0x50, 0x5f, 0x45, 0x58, 0x50, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x69, 0x0a, 0x08, 0x53, 0x48, 0x49,
	0x50, 0x5f, 0x45, 0x58, 0x50, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x68, 0x69, 0x70, 0x49, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x65, 0x78, 0x70, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x78, 0x70,
	0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x69, 0x6d, 0x61, 0x63, 0x79, 0x18, 0x03, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x08, 0x69, 0x6e, 0x74, 0x69, 0x6d, 0x61, 0x63, 0x79, 0x12, 0x16, 0x0a, 0x06,
	0x65, 0x6e, 0x65, 0x72, 0x67, 0x79, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x65, 0x6e,
	0x65, 0x72, 0x67, 0x79, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66,
}

var (
	file_SHIP_EXP_proto_rawDescOnce sync.Once
	file_SHIP_EXP_proto_rawDescData = file_SHIP_EXP_proto_rawDesc
)

func file_SHIP_EXP_proto_rawDescGZIP() []byte {
	file_SHIP_EXP_proto_rawDescOnce.Do(func() {
		file_SHIP_EXP_proto_rawDescData = protoimpl.X.CompressGZIP(file_SHIP_EXP_proto_rawDescData)
	})
	return file_SHIP_EXP_proto_rawDescData
}

var file_SHIP_EXP_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SHIP_EXP_proto_goTypes = []any{
	(*SHIP_EXP)(nil), // 0: belfast.SHIP_EXP
}
var file_SHIP_EXP_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SHIP_EXP_proto_init() }
func file_SHIP_EXP_proto_init() {
	if File_SHIP_EXP_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SHIP_EXP_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SHIP_EXP); i {
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
			RawDescriptor: file_SHIP_EXP_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SHIP_EXP_proto_goTypes,
		DependencyIndexes: file_SHIP_EXP_proto_depIdxs,
		MessageInfos:      file_SHIP_EXP_proto_msgTypes,
	}.Build()
	File_SHIP_EXP_proto = out.File
	file_SHIP_EXP_proto_rawDesc = nil
	file_SHIP_EXP_proto_goTypes = nil
	file_SHIP_EXP_proto_depIdxs = nil
}
