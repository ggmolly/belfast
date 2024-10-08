// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: CS_40003.proto

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

type CS_40003 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	System          *uint32           `protobuf:"varint,1,req,name=system" json:"system,omitempty"`
	Data            *uint32           `protobuf:"varint,2,req,name=data" json:"data,omitempty"`
	Key             *uint32           `protobuf:"varint,3,req,name=key" json:"key,omitempty"`
	Score           *uint32           `protobuf:"varint,4,opt,name=score" json:"score,omitempty"`
	Statistics      []*STATISTICSINFO `protobuf:"bytes,5,rep,name=statistics" json:"statistics,omitempty"`
	KillIdList      []uint32          `protobuf:"varint,6,rep,name=kill_id_list,json=killIdList" json:"kill_id_list,omitempty"`
	TotalTime       *uint32           `protobuf:"varint,7,req,name=total_time,json=totalTime" json:"total_time,omitempty"`
	BotPercentage   *uint32           `protobuf:"varint,8,req,name=bot_percentage,json=botPercentage" json:"bot_percentage,omitempty"`
	ExtraParam      *uint32           `protobuf:"varint,9,req,name=extra_param,json=extraParam" json:"extra_param,omitempty"`
	FileCheck       *string           `protobuf:"bytes,10,opt,name=file_check,json=fileCheck" json:"file_check,omitempty"`
	BossHp          *uint32           `protobuf:"varint,11,opt,name=boss_hp,json=bossHp" json:"boss_hp,omitempty"`
	EnemyInfo       []*ENEMYINFO      `protobuf:"bytes,12,rep,name=enemy_info,json=enemyInfo" json:"enemy_info,omitempty"`
	Data2           []uint32          `protobuf:"varint,13,rep,name=data2" json:"data2,omitempty"`
	CommanderIdList []uint32          `protobuf:"varint,14,rep,name=commander_id_list,json=commanderIdList" json:"commander_id_list,omitempty"`
	Otherstatistics []*STATISTICSINFO `protobuf:"bytes,15,rep,name=otherstatistics" json:"otherstatistics,omitempty"`
	AutoBefore      *uint32           `protobuf:"varint,16,req,name=auto_before,json=autoBefore" json:"auto_before,omitempty"`
	AutoSwitchTime  *uint32           `protobuf:"varint,17,req,name=auto_switch_time,json=autoSwitchTime" json:"auto_switch_time,omitempty"`
	AutoAfter       *uint32           `protobuf:"varint,18,req,name=auto_after,json=autoAfter" json:"auto_after,omitempty"`
}

func (x *CS_40003) Reset() {
	*x = CS_40003{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CS_40003_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CS_40003) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CS_40003) ProtoMessage() {}

func (x *CS_40003) ProtoReflect() protoreflect.Message {
	mi := &file_CS_40003_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CS_40003.ProtoReflect.Descriptor instead.
func (*CS_40003) Descriptor() ([]byte, []int) {
	return file_CS_40003_proto_rawDescGZIP(), []int{0}
}

func (x *CS_40003) GetSystem() uint32 {
	if x != nil && x.System != nil {
		return *x.System
	}
	return 0
}

func (x *CS_40003) GetData() uint32 {
	if x != nil && x.Data != nil {
		return *x.Data
	}
	return 0
}

func (x *CS_40003) GetKey() uint32 {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return 0
}

func (x *CS_40003) GetScore() uint32 {
	if x != nil && x.Score != nil {
		return *x.Score
	}
	return 0
}

func (x *CS_40003) GetStatistics() []*STATISTICSINFO {
	if x != nil {
		return x.Statistics
	}
	return nil
}

func (x *CS_40003) GetKillIdList() []uint32 {
	if x != nil {
		return x.KillIdList
	}
	return nil
}

func (x *CS_40003) GetTotalTime() uint32 {
	if x != nil && x.TotalTime != nil {
		return *x.TotalTime
	}
	return 0
}

func (x *CS_40003) GetBotPercentage() uint32 {
	if x != nil && x.BotPercentage != nil {
		return *x.BotPercentage
	}
	return 0
}

func (x *CS_40003) GetExtraParam() uint32 {
	if x != nil && x.ExtraParam != nil {
		return *x.ExtraParam
	}
	return 0
}

func (x *CS_40003) GetFileCheck() string {
	if x != nil && x.FileCheck != nil {
		return *x.FileCheck
	}
	return ""
}

func (x *CS_40003) GetBossHp() uint32 {
	if x != nil && x.BossHp != nil {
		return *x.BossHp
	}
	return 0
}

func (x *CS_40003) GetEnemyInfo() []*ENEMYINFO {
	if x != nil {
		return x.EnemyInfo
	}
	return nil
}

func (x *CS_40003) GetData2() []uint32 {
	if x != nil {
		return x.Data2
	}
	return nil
}

func (x *CS_40003) GetCommanderIdList() []uint32 {
	if x != nil {
		return x.CommanderIdList
	}
	return nil
}

func (x *CS_40003) GetOtherstatistics() []*STATISTICSINFO {
	if x != nil {
		return x.Otherstatistics
	}
	return nil
}

func (x *CS_40003) GetAutoBefore() uint32 {
	if x != nil && x.AutoBefore != nil {
		return *x.AutoBefore
	}
	return 0
}

func (x *CS_40003) GetAutoSwitchTime() uint32 {
	if x != nil && x.AutoSwitchTime != nil {
		return *x.AutoSwitchTime
	}
	return 0
}

func (x *CS_40003) GetAutoAfter() uint32 {
	if x != nil && x.AutoAfter != nil {
		return *x.AutoAfter
	}
	return 0
}

var File_CS_40003_proto protoreflect.FileDescriptor

var file_CS_40003_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x43, 0x53, 0x5f, 0x34, 0x30, 0x30, 0x30, 0x33, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x14, 0x53, 0x54, 0x41, 0x54, 0x49,
	0x53, 0x54, 0x49, 0x43, 0x53, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0f, 0x45, 0x4e, 0x45, 0x4d, 0x59, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xfa, 0x04, 0x0a, 0x08, 0x43, 0x53, 0x5f, 0x34, 0x30, 0x30, 0x30, 0x33, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x73,
	0x63, 0x6f, 0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x73, 0x63, 0x6f, 0x72,
	0x65, 0x12, 0x37, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x18,
	0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e,
	0x53, 0x54, 0x41, 0x54, 0x49, 0x53, 0x54, 0x49, 0x43, 0x53, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0a,
	0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73, 0x12, 0x20, 0x0a, 0x0c, 0x6b, 0x69,
	0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0d,
	0x52, 0x0a, 0x6b, 0x69, 0x6c, 0x6c, 0x49, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x62,
	0x6f, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x18, 0x08, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x0d, 0x62, 0x6f, 0x74, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61,
	0x67, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x78, 0x74, 0x72, 0x61, 0x5f, 0x70, 0x61, 0x72, 0x61,
	0x6d, 0x18, 0x09, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x78, 0x74, 0x72, 0x61, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x12, 0x1d, 0x0a, 0x0a, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x6f, 0x73, 0x73, 0x5f, 0x68, 0x70, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x06, 0x62, 0x6f, 0x73, 0x73, 0x48, 0x70, 0x12, 0x31, 0x0a, 0x0a, 0x65,
	0x6e, 0x65, 0x6d, 0x79, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x45, 0x4e, 0x45, 0x4d, 0x59, 0x49,
	0x4e, 0x46, 0x4f, 0x52, 0x09, 0x65, 0x6e, 0x65, 0x6d, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14,
	0x0a, 0x05, 0x64, 0x61, 0x74, 0x61, 0x32, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x05, 0x64,
	0x61, 0x74, 0x61, 0x32, 0x12, 0x2a, 0x0a, 0x11, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0d, 0x52,
	0x0f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x65, 0x72, 0x49, 0x64, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x41, 0x0a, 0x0f, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74,
	0x69, 0x63, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66,
	0x61, 0x73, 0x74, 0x2e, 0x53, 0x54, 0x41, 0x54, 0x49, 0x53, 0x54, 0x49, 0x43, 0x53, 0x49, 0x4e,
	0x46, 0x4f, 0x52, 0x0f, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x73, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74,
	0x69, 0x63, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x62, 0x65, 0x66, 0x6f,
	0x72, 0x65, 0x18, 0x10, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x61, 0x75, 0x74, 0x6f, 0x42, 0x65,
	0x66, 0x6f, 0x72, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x73, 0x77, 0x69,
	0x74, 0x63, 0x68, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x11, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0e,
	0x61, 0x75, 0x74, 0x6f, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x61, 0x66, 0x74, 0x65, 0x72, 0x18, 0x12, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x09, 0x61, 0x75, 0x74, 0x6f, 0x41, 0x66, 0x74, 0x65, 0x72, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CS_40003_proto_rawDescOnce sync.Once
	file_CS_40003_proto_rawDescData = file_CS_40003_proto_rawDesc
)

func file_CS_40003_proto_rawDescGZIP() []byte {
	file_CS_40003_proto_rawDescOnce.Do(func() {
		file_CS_40003_proto_rawDescData = protoimpl.X.CompressGZIP(file_CS_40003_proto_rawDescData)
	})
	return file_CS_40003_proto_rawDescData
}

var file_CS_40003_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CS_40003_proto_goTypes = []any{
	(*CS_40003)(nil),       // 0: belfast.CS_40003
	(*STATISTICSINFO)(nil), // 1: belfast.STATISTICSINFO
	(*ENEMYINFO)(nil),      // 2: belfast.ENEMYINFO
}
var file_CS_40003_proto_depIdxs = []int32{
	1, // 0: belfast.CS_40003.statistics:type_name -> belfast.STATISTICSINFO
	2, // 1: belfast.CS_40003.enemy_info:type_name -> belfast.ENEMYINFO
	1, // 2: belfast.CS_40003.otherstatistics:type_name -> belfast.STATISTICSINFO
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_CS_40003_proto_init() }
func file_CS_40003_proto_init() {
	if File_CS_40003_proto != nil {
		return
	}
	file_STATISTICSINFO_proto_init()
	file_ENEMYINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_CS_40003_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*CS_40003); i {
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
			RawDescriptor: file_CS_40003_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CS_40003_proto_goTypes,
		DependencyIndexes: file_CS_40003_proto_depIdxs,
		MessageInfos:      file_CS_40003_proto_msgTypes,
	}.Build()
	File_CS_40003_proto = out.File
	file_CS_40003_proto_rawDesc = nil
	file_CS_40003_proto_goTypes = nil
	file_CS_40003_proto_depIdxs = nil
}
