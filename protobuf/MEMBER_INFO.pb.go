// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: MEMBER_INFO.proto

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

type MEMBER_INFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Liveness      *uint32      `protobuf:"varint,1,req,name=liveness" json:"liveness,omitempty"`
	Duty          *uint32      `protobuf:"varint,2,req,name=duty" json:"duty,omitempty"`
	Id            *uint32      `protobuf:"varint,3,req,name=id" json:"id,omitempty"`
	Name          *string      `protobuf:"bytes,4,req,name=name" json:"name,omitempty"`
	Lv            *uint32      `protobuf:"varint,5,req,name=lv" json:"lv,omitempty"`
	Adv           *string      `protobuf:"bytes,6,req,name=adv" json:"adv,omitempty"`
	Online        *uint32      `protobuf:"varint,7,req,name=online" json:"online,omitempty"`
	PreOnlineTime *uint32      `protobuf:"varint,8,req,name=pre_online_time,json=preOnlineTime" json:"pre_online_time,omitempty"`
	Display       *DISPLAYINFO `protobuf:"bytes,9,opt,name=display" json:"display,omitempty"`
	JoinTime      *uint32      `protobuf:"varint,12,req,name=join_time,json=joinTime" json:"join_time,omitempty"`
}

func (x *MEMBER_INFO) Reset() {
	*x = MEMBER_INFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_MEMBER_INFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MEMBER_INFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MEMBER_INFO) ProtoMessage() {}

func (x *MEMBER_INFO) ProtoReflect() protoreflect.Message {
	mi := &file_MEMBER_INFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MEMBER_INFO.ProtoReflect.Descriptor instead.
func (*MEMBER_INFO) Descriptor() ([]byte, []int) {
	return file_MEMBER_INFO_proto_rawDescGZIP(), []int{0}
}

func (x *MEMBER_INFO) GetLiveness() uint32 {
	if x != nil && x.Liveness != nil {
		return *x.Liveness
	}
	return 0
}

func (x *MEMBER_INFO) GetDuty() uint32 {
	if x != nil && x.Duty != nil {
		return *x.Duty
	}
	return 0
}

func (x *MEMBER_INFO) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *MEMBER_INFO) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *MEMBER_INFO) GetLv() uint32 {
	if x != nil && x.Lv != nil {
		return *x.Lv
	}
	return 0
}

func (x *MEMBER_INFO) GetAdv() string {
	if x != nil && x.Adv != nil {
		return *x.Adv
	}
	return ""
}

func (x *MEMBER_INFO) GetOnline() uint32 {
	if x != nil && x.Online != nil {
		return *x.Online
	}
	return 0
}

func (x *MEMBER_INFO) GetPreOnlineTime() uint32 {
	if x != nil && x.PreOnlineTime != nil {
		return *x.PreOnlineTime
	}
	return 0
}

func (x *MEMBER_INFO) GetDisplay() *DISPLAYINFO {
	if x != nil {
		return x.Display
	}
	return nil
}

func (x *MEMBER_INFO) GetJoinTime() uint32 {
	if x != nil && x.JoinTime != nil {
		return *x.JoinTime
	}
	return 0
}

var File_MEMBER_INFO_proto protoreflect.FileDescriptor

var file_MEMBER_INFO_proto_rawDesc = []byte{
	0x0a, 0x11, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x11, 0x44, 0x49,
	0x53, 0x50, 0x4c, 0x41, 0x59, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x90, 0x02, 0x0a, 0x0b, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x69, 0x76, 0x65, 0x6e, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x08, 0x6c, 0x69, 0x76, 0x65, 0x6e, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x75, 0x74, 0x79, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x64, 0x75, 0x74, 0x79, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x02, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x6c, 0x76, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x02, 0x6c, 0x76, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x64, 0x76, 0x18, 0x06, 0x20, 0x02, 0x28, 0x09,
	0x52, 0x03, 0x61, 0x64, 0x76, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x18,
	0x07, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x26, 0x0a,
	0x0f, 0x70, 0x72, 0x65, 0x5f, 0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x08, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x4f, 0x6e, 0x6c, 0x69, 0x6e,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74,
	0x2e, 0x44, 0x49, 0x53, 0x50, 0x4c, 0x41, 0x59, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x07, 0x64, 0x69,
	0x73, 0x70, 0x6c, 0x61, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x6a, 0x6f, 0x69, 0x6e, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x0c, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x6a, 0x6f, 0x69, 0x6e, 0x54, 0x69,
	0x6d, 0x65, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_MEMBER_INFO_proto_rawDescOnce sync.Once
	file_MEMBER_INFO_proto_rawDescData = file_MEMBER_INFO_proto_rawDesc
)

func file_MEMBER_INFO_proto_rawDescGZIP() []byte {
	file_MEMBER_INFO_proto_rawDescOnce.Do(func() {
		file_MEMBER_INFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_MEMBER_INFO_proto_rawDescData)
	})
	return file_MEMBER_INFO_proto_rawDescData
}

var file_MEMBER_INFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_MEMBER_INFO_proto_goTypes = []any{
	(*MEMBER_INFO)(nil), // 0: belfast.MEMBER_INFO
	(*DISPLAYINFO)(nil), // 1: belfast.DISPLAYINFO
}
var file_MEMBER_INFO_proto_depIdxs = []int32{
	1, // 0: belfast.MEMBER_INFO.display:type_name -> belfast.DISPLAYINFO
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_MEMBER_INFO_proto_init() }
func file_MEMBER_INFO_proto_init() {
	if File_MEMBER_INFO_proto != nil {
		return
	}
	file_DISPLAYINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_MEMBER_INFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MEMBER_INFO); i {
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
			RawDescriptor: file_MEMBER_INFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_MEMBER_INFO_proto_goTypes,
		DependencyIndexes: file_MEMBER_INFO_proto_depIdxs,
		MessageInfos:      file_MEMBER_INFO_proto_msgTypes,
	}.Build()
	File_MEMBER_INFO_proto = out.File
	file_MEMBER_INFO_proto_rawDesc = nil
	file_MEMBER_INFO_proto_goTypes = nil
	file_MEMBER_INFO_proto_depIdxs = nil
}
