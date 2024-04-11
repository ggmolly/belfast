// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.25.3
// source: CS_19129.proto

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

type CS_19129 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TargetId   *uint32 `protobuf:"varint,1,req,name=target_id,json=targetId" json:"target_id,omitempty"`
	TargetName *string `protobuf:"bytes,2,req,name=target_name,json=targetName" json:"target_name,omitempty"`
	ThemeId    *string `protobuf:"bytes,3,req,name=theme_id,json=themeId" json:"theme_id,omitempty"`
	ThemeName  *string `protobuf:"bytes,4,req,name=theme_name,json=themeName" json:"theme_name,omitempty"`
	Reason     *uint32 `protobuf:"varint,5,req,name=reason" json:"reason,omitempty"`
}

func (x *CS_19129) Reset() {
	*x = CS_19129{}
	if protoimpl.UnsafeEnabled {
		mi := &file_CS_19129_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CS_19129) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CS_19129) ProtoMessage() {}

func (x *CS_19129) ProtoReflect() protoreflect.Message {
	mi := &file_CS_19129_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CS_19129.ProtoReflect.Descriptor instead.
func (*CS_19129) Descriptor() ([]byte, []int) {
	return file_CS_19129_proto_rawDescGZIP(), []int{0}
}

func (x *CS_19129) GetTargetId() uint32 {
	if x != nil && x.TargetId != nil {
		return *x.TargetId
	}
	return 0
}

func (x *CS_19129) GetTargetName() string {
	if x != nil && x.TargetName != nil {
		return *x.TargetName
	}
	return ""
}

func (x *CS_19129) GetThemeId() string {
	if x != nil && x.ThemeId != nil {
		return *x.ThemeId
	}
	return ""
}

func (x *CS_19129) GetThemeName() string {
	if x != nil && x.ThemeName != nil {
		return *x.ThemeName
	}
	return ""
}

func (x *CS_19129) GetReason() uint32 {
	if x != nil && x.Reason != nil {
		return *x.Reason
	}
	return 0
}

var File_CS_19129_proto protoreflect.FileDescriptor

var file_CS_19129_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x43, 0x53, 0x5f, 0x31, 0x39, 0x31, 0x32, 0x39, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x9a, 0x01, 0x0a, 0x08, 0x43, 0x53,
	0x5f, 0x31, 0x39, 0x31, 0x32, 0x39, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x68, 0x65, 0x6d, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x02, 0x28, 0x09, 0x52, 0x07, 0x74, 0x68, 0x65, 0x6d, 0x65, 0x49, 0x64, 0x12,
	0x1d, 0x0a, 0x0a, 0x74, 0x68, 0x65, 0x6d, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x02, 0x28, 0x09, 0x52, 0x09, 0x74, 0x68, 0x65, 0x6d, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06,
	0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66,
}

var (
	file_CS_19129_proto_rawDescOnce sync.Once
	file_CS_19129_proto_rawDescData = file_CS_19129_proto_rawDesc
)

func file_CS_19129_proto_rawDescGZIP() []byte {
	file_CS_19129_proto_rawDescOnce.Do(func() {
		file_CS_19129_proto_rawDescData = protoimpl.X.CompressGZIP(file_CS_19129_proto_rawDescData)
	})
	return file_CS_19129_proto_rawDescData
}

var file_CS_19129_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_CS_19129_proto_goTypes = []interface{}{
	(*CS_19129)(nil), // 0: belfast.CS_19129
}
var file_CS_19129_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_CS_19129_proto_init() }
func file_CS_19129_proto_init() {
	if File_CS_19129_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_CS_19129_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CS_19129); i {
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
			RawDescriptor: file_CS_19129_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_CS_19129_proto_goTypes,
		DependencyIndexes: file_CS_19129_proto_depIdxs,
		MessageInfos:      file_CS_19129_proto_msgTypes,
	}.Build()
	File_CS_19129_proto = out.File
	file_CS_19129_proto_rawDesc = nil
	file_CS_19129_proto_goTypes = nil
	file_CS_19129_proto_depIdxs = nil
}