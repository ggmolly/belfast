// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SPWEAPONINFO.proto

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

type SPWEAPONINFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         *uint32 `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	TemplateId *uint32 `protobuf:"varint,2,req,name=template_id,json=templateId" json:"template_id,omitempty"`
	Attr_1     *uint32 `protobuf:"varint,3,req,name=attr_1,json=attr1" json:"attr_1,omitempty"`
	Attr_2     *uint32 `protobuf:"varint,4,req,name=attr_2,json=attr2" json:"attr_2,omitempty"`
	AttrTemp_1 *uint32 `protobuf:"varint,5,req,name=attr_temp_1,json=attrTemp1" json:"attr_temp_1,omitempty"`
	AttrTemp_2 *uint32 `protobuf:"varint,6,req,name=attr_temp_2,json=attrTemp2" json:"attr_temp_2,omitempty"`
	Effect     *uint32 `protobuf:"varint,7,req,name=effect" json:"effect,omitempty"`
	Pt         *uint32 `protobuf:"varint,8,req,name=pt" json:"pt,omitempty"`
}

func (x *SPWEAPONINFO) Reset() {
	*x = SPWEAPONINFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SPWEAPONINFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SPWEAPONINFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SPWEAPONINFO) ProtoMessage() {}

func (x *SPWEAPONINFO) ProtoReflect() protoreflect.Message {
	mi := &file_SPWEAPONINFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SPWEAPONINFO.ProtoReflect.Descriptor instead.
func (*SPWEAPONINFO) Descriptor() ([]byte, []int) {
	return file_SPWEAPONINFO_proto_rawDescGZIP(), []int{0}
}

func (x *SPWEAPONINFO) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *SPWEAPONINFO) GetTemplateId() uint32 {
	if x != nil && x.TemplateId != nil {
		return *x.TemplateId
	}
	return 0
}

func (x *SPWEAPONINFO) GetAttr_1() uint32 {
	if x != nil && x.Attr_1 != nil {
		return *x.Attr_1
	}
	return 0
}

func (x *SPWEAPONINFO) GetAttr_2() uint32 {
	if x != nil && x.Attr_2 != nil {
		return *x.Attr_2
	}
	return 0
}

func (x *SPWEAPONINFO) GetAttrTemp_1() uint32 {
	if x != nil && x.AttrTemp_1 != nil {
		return *x.AttrTemp_1
	}
	return 0
}

func (x *SPWEAPONINFO) GetAttrTemp_2() uint32 {
	if x != nil && x.AttrTemp_2 != nil {
		return *x.AttrTemp_2
	}
	return 0
}

func (x *SPWEAPONINFO) GetEffect() uint32 {
	if x != nil && x.Effect != nil {
		return *x.Effect
	}
	return 0
}

func (x *SPWEAPONINFO) GetPt() uint32 {
	if x != nil && x.Pt != nil {
		return *x.Pt
	}
	return 0
}

var File_SPWEAPONINFO_proto protoreflect.FileDescriptor

var file_SPWEAPONINFO_proto_rawDesc = []byte{
	0x0a, 0x12, 0x53, 0x50, 0x57, 0x45, 0x41, 0x50, 0x4f, 0x4e, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0xd5, 0x01,
	0x0a, 0x0c, 0x53, 0x50, 0x57, 0x45, 0x41, 0x50, 0x4f, 0x4e, 0x49, 0x4e, 0x46, 0x4f, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1f,
	0x0a, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x0a, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x64, 0x12,
	0x15, 0x0a, 0x06, 0x61, 0x74, 0x74, 0x72, 0x5f, 0x31, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x05, 0x61, 0x74, 0x74, 0x72, 0x31, 0x12, 0x15, 0x0a, 0x06, 0x61, 0x74, 0x74, 0x72, 0x5f, 0x32,
	0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x61, 0x74, 0x74, 0x72, 0x32, 0x12, 0x1e, 0x0a,
	0x0b, 0x61, 0x74, 0x74, 0x72, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x31, 0x18, 0x05, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x09, 0x61, 0x74, 0x74, 0x72, 0x54, 0x65, 0x6d, 0x70, 0x31, 0x12, 0x1e, 0x0a,
	0x0b, 0x61, 0x74, 0x74, 0x72, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x32, 0x18, 0x06, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x09, 0x61, 0x74, 0x74, 0x72, 0x54, 0x65, 0x6d, 0x70, 0x32, 0x12, 0x16, 0x0a,
	0x06, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x18, 0x07, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x65,
	0x66, 0x66, 0x65, 0x63, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x70, 0x74, 0x18, 0x08, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x02, 0x70, 0x74, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_SPWEAPONINFO_proto_rawDescOnce sync.Once
	file_SPWEAPONINFO_proto_rawDescData = file_SPWEAPONINFO_proto_rawDesc
)

func file_SPWEAPONINFO_proto_rawDescGZIP() []byte {
	file_SPWEAPONINFO_proto_rawDescOnce.Do(func() {
		file_SPWEAPONINFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_SPWEAPONINFO_proto_rawDescData)
	})
	return file_SPWEAPONINFO_proto_rawDescData
}

var file_SPWEAPONINFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SPWEAPONINFO_proto_goTypes = []any{
	(*SPWEAPONINFO)(nil), // 0: belfast.SPWEAPONINFO
}
var file_SPWEAPONINFO_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SPWEAPONINFO_proto_init() }
func file_SPWEAPONINFO_proto_init() {
	if File_SPWEAPONINFO_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SPWEAPONINFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SPWEAPONINFO); i {
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
			RawDescriptor: file_SPWEAPONINFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SPWEAPONINFO_proto_goTypes,
		DependencyIndexes: file_SPWEAPONINFO_proto_depIdxs,
		MessageInfos:      file_SPWEAPONINFO_proto_msgTypes,
	}.Build()
	File_SPWEAPONINFO_proto = out.File
	file_SPWEAPONINFO_proto_rawDesc = nil
	file_SPWEAPONINFO_proto_goTypes = nil
	file_SPWEAPONINFO_proto_depIdxs = nil
}
