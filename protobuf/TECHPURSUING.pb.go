// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: TECHPURSUING.proto

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

type TECHPURSUING struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version   *uint32      `protobuf:"varint,1,req,name=version" json:"version,omitempty"`
	Number    *uint32      `protobuf:"varint,2,req,name=number" json:"number,omitempty"`
	DrNumbers []*DR_NUMBER `protobuf:"bytes,3,rep,name=dr_numbers,json=drNumbers" json:"dr_numbers,omitempty"`
}

func (x *TECHPURSUING) Reset() {
	*x = TECHPURSUING{}
	if protoimpl.UnsafeEnabled {
		mi := &file_TECHPURSUING_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TECHPURSUING) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TECHPURSUING) ProtoMessage() {}

func (x *TECHPURSUING) ProtoReflect() protoreflect.Message {
	mi := &file_TECHPURSUING_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TECHPURSUING.ProtoReflect.Descriptor instead.
func (*TECHPURSUING) Descriptor() ([]byte, []int) {
	return file_TECHPURSUING_proto_rawDescGZIP(), []int{0}
}

func (x *TECHPURSUING) GetVersion() uint32 {
	if x != nil && x.Version != nil {
		return *x.Version
	}
	return 0
}

func (x *TECHPURSUING) GetNumber() uint32 {
	if x != nil && x.Number != nil {
		return *x.Number
	}
	return 0
}

func (x *TECHPURSUING) GetDrNumbers() []*DR_NUMBER {
	if x != nil {
		return x.DrNumbers
	}
	return nil
}

var File_TECHPURSUING_proto protoreflect.FileDescriptor

var file_TECHPURSUING_proto_rawDesc = []byte{
	0x0a, 0x12, 0x54, 0x45, 0x43, 0x48, 0x50, 0x55, 0x52, 0x53, 0x55, 0x49, 0x4e, 0x47, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x0f, 0x44,
	0x52, 0x5f, 0x4e, 0x55, 0x4d, 0x42, 0x45, 0x52, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x73,
	0x0a, 0x0c, 0x54, 0x45, 0x43, 0x48, 0x50, 0x55, 0x52, 0x53, 0x55, 0x49, 0x4e, 0x47, 0x12, 0x18,
	0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x12, 0x31, 0x0a, 0x0a, 0x64, 0x72, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x44,
	0x52, 0x5f, 0x4e, 0x55, 0x4d, 0x42, 0x45, 0x52, 0x52, 0x09, 0x64, 0x72, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66,
}

var (
	file_TECHPURSUING_proto_rawDescOnce sync.Once
	file_TECHPURSUING_proto_rawDescData = file_TECHPURSUING_proto_rawDesc
)

func file_TECHPURSUING_proto_rawDescGZIP() []byte {
	file_TECHPURSUING_proto_rawDescOnce.Do(func() {
		file_TECHPURSUING_proto_rawDescData = protoimpl.X.CompressGZIP(file_TECHPURSUING_proto_rawDescData)
	})
	return file_TECHPURSUING_proto_rawDescData
}

var file_TECHPURSUING_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_TECHPURSUING_proto_goTypes = []any{
	(*TECHPURSUING)(nil), // 0: belfast.TECHPURSUING
	(*DR_NUMBER)(nil),    // 1: belfast.DR_NUMBER
}
var file_TECHPURSUING_proto_depIdxs = []int32{
	1, // 0: belfast.TECHPURSUING.dr_numbers:type_name -> belfast.DR_NUMBER
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_TECHPURSUING_proto_init() }
func file_TECHPURSUING_proto_init() {
	if File_TECHPURSUING_proto != nil {
		return
	}
	file_DR_NUMBER_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_TECHPURSUING_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*TECHPURSUING); i {
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
			RawDescriptor: file_TECHPURSUING_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_TECHPURSUING_proto_goTypes,
		DependencyIndexes: file_TECHPURSUING_proto_depIdxs,
		MessageInfos:      file_TECHPURSUING_proto_msgTypes,
	}.Build()
	File_TECHPURSUING_proto = out.File
	file_TECHPURSUING_proto_rawDesc = nil
	file_TECHPURSUING_proto_goTypes = nil
	file_TECHPURSUING_proto_depIdxs = nil
}
