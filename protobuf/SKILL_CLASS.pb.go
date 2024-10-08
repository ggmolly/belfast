// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SKILL_CLASS.proto

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

type SKILL_CLASS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RoomId     *uint32 `protobuf:"varint,1,req,name=room_id,json=roomId" json:"room_id,omitempty"`
	ShipId     *uint32 `protobuf:"varint,2,req,name=ship_id,json=shipId" json:"ship_id,omitempty"`
	StartTime  *uint32 `protobuf:"varint,3,req,name=start_time,json=startTime" json:"start_time,omitempty"`
	FinishTime *uint32 `protobuf:"varint,4,req,name=finish_time,json=finishTime" json:"finish_time,omitempty"`
	SkillPos   *uint32 `protobuf:"varint,5,req,name=skill_pos,json=skillPos" json:"skill_pos,omitempty"`
	Exp        *uint32 `protobuf:"varint,6,req,name=exp" json:"exp,omitempty"`
}

func (x *SKILL_CLASS) Reset() {
	*x = SKILL_CLASS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SKILL_CLASS_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SKILL_CLASS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SKILL_CLASS) ProtoMessage() {}

func (x *SKILL_CLASS) ProtoReflect() protoreflect.Message {
	mi := &file_SKILL_CLASS_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SKILL_CLASS.ProtoReflect.Descriptor instead.
func (*SKILL_CLASS) Descriptor() ([]byte, []int) {
	return file_SKILL_CLASS_proto_rawDescGZIP(), []int{0}
}

func (x *SKILL_CLASS) GetRoomId() uint32 {
	if x != nil && x.RoomId != nil {
		return *x.RoomId
	}
	return 0
}

func (x *SKILL_CLASS) GetShipId() uint32 {
	if x != nil && x.ShipId != nil {
		return *x.ShipId
	}
	return 0
}

func (x *SKILL_CLASS) GetStartTime() uint32 {
	if x != nil && x.StartTime != nil {
		return *x.StartTime
	}
	return 0
}

func (x *SKILL_CLASS) GetFinishTime() uint32 {
	if x != nil && x.FinishTime != nil {
		return *x.FinishTime
	}
	return 0
}

func (x *SKILL_CLASS) GetSkillPos() uint32 {
	if x != nil && x.SkillPos != nil {
		return *x.SkillPos
	}
	return 0
}

func (x *SKILL_CLASS) GetExp() uint32 {
	if x != nil && x.Exp != nil {
		return *x.Exp
	}
	return 0
}

var File_SKILL_CLASS_proto protoreflect.FileDescriptor

var file_SKILL_CLASS_proto_rawDesc = []byte{
	0x0a, 0x11, 0x53, 0x4b, 0x49, 0x4c, 0x4c, 0x5f, 0x43, 0x4c, 0x41, 0x53, 0x53, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0xae, 0x01, 0x0a,
	0x0b, 0x53, 0x4b, 0x49, 0x4c, 0x4c, 0x5f, 0x43, 0x4c, 0x41, 0x53, 0x53, 0x12, 0x17, 0x0a, 0x07,
	0x72, 0x6f, 0x6f, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x72,
	0x6f, 0x6f, 0x6d, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x68, 0x69, 0x70, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x68, 0x69, 0x70, 0x49, 0x64, 0x12, 0x1d,
	0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1f, 0x0a,
	0x0b, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x0a, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1b,
	0x0a, 0x09, 0x73, 0x6b, 0x69, 0x6c, 0x6c, 0x5f, 0x70, 0x6f, 0x73, 0x18, 0x05, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x08, 0x73, 0x6b, 0x69, 0x6c, 0x6c, 0x50, 0x6f, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x65,
	0x78, 0x70, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x78, 0x70, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_SKILL_CLASS_proto_rawDescOnce sync.Once
	file_SKILL_CLASS_proto_rawDescData = file_SKILL_CLASS_proto_rawDesc
)

func file_SKILL_CLASS_proto_rawDescGZIP() []byte {
	file_SKILL_CLASS_proto_rawDescOnce.Do(func() {
		file_SKILL_CLASS_proto_rawDescData = protoimpl.X.CompressGZIP(file_SKILL_CLASS_proto_rawDescData)
	})
	return file_SKILL_CLASS_proto_rawDescData
}

var file_SKILL_CLASS_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SKILL_CLASS_proto_goTypes = []any{
	(*SKILL_CLASS)(nil), // 0: belfast.SKILL_CLASS
}
var file_SKILL_CLASS_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SKILL_CLASS_proto_init() }
func file_SKILL_CLASS_proto_init() {
	if File_SKILL_CLASS_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SKILL_CLASS_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SKILL_CLASS); i {
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
			RawDescriptor: file_SKILL_CLASS_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SKILL_CLASS_proto_goTypes,
		DependencyIndexes: file_SKILL_CLASS_proto_depIdxs,
		MessageInfos:      file_SKILL_CLASS_proto_msgTypes,
	}.Build()
	File_SKILL_CLASS_proto = out.File
	file_SKILL_CLASS_proto_rawDesc = nil
	file_SKILL_CLASS_proto_goTypes = nil
	file_SKILL_CLASS_proto_depIdxs = nil
}
