// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: REPORT.proto

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

type REPORT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        *uint32        `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	EventId   *uint32        `protobuf:"varint,2,req,name=event_id,json=eventId" json:"event_id,omitempty"`
	EventType *uint32        `protobuf:"varint,3,req,name=event_type,json=eventType" json:"event_type,omitempty"`
	Score     *uint32        `protobuf:"varint,4,req,name=score" json:"score,omitempty"`
	Nodes     []*REPORT_NODE `protobuf:"bytes,5,rep,name=nodes" json:"nodes,omitempty"`
	Status    *uint32        `protobuf:"varint,6,req,name=status" json:"status,omitempty"`
}

func (x *REPORT) Reset() {
	*x = REPORT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_REPORT_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *REPORT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*REPORT) ProtoMessage() {}

func (x *REPORT) ProtoReflect() protoreflect.Message {
	mi := &file_REPORT_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use REPORT.ProtoReflect.Descriptor instead.
func (*REPORT) Descriptor() ([]byte, []int) {
	return file_REPORT_proto_rawDescGZIP(), []int{0}
}

func (x *REPORT) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *REPORT) GetEventId() uint32 {
	if x != nil && x.EventId != nil {
		return *x.EventId
	}
	return 0
}

func (x *REPORT) GetEventType() uint32 {
	if x != nil && x.EventType != nil {
		return *x.EventType
	}
	return 0
}

func (x *REPORT) GetScore() uint32 {
	if x != nil && x.Score != nil {
		return *x.Score
	}
	return 0
}

func (x *REPORT) GetNodes() []*REPORT_NODE {
	if x != nil {
		return x.Nodes
	}
	return nil
}

func (x *REPORT) GetStatus() uint32 {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return 0
}

var File_REPORT_proto protoreflect.FileDescriptor

var file_REPORT_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x52, 0x45, 0x50, 0x4f, 0x52, 0x54, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07,
	0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x11, 0x52, 0x45, 0x50, 0x4f, 0x52, 0x54, 0x5f,
	0x4e, 0x4f, 0x44, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xac, 0x01, 0x0a, 0x06, 0x52,
	0x45, 0x50, 0x4f, 0x52, 0x54, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64,
	0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05,
	0x73, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x2a, 0x0a, 0x05, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x52,
	0x45, 0x50, 0x4f, 0x52, 0x54, 0x5f, 0x4e, 0x4f, 0x44, 0x45, 0x52, 0x05, 0x6e, 0x6f, 0x64, 0x65,
	0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
}

var (
	file_REPORT_proto_rawDescOnce sync.Once
	file_REPORT_proto_rawDescData = file_REPORT_proto_rawDesc
)

func file_REPORT_proto_rawDescGZIP() []byte {
	file_REPORT_proto_rawDescOnce.Do(func() {
		file_REPORT_proto_rawDescData = protoimpl.X.CompressGZIP(file_REPORT_proto_rawDescData)
	})
	return file_REPORT_proto_rawDescData
}

var file_REPORT_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_REPORT_proto_goTypes = []any{
	(*REPORT)(nil),      // 0: belfast.REPORT
	(*REPORT_NODE)(nil), // 1: belfast.REPORT_NODE
}
var file_REPORT_proto_depIdxs = []int32{
	1, // 0: belfast.REPORT.nodes:type_name -> belfast.REPORT_NODE
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_REPORT_proto_init() }
func file_REPORT_proto_init() {
	if File_REPORT_proto != nil {
		return
	}
	file_REPORT_NODE_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_REPORT_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*REPORT); i {
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
			RawDescriptor: file_REPORT_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_REPORT_proto_goTypes,
		DependencyIndexes: file_REPORT_proto_depIdxs,
		MessageInfos:      file_REPORT_proto_msgTypes,
	}.Build()
	File_REPORT_proto = out.File
	file_REPORT_proto_rawDesc = nil
	file_REPORT_proto_goTypes = nil
	file_REPORT_proto_depIdxs = nil
}
