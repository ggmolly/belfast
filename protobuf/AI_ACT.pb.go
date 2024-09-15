// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: AI_ACT.proto

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

type AI_ACT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AiPos      *CHAPTERCELLPOS   `protobuf:"bytes,1,req,name=ai_pos,json=aiPos" json:"ai_pos,omitempty"`
	StrategyId *uint32           `protobuf:"varint,2,opt,name=strategy_id,json=strategyId" json:"strategy_id,omitempty"`
	TargetPos  *CHAPTERCELLPOS   `protobuf:"bytes,3,opt,name=target_pos,json=targetPos" json:"target_pos,omitempty"`
	MovePath   []*CHAPTERCELLPOS `protobuf:"bytes,4,rep,name=move_path,json=movePath" json:"move_path,omitempty"`
	ShipUpdate []*SHIPINCHAPTER  `protobuf:"bytes,6,rep,name=ship_update,json=shipUpdate" json:"ship_update,omitempty"`
	Type       *uint32           `protobuf:"varint,7,req,name=type" json:"type,omitempty"`
	PosList    []*WORLDPOSINFO   `protobuf:"bytes,8,rep,name=pos_list,json=posList" json:"pos_list,omitempty"`
}

func (x *AI_ACT) Reset() {
	*x = AI_ACT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_AI_ACT_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AI_ACT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AI_ACT) ProtoMessage() {}

func (x *AI_ACT) ProtoReflect() protoreflect.Message {
	mi := &file_AI_ACT_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AI_ACT.ProtoReflect.Descriptor instead.
func (*AI_ACT) Descriptor() ([]byte, []int) {
	return file_AI_ACT_proto_rawDescGZIP(), []int{0}
}

func (x *AI_ACT) GetAiPos() *CHAPTERCELLPOS {
	if x != nil {
		return x.AiPos
	}
	return nil
}

func (x *AI_ACT) GetStrategyId() uint32 {
	if x != nil && x.StrategyId != nil {
		return *x.StrategyId
	}
	return 0
}

func (x *AI_ACT) GetTargetPos() *CHAPTERCELLPOS {
	if x != nil {
		return x.TargetPos
	}
	return nil
}

func (x *AI_ACT) GetMovePath() []*CHAPTERCELLPOS {
	if x != nil {
		return x.MovePath
	}
	return nil
}

func (x *AI_ACT) GetShipUpdate() []*SHIPINCHAPTER {
	if x != nil {
		return x.ShipUpdate
	}
	return nil
}

func (x *AI_ACT) GetType() uint32 {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return 0
}

func (x *AI_ACT) GetPosList() []*WORLDPOSINFO {
	if x != nil {
		return x.PosList
	}
	return nil
}

var File_AI_ACT_proto protoreflect.FileDescriptor

var file_AI_ACT_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x41, 0x49, 0x5f, 0x41, 0x43, 0x54, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07,
	0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x14, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52,
	0x43, 0x45, 0x4c, 0x4c, 0x50, 0x4f, 0x53, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x53,
	0x48, 0x49, 0x50, 0x49, 0x4e, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x12, 0x57, 0x4f, 0x52, 0x4c, 0x44, 0x50, 0x4f, 0x53, 0x49, 0x4e, 0x46, 0x4f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x02, 0x0a, 0x06, 0x41, 0x49, 0x5f, 0x41, 0x43,
	0x54, 0x12, 0x2e, 0x0a, 0x06, 0x61, 0x69, 0x5f, 0x70, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x43, 0x48, 0x41, 0x50,
	0x54, 0x45, 0x52, 0x43, 0x45, 0x4c, 0x4c, 0x50, 0x4f, 0x53, 0x52, 0x05, 0x61, 0x69, 0x50, 0x6f,
	0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79,
	0x49, 0x64, 0x12, 0x36, 0x0a, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x70, 0x6f, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74,
	0x2e, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x43, 0x45, 0x4c, 0x4c, 0x50, 0x4f, 0x53, 0x52,
	0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x12, 0x34, 0x0a, 0x09, 0x6d, 0x6f,
	0x76, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x43,
	0x45, 0x4c, 0x4c, 0x50, 0x4f, 0x53, 0x52, 0x08, 0x6d, 0x6f, 0x76, 0x65, 0x50, 0x61, 0x74, 0x68,
	0x12, 0x37, 0x0a, 0x0b, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x18,
	0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e,
	0x53, 0x48, 0x49, 0x50, 0x49, 0x4e, 0x43, 0x48, 0x41, 0x50, 0x54, 0x45, 0x52, 0x52, 0x0a, 0x73,
	0x68, 0x69, 0x70, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x07, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x30, 0x0a,
	0x08, 0x70, 0x6f, 0x73, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x57, 0x4f, 0x52, 0x4c, 0x44, 0x50,
	0x4f, 0x53, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x07, 0x70, 0x6f, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x42,
	0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_AI_ACT_proto_rawDescOnce sync.Once
	file_AI_ACT_proto_rawDescData = file_AI_ACT_proto_rawDesc
)

func file_AI_ACT_proto_rawDescGZIP() []byte {
	file_AI_ACT_proto_rawDescOnce.Do(func() {
		file_AI_ACT_proto_rawDescData = protoimpl.X.CompressGZIP(file_AI_ACT_proto_rawDescData)
	})
	return file_AI_ACT_proto_rawDescData
}

var file_AI_ACT_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_AI_ACT_proto_goTypes = []any{
	(*AI_ACT)(nil),         // 0: belfast.AI_ACT
	(*CHAPTERCELLPOS)(nil), // 1: belfast.CHAPTERCELLPOS
	(*SHIPINCHAPTER)(nil),  // 2: belfast.SHIPINCHAPTER
	(*WORLDPOSINFO)(nil),   // 3: belfast.WORLDPOSINFO
}
var file_AI_ACT_proto_depIdxs = []int32{
	1, // 0: belfast.AI_ACT.ai_pos:type_name -> belfast.CHAPTERCELLPOS
	1, // 1: belfast.AI_ACT.target_pos:type_name -> belfast.CHAPTERCELLPOS
	1, // 2: belfast.AI_ACT.move_path:type_name -> belfast.CHAPTERCELLPOS
	2, // 3: belfast.AI_ACT.ship_update:type_name -> belfast.SHIPINCHAPTER
	3, // 4: belfast.AI_ACT.pos_list:type_name -> belfast.WORLDPOSINFO
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_AI_ACT_proto_init() }
func file_AI_ACT_proto_init() {
	if File_AI_ACT_proto != nil {
		return
	}
	file_CHAPTERCELLPOS_proto_init()
	file_SHIPINCHAPTER_proto_init()
	file_WORLDPOSINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_AI_ACT_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*AI_ACT); i {
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
			RawDescriptor: file_AI_ACT_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_AI_ACT_proto_goTypes,
		DependencyIndexes: file_AI_ACT_proto_depIdxs,
		MessageInfos:      file_AI_ACT_proto_msgTypes,
	}.Build()
	File_AI_ACT_proto = out.File
	file_AI_ACT_proto_rawDesc = nil
	file_AI_ACT_proto_goTypes = nil
	file_AI_ACT_proto_depIdxs = nil
}
