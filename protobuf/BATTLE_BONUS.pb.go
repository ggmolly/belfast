// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.25.3
// source: BATTLE_BONUS.proto

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

type BATTLE_BONUS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bonus      []*DROPINFO       `protobuf:"bytes,1,rep,name=bonus" json:"bonus,omitempty"`
	CoinBonus  []*ROGUE_DROPINFO `protobuf:"bytes,2,rep,name=coin_bonus,json=coinBonus" json:"coin_bonus,omitempty"`
	Cardlist   []uint32          `protobuf:"varint,3,rep,name=cardlist" json:"cardlist,omitempty"`
	RogueBonus []*ROGUE_DROPINFO `protobuf:"bytes,4,rep,name=rogue_bonus,json=rogueBonus" json:"rogue_bonus,omitempty"`
}

func (x *BATTLE_BONUS) Reset() {
	*x = BATTLE_BONUS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_BATTLE_BONUS_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BATTLE_BONUS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BATTLE_BONUS) ProtoMessage() {}

func (x *BATTLE_BONUS) ProtoReflect() protoreflect.Message {
	mi := &file_BATTLE_BONUS_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BATTLE_BONUS.ProtoReflect.Descriptor instead.
func (*BATTLE_BONUS) Descriptor() ([]byte, []int) {
	return file_BATTLE_BONUS_proto_rawDescGZIP(), []int{0}
}

func (x *BATTLE_BONUS) GetBonus() []*DROPINFO {
	if x != nil {
		return x.Bonus
	}
	return nil
}

func (x *BATTLE_BONUS) GetCoinBonus() []*ROGUE_DROPINFO {
	if x != nil {
		return x.CoinBonus
	}
	return nil
}

func (x *BATTLE_BONUS) GetCardlist() []uint32 {
	if x != nil {
		return x.Cardlist
	}
	return nil
}

func (x *BATTLE_BONUS) GetRogueBonus() []*ROGUE_DROPINFO {
	if x != nil {
		return x.RogueBonus
	}
	return nil
}

var File_BATTLE_BONUS_proto protoreflect.FileDescriptor

var file_BATTLE_BONUS_proto_rawDesc = []byte{
	0x0a, 0x12, 0x42, 0x41, 0x54, 0x54, 0x4c, 0x45, 0x5f, 0x42, 0x4f, 0x4e, 0x55, 0x53, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x0e, 0x44,
	0x52, 0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x52,
	0x4f, 0x47, 0x55, 0x45, 0x5f, 0x44, 0x52, 0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xc5, 0x01, 0x0a, 0x0c, 0x42, 0x41, 0x54, 0x54, 0x4c, 0x45, 0x5f, 0x42,
	0x4f, 0x4e, 0x55, 0x53, 0x12, 0x27, 0x0a, 0x05, 0x62, 0x6f, 0x6e, 0x75, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x44, 0x52,
	0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x05, 0x62, 0x6f, 0x6e, 0x75, 0x73, 0x12, 0x36, 0x0a,
	0x0a, 0x63, 0x6f, 0x69, 0x6e, 0x5f, 0x62, 0x6f, 0x6e, 0x75, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x52, 0x4f, 0x47, 0x55,
	0x45, 0x5f, 0x44, 0x52, 0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x09, 0x63, 0x6f, 0x69, 0x6e,
	0x42, 0x6f, 0x6e, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x72, 0x64, 0x6c, 0x69, 0x73,
	0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x08, 0x63, 0x61, 0x72, 0x64, 0x6c, 0x69, 0x73,
	0x74, 0x12, 0x38, 0x0a, 0x0b, 0x72, 0x6f, 0x67, 0x75, 0x65, 0x5f, 0x62, 0x6f, 0x6e, 0x75, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74,
	0x2e, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x5f, 0x44, 0x52, 0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x52,
	0x0a, 0x72, 0x6f, 0x67, 0x75, 0x65, 0x42, 0x6f, 0x6e, 0x75, 0x73, 0x42, 0x0c, 0x5a, 0x0a, 0x2e,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_BATTLE_BONUS_proto_rawDescOnce sync.Once
	file_BATTLE_BONUS_proto_rawDescData = file_BATTLE_BONUS_proto_rawDesc
)

func file_BATTLE_BONUS_proto_rawDescGZIP() []byte {
	file_BATTLE_BONUS_proto_rawDescOnce.Do(func() {
		file_BATTLE_BONUS_proto_rawDescData = protoimpl.X.CompressGZIP(file_BATTLE_BONUS_proto_rawDescData)
	})
	return file_BATTLE_BONUS_proto_rawDescData
}

var file_BATTLE_BONUS_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_BATTLE_BONUS_proto_goTypes = []interface{}{
	(*BATTLE_BONUS)(nil),   // 0: belfast.BATTLE_BONUS
	(*DROPINFO)(nil),       // 1: belfast.DROPINFO
	(*ROGUE_DROPINFO)(nil), // 2: belfast.ROGUE_DROPINFO
}
var file_BATTLE_BONUS_proto_depIdxs = []int32{
	1, // 0: belfast.BATTLE_BONUS.bonus:type_name -> belfast.DROPINFO
	2, // 1: belfast.BATTLE_BONUS.coin_bonus:type_name -> belfast.ROGUE_DROPINFO
	2, // 2: belfast.BATTLE_BONUS.rogue_bonus:type_name -> belfast.ROGUE_DROPINFO
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_BATTLE_BONUS_proto_init() }
func file_BATTLE_BONUS_proto_init() {
	if File_BATTLE_BONUS_proto != nil {
		return
	}
	file_DROPINFO_proto_init()
	file_ROGUE_DROPINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_BATTLE_BONUS_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BATTLE_BONUS); i {
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
			RawDescriptor: file_BATTLE_BONUS_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_BATTLE_BONUS_proto_goTypes,
		DependencyIndexes: file_BATTLE_BONUS_proto_depIdxs,
		MessageInfos:      file_BATTLE_BONUS_proto_msgTypes,
	}.Build()
	File_BATTLE_BONUS_proto = out.File
	file_BATTLE_BONUS_proto_rawDesc = nil
	file_BATTLE_BONUS_proto_goTypes = nil
	file_BATTLE_BONUS_proto_depIdxs = nil
}
