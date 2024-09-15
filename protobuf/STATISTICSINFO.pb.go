// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: STATISTICSINFO.proto

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

type STATISTICSINFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShipId        *uint32 `protobuf:"varint,1,req,name=ship_id,json=shipId" json:"ship_id,omitempty"`
	DamageCause   *uint32 `protobuf:"varint,2,req,name=damage_cause,json=damageCause" json:"damage_cause,omitempty"`
	DamageCaused  *uint32 `protobuf:"varint,3,req,name=damage_caused,json=damageCaused" json:"damage_caused,omitempty"`
	HpRest        *uint32 `protobuf:"varint,4,req,name=hp_rest,json=hpRest" json:"hp_rest,omitempty"`
	MaxDamageOnce *uint32 `protobuf:"varint,5,req,name=max_damage_once,json=maxDamageOnce" json:"max_damage_once,omitempty"`
	ShipGearScore *uint32 `protobuf:"varint,6,req,name=ship_gear_score,json=shipGearScore" json:"ship_gear_score,omitempty"`
}

func (x *STATISTICSINFO) Reset() {
	*x = STATISTICSINFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_STATISTICSINFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *STATISTICSINFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*STATISTICSINFO) ProtoMessage() {}

func (x *STATISTICSINFO) ProtoReflect() protoreflect.Message {
	mi := &file_STATISTICSINFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use STATISTICSINFO.ProtoReflect.Descriptor instead.
func (*STATISTICSINFO) Descriptor() ([]byte, []int) {
	return file_STATISTICSINFO_proto_rawDescGZIP(), []int{0}
}

func (x *STATISTICSINFO) GetShipId() uint32 {
	if x != nil && x.ShipId != nil {
		return *x.ShipId
	}
	return 0
}

func (x *STATISTICSINFO) GetDamageCause() uint32 {
	if x != nil && x.DamageCause != nil {
		return *x.DamageCause
	}
	return 0
}

func (x *STATISTICSINFO) GetDamageCaused() uint32 {
	if x != nil && x.DamageCaused != nil {
		return *x.DamageCaused
	}
	return 0
}

func (x *STATISTICSINFO) GetHpRest() uint32 {
	if x != nil && x.HpRest != nil {
		return *x.HpRest
	}
	return 0
}

func (x *STATISTICSINFO) GetMaxDamageOnce() uint32 {
	if x != nil && x.MaxDamageOnce != nil {
		return *x.MaxDamageOnce
	}
	return 0
}

func (x *STATISTICSINFO) GetShipGearScore() uint32 {
	if x != nil && x.ShipGearScore != nil {
		return *x.ShipGearScore
	}
	return 0
}

var File_STATISTICSINFO_proto protoreflect.FileDescriptor

var file_STATISTICSINFO_proto_rawDesc = []byte{
	0x0a, 0x14, 0x53, 0x54, 0x41, 0x54, 0x49, 0x53, 0x54, 0x49, 0x43, 0x53, 0x49, 0x4e, 0x46, 0x4f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22,
	0xda, 0x01, 0x0a, 0x0e, 0x53, 0x54, 0x41, 0x54, 0x49, 0x53, 0x54, 0x49, 0x43, 0x53, 0x49, 0x4e,
	0x46, 0x4f, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x68, 0x69, 0x70, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64,
	0x61, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x61, 0x75, 0x73, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x0b, 0x64, 0x61, 0x6d, 0x61, 0x67, 0x65, 0x43, 0x61, 0x75, 0x73, 0x65, 0x12, 0x23,
	0x0a, 0x0d, 0x64, 0x61, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x61, 0x75, 0x73, 0x65, 0x64, 0x18,
	0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0c, 0x64, 0x61, 0x6d, 0x61, 0x67, 0x65, 0x43, 0x61, 0x75,
	0x73, 0x65, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x70, 0x5f, 0x72, 0x65, 0x73, 0x74, 0x18, 0x04,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x68, 0x70, 0x52, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f,
	0x6d, 0x61, 0x78, 0x5f, 0x64, 0x61, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x6f, 0x6e, 0x63, 0x65, 0x18,
	0x05, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0d, 0x6d, 0x61, 0x78, 0x44, 0x61, 0x6d, 0x61, 0x67, 0x65,
	0x4f, 0x6e, 0x63, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x67, 0x65, 0x61,
	0x72, 0x5f, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0d, 0x73,
	0x68, 0x69, 0x70, 0x47, 0x65, 0x61, 0x72, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x42, 0x0c, 0x5a, 0x0a,
	0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_STATISTICSINFO_proto_rawDescOnce sync.Once
	file_STATISTICSINFO_proto_rawDescData = file_STATISTICSINFO_proto_rawDesc
)

func file_STATISTICSINFO_proto_rawDescGZIP() []byte {
	file_STATISTICSINFO_proto_rawDescOnce.Do(func() {
		file_STATISTICSINFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_STATISTICSINFO_proto_rawDescData)
	})
	return file_STATISTICSINFO_proto_rawDescData
}

var file_STATISTICSINFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_STATISTICSINFO_proto_goTypes = []any{
	(*STATISTICSINFO)(nil), // 0: belfast.STATISTICSINFO
}
var file_STATISTICSINFO_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_STATISTICSINFO_proto_init() }
func file_STATISTICSINFO_proto_init() {
	if File_STATISTICSINFO_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_STATISTICSINFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*STATISTICSINFO); i {
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
			RawDescriptor: file_STATISTICSINFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_STATISTICSINFO_proto_goTypes,
		DependencyIndexes: file_STATISTICSINFO_proto_depIdxs,
		MessageInfos:      file_STATISTICSINFO_proto_msgTypes,
	}.Build()
	File_STATISTICSINFO_proto = out.File
	file_STATISTICSINFO_proto_rawDesc = nil
	file_STATISTICSINFO_proto_goTypes = nil
	file_STATISTICSINFO_proto_depIdxs = nil
}
