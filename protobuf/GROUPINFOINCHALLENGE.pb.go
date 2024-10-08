// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: GROUPINFOINCHALLENGE.proto

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

type GROUPINFOINCHALLENGE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         *uint32                 `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Ships      []*SHIPINCHALLENGE      `protobuf:"bytes,2,rep,name=ships" json:"ships,omitempty"`
	Commanders []*COMMANDERINCHALLENGE `protobuf:"bytes,3,rep,name=commanders" json:"commanders,omitempty"`
}

func (x *GROUPINFOINCHALLENGE) Reset() {
	*x = GROUPINFOINCHALLENGE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_GROUPINFOINCHALLENGE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GROUPINFOINCHALLENGE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GROUPINFOINCHALLENGE) ProtoMessage() {}

func (x *GROUPINFOINCHALLENGE) ProtoReflect() protoreflect.Message {
	mi := &file_GROUPINFOINCHALLENGE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GROUPINFOINCHALLENGE.ProtoReflect.Descriptor instead.
func (*GROUPINFOINCHALLENGE) Descriptor() ([]byte, []int) {
	return file_GROUPINFOINCHALLENGE_proto_rawDescGZIP(), []int{0}
}

func (x *GROUPINFOINCHALLENGE) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *GROUPINFOINCHALLENGE) GetShips() []*SHIPINCHALLENGE {
	if x != nil {
		return x.Ships
	}
	return nil
}

func (x *GROUPINFOINCHALLENGE) GetCommanders() []*COMMANDERINCHALLENGE {
	if x != nil {
		return x.Commanders
	}
	return nil
}

var File_GROUPINFOINCHALLENGE_proto protoreflect.FileDescriptor

var file_GROUPINFOINCHALLENGE_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x49, 0x4e, 0x43, 0x48, 0x41,
	0x4c, 0x4c, 0x45, 0x4e, 0x47, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x62, 0x65,
	0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x15, 0x53, 0x48, 0x49, 0x50, 0x49, 0x4e, 0x43, 0x48, 0x41,
	0x4c, 0x4c, 0x45, 0x4e, 0x47, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x43, 0x4f,
	0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x45, 0x52, 0x49, 0x4e, 0x43, 0x48, 0x41, 0x4c, 0x4c, 0x45, 0x4e,
	0x47, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x95, 0x01, 0x0a, 0x14, 0x47, 0x52, 0x4f,
	0x55, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x49, 0x4e, 0x43, 0x48, 0x41, 0x4c, 0x4c, 0x45, 0x4e, 0x47,
	0x45, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x2e, 0x0a, 0x05, 0x73, 0x68, 0x69, 0x70, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53, 0x48, 0x49, 0x50, 0x49,
	0x4e, 0x43, 0x48, 0x41, 0x4c, 0x4c, 0x45, 0x4e, 0x47, 0x45, 0x52, 0x05, 0x73, 0x68, 0x69, 0x70,
	0x73, 0x12, 0x3d, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x65, 0x72, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e,
	0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x45, 0x52, 0x49, 0x4e, 0x43, 0x48, 0x41, 0x4c, 0x4c,
	0x45, 0x4e, 0x47, 0x45, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x65, 0x72, 0x73,
	0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_GROUPINFOINCHALLENGE_proto_rawDescOnce sync.Once
	file_GROUPINFOINCHALLENGE_proto_rawDescData = file_GROUPINFOINCHALLENGE_proto_rawDesc
)

func file_GROUPINFOINCHALLENGE_proto_rawDescGZIP() []byte {
	file_GROUPINFOINCHALLENGE_proto_rawDescOnce.Do(func() {
		file_GROUPINFOINCHALLENGE_proto_rawDescData = protoimpl.X.CompressGZIP(file_GROUPINFOINCHALLENGE_proto_rawDescData)
	})
	return file_GROUPINFOINCHALLENGE_proto_rawDescData
}

var file_GROUPINFOINCHALLENGE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_GROUPINFOINCHALLENGE_proto_goTypes = []any{
	(*GROUPINFOINCHALLENGE)(nil), // 0: belfast.GROUPINFOINCHALLENGE
	(*SHIPINCHALLENGE)(nil),      // 1: belfast.SHIPINCHALLENGE
	(*COMMANDERINCHALLENGE)(nil), // 2: belfast.COMMANDERINCHALLENGE
}
var file_GROUPINFOINCHALLENGE_proto_depIdxs = []int32{
	1, // 0: belfast.GROUPINFOINCHALLENGE.ships:type_name -> belfast.SHIPINCHALLENGE
	2, // 1: belfast.GROUPINFOINCHALLENGE.commanders:type_name -> belfast.COMMANDERINCHALLENGE
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_GROUPINFOINCHALLENGE_proto_init() }
func file_GROUPINFOINCHALLENGE_proto_init() {
	if File_GROUPINFOINCHALLENGE_proto != nil {
		return
	}
	file_SHIPINCHALLENGE_proto_init()
	file_COMMANDERINCHALLENGE_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_GROUPINFOINCHALLENGE_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GROUPINFOINCHALLENGE); i {
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
			RawDescriptor: file_GROUPINFOINCHALLENGE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_GROUPINFOINCHALLENGE_proto_goTypes,
		DependencyIndexes: file_GROUPINFOINCHALLENGE_proto_depIdxs,
		MessageInfos:      file_GROUPINFOINCHALLENGE_proto_msgTypes,
	}.Build()
	File_GROUPINFOINCHALLENGE_proto = out.File
	file_GROUPINFOINCHALLENGE_proto_rawDesc = nil
	file_GROUPINFOINCHALLENGE_proto_goTypes = nil
	file_GROUPINFOINCHALLENGE_proto_depIdxs = nil
}
