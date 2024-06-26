// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.25.3
// source: PB_ROGUE.proto

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

type PB_ROGUE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hp           *uint32          `protobuf:"varint,1,req,name=hp" json:"hp,omitempty"`
	Level        *uint32          `protobuf:"varint,2,req,name=level" json:"level,omitempty"`
	Mode         *uint32          `protobuf:"varint,3,req,name=mode" json:"mode,omitempty"`
	Index        *uint32          `protobuf:"varint,4,req,name=index" json:"index,omitempty"`
	Coin         *uint32          `protobuf:"varint,5,req,name=coin" json:"coin,omitempty"`
	CacheBonus   []*DROPINFO      `protobuf:"bytes,6,rep,name=cache_bonus,json=cacheBonus" json:"cache_bonus,omitempty"`
	CardList     []*ROGUECARD     `protobuf:"bytes,7,rep,name=card_list,json=cardList" json:"card_list,omitempty"`
	TreasureList []*ROGUETREASURE `protobuf:"bytes,8,rep,name=treasure_list,json=treasureList" json:"treasure_list,omitempty"`
	Map          *ROGUEMAP        `protobuf:"bytes,9,opt,name=map" json:"map,omitempty"`
	FrontId      *uint32          `protobuf:"varint,10,req,name=front_id,json=frontId" json:"front_id,omitempty"`
	BackId       *uint32          `protobuf:"varint,11,req,name=back_id,json=backId" json:"back_id,omitempty"`
	CacheOil     *uint32          `protobuf:"varint,12,req,name=cache_oil,json=cacheOil" json:"cache_oil,omitempty"`
	Bundles      []*CARDBUNDLE    `protobuf:"bytes,13,rep,name=bundles" json:"bundles,omitempty"`
	Time         *uint32          `protobuf:"varint,14,req,name=time" json:"time,omitempty"`
}

func (x *PB_ROGUE) Reset() {
	*x = PB_ROGUE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_PB_ROGUE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PB_ROGUE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PB_ROGUE) ProtoMessage() {}

func (x *PB_ROGUE) ProtoReflect() protoreflect.Message {
	mi := &file_PB_ROGUE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PB_ROGUE.ProtoReflect.Descriptor instead.
func (*PB_ROGUE) Descriptor() ([]byte, []int) {
	return file_PB_ROGUE_proto_rawDescGZIP(), []int{0}
}

func (x *PB_ROGUE) GetHp() uint32 {
	if x != nil && x.Hp != nil {
		return *x.Hp
	}
	return 0
}

func (x *PB_ROGUE) GetLevel() uint32 {
	if x != nil && x.Level != nil {
		return *x.Level
	}
	return 0
}

func (x *PB_ROGUE) GetMode() uint32 {
	if x != nil && x.Mode != nil {
		return *x.Mode
	}
	return 0
}

func (x *PB_ROGUE) GetIndex() uint32 {
	if x != nil && x.Index != nil {
		return *x.Index
	}
	return 0
}

func (x *PB_ROGUE) GetCoin() uint32 {
	if x != nil && x.Coin != nil {
		return *x.Coin
	}
	return 0
}

func (x *PB_ROGUE) GetCacheBonus() []*DROPINFO {
	if x != nil {
		return x.CacheBonus
	}
	return nil
}

func (x *PB_ROGUE) GetCardList() []*ROGUECARD {
	if x != nil {
		return x.CardList
	}
	return nil
}

func (x *PB_ROGUE) GetTreasureList() []*ROGUETREASURE {
	if x != nil {
		return x.TreasureList
	}
	return nil
}

func (x *PB_ROGUE) GetMap() *ROGUEMAP {
	if x != nil {
		return x.Map
	}
	return nil
}

func (x *PB_ROGUE) GetFrontId() uint32 {
	if x != nil && x.FrontId != nil {
		return *x.FrontId
	}
	return 0
}

func (x *PB_ROGUE) GetBackId() uint32 {
	if x != nil && x.BackId != nil {
		return *x.BackId
	}
	return 0
}

func (x *PB_ROGUE) GetCacheOil() uint32 {
	if x != nil && x.CacheOil != nil {
		return *x.CacheOil
	}
	return 0
}

func (x *PB_ROGUE) GetBundles() []*CARDBUNDLE {
	if x != nil {
		return x.Bundles
	}
	return nil
}

func (x *PB_ROGUE) GetTime() uint32 {
	if x != nil && x.Time != nil {
		return *x.Time
	}
	return 0
}

var File_PB_ROGUE_proto protoreflect.FileDescriptor

var file_PB_ROGUE_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x50, 0x42, 0x5f, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x0e, 0x44, 0x52, 0x4f, 0x50, 0x49,
	0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x52, 0x4f, 0x47, 0x55, 0x45,
	0x43, 0x41, 0x52, 0x44, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x52, 0x4f, 0x47, 0x55,
	0x45, 0x54, 0x52, 0x45, 0x41, 0x53, 0x55, 0x52, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0e, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x4d, 0x41, 0x50, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x10, 0x43, 0x41, 0x52, 0x44, 0x42, 0x55, 0x4e, 0x44, 0x4c, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xc9, 0x03, 0x0a, 0x08, 0x50, 0x42, 0x5f, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x12, 0x0e,
	0x0a, 0x02, 0x68, 0x70, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x68, 0x70, 0x12, 0x14,
	0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x6c,
	0x65, 0x76, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x6f, 0x69, 0x6e, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x63, 0x6f,
	0x69, 0x6e, 0x12, 0x32, 0x0a, 0x0b, 0x63, 0x61, 0x63, 0x68, 0x65, 0x5f, 0x62, 0x6f, 0x6e, 0x75,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x2e, 0x44, 0x52, 0x4f, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0a, 0x63, 0x61, 0x63, 0x68,
	0x65, 0x42, 0x6f, 0x6e, 0x75, 0x73, 0x12, 0x2f, 0x0a, 0x09, 0x63, 0x61, 0x72, 0x64, 0x5f, 0x6c,
	0x69, 0x73, 0x74, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66,
	0x61, 0x73, 0x74, 0x2e, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x43, 0x41, 0x52, 0x44, 0x52, 0x08, 0x63,
	0x61, 0x72, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x3b, 0x0a, 0x0d, 0x74, 0x72, 0x65, 0x61, 0x73,
	0x75, 0x72, 0x65, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16,
	0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x52, 0x4f, 0x47, 0x55, 0x45, 0x54, 0x52,
	0x45, 0x41, 0x53, 0x55, 0x52, 0x45, 0x52, 0x0c, 0x74, 0x72, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x03, 0x6d, 0x61, 0x70, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x52, 0x4f, 0x47, 0x55,
	0x45, 0x4d, 0x41, 0x50, 0x52, 0x03, 0x6d, 0x61, 0x70, 0x12, 0x19, 0x0a, 0x08, 0x66, 0x72, 0x6f,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x07, 0x66, 0x72, 0x6f,
	0x6e, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x69, 0x64, 0x18,
	0x0b, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x62, 0x61, 0x63, 0x6b, 0x49, 0x64, 0x12, 0x1b, 0x0a,
	0x09, 0x63, 0x61, 0x63, 0x68, 0x65, 0x5f, 0x6f, 0x69, 0x6c, 0x18, 0x0c, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x08, 0x63, 0x61, 0x63, 0x68, 0x65, 0x4f, 0x69, 0x6c, 0x12, 0x2d, 0x0a, 0x07, 0x62, 0x75,
	0x6e, 0x64, 0x6c, 0x65, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x65,
	0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x43, 0x41, 0x52, 0x44, 0x42, 0x55, 0x4e, 0x44, 0x4c, 0x45,
	0x52, 0x07, 0x62, 0x75, 0x6e, 0x64, 0x6c, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x0e, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_PB_ROGUE_proto_rawDescOnce sync.Once
	file_PB_ROGUE_proto_rawDescData = file_PB_ROGUE_proto_rawDesc
)

func file_PB_ROGUE_proto_rawDescGZIP() []byte {
	file_PB_ROGUE_proto_rawDescOnce.Do(func() {
		file_PB_ROGUE_proto_rawDescData = protoimpl.X.CompressGZIP(file_PB_ROGUE_proto_rawDescData)
	})
	return file_PB_ROGUE_proto_rawDescData
}

var file_PB_ROGUE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_PB_ROGUE_proto_goTypes = []interface{}{
	(*PB_ROGUE)(nil),      // 0: belfast.PB_ROGUE
	(*DROPINFO)(nil),      // 1: belfast.DROPINFO
	(*ROGUECARD)(nil),     // 2: belfast.ROGUECARD
	(*ROGUETREASURE)(nil), // 3: belfast.ROGUETREASURE
	(*ROGUEMAP)(nil),      // 4: belfast.ROGUEMAP
	(*CARDBUNDLE)(nil),    // 5: belfast.CARDBUNDLE
}
var file_PB_ROGUE_proto_depIdxs = []int32{
	1, // 0: belfast.PB_ROGUE.cache_bonus:type_name -> belfast.DROPINFO
	2, // 1: belfast.PB_ROGUE.card_list:type_name -> belfast.ROGUECARD
	3, // 2: belfast.PB_ROGUE.treasure_list:type_name -> belfast.ROGUETREASURE
	4, // 3: belfast.PB_ROGUE.map:type_name -> belfast.ROGUEMAP
	5, // 4: belfast.PB_ROGUE.bundles:type_name -> belfast.CARDBUNDLE
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_PB_ROGUE_proto_init() }
func file_PB_ROGUE_proto_init() {
	if File_PB_ROGUE_proto != nil {
		return
	}
	file_DROPINFO_proto_init()
	file_ROGUECARD_proto_init()
	file_ROGUETREASURE_proto_init()
	file_ROGUEMAP_proto_init()
	file_CARDBUNDLE_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_PB_ROGUE_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PB_ROGUE); i {
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
			RawDescriptor: file_PB_ROGUE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_PB_ROGUE_proto_goTypes,
		DependencyIndexes: file_PB_ROGUE_proto_depIdxs,
		MessageInfos:      file_PB_ROGUE_proto_msgTypes,
	}.Build()
	File_PB_ROGUE_proto = out.File
	file_PB_ROGUE_proto_rawDesc = nil
	file_PB_ROGUE_proto_goTypes = nil
	file_PB_ROGUE_proto_depIdxs = nil
}
