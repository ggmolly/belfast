// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: EVENT_BASE.proto

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

type EVENT_BASE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId       *uint32                `protobuf:"varint,1,req,name=event_id,json=eventId" json:"event_id,omitempty"`
	Position      *uint32                `protobuf:"varint,2,req,name=position" json:"position,omitempty"`
	StartTime     *uint32                `protobuf:"varint,3,req,name=start_time,json=startTime" json:"start_time,omitempty"`
	CompleteTime  *uint32                `protobuf:"varint,4,req,name=complete_time,json=completeTime" json:"complete_time,omitempty"`
	Shipinevent   []*SHIP_IN_EVENT       `protobuf:"bytes,5,rep,name=shipinevent" json:"shipinevent,omitempty"`
	AttrAccList   []*KEYVALUE            `protobuf:"bytes,6,rep,name=attr_acc_list,json=attrAccList" json:"attr_acc_list,omitempty"`
	AttrCountList []*KEYVALUE            `protobuf:"bytes,7,rep,name=attr_count_list,json=attrCountList" json:"attr_count_list,omitempty"`
	Eventnodes    []*EVENT_NODE          `protobuf:"bytes,8,rep,name=eventnodes" json:"eventnodes,omitempty"`
	Efficiency    *uint32                `protobuf:"varint,9,req,name=efficiency" json:"efficiency,omitempty"`
	Personship    []*PERSON_SHIP_IN_PAGE `protobuf:"bytes,10,rep,name=personship" json:"personship,omitempty"`
}

func (x *EVENT_BASE) Reset() {
	*x = EVENT_BASE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_EVENT_BASE_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EVENT_BASE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EVENT_BASE) ProtoMessage() {}

func (x *EVENT_BASE) ProtoReflect() protoreflect.Message {
	mi := &file_EVENT_BASE_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EVENT_BASE.ProtoReflect.Descriptor instead.
func (*EVENT_BASE) Descriptor() ([]byte, []int) {
	return file_EVENT_BASE_proto_rawDescGZIP(), []int{0}
}

func (x *EVENT_BASE) GetEventId() uint32 {
	if x != nil && x.EventId != nil {
		return *x.EventId
	}
	return 0
}

func (x *EVENT_BASE) GetPosition() uint32 {
	if x != nil && x.Position != nil {
		return *x.Position
	}
	return 0
}

func (x *EVENT_BASE) GetStartTime() uint32 {
	if x != nil && x.StartTime != nil {
		return *x.StartTime
	}
	return 0
}

func (x *EVENT_BASE) GetCompleteTime() uint32 {
	if x != nil && x.CompleteTime != nil {
		return *x.CompleteTime
	}
	return 0
}

func (x *EVENT_BASE) GetShipinevent() []*SHIP_IN_EVENT {
	if x != nil {
		return x.Shipinevent
	}
	return nil
}

func (x *EVENT_BASE) GetAttrAccList() []*KEYVALUE {
	if x != nil {
		return x.AttrAccList
	}
	return nil
}

func (x *EVENT_BASE) GetAttrCountList() []*KEYVALUE {
	if x != nil {
		return x.AttrCountList
	}
	return nil
}

func (x *EVENT_BASE) GetEventnodes() []*EVENT_NODE {
	if x != nil {
		return x.Eventnodes
	}
	return nil
}

func (x *EVENT_BASE) GetEfficiency() uint32 {
	if x != nil && x.Efficiency != nil {
		return *x.Efficiency
	}
	return 0
}

func (x *EVENT_BASE) GetPersonship() []*PERSON_SHIP_IN_PAGE {
	if x != nil {
		return x.Personship
	}
	return nil
}

var File_EVENT_BASE_proto protoreflect.FileDescriptor

var file_EVENT_BASE_proto_rawDesc = []byte{
	0x0a, 0x10, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x42, 0x41, 0x53, 0x45, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x13, 0x53, 0x48, 0x49,
	0x50, 0x5f, 0x49, 0x4e, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x0e, 0x4b, 0x45, 0x59, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x10, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x4e, 0x4f, 0x44, 0x45, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x19, 0x50, 0x45, 0x52, 0x53, 0x4f, 0x4e, 0x5f, 0x53, 0x48, 0x49, 0x50, 0x5f,
	0x49, 0x4e, 0x5f, 0x50, 0x41, 0x47, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x03,
	0x0a, 0x0a, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x42, 0x41, 0x53, 0x45, 0x12, 0x19, 0x0a, 0x08,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x07,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69,
	0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0c, 0x63, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x38, 0x0a, 0x0b, 0x73, 0x68, 0x69, 0x70, 0x69,
	0x6e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x62,
	0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53, 0x48, 0x49, 0x50, 0x5f, 0x49, 0x4e, 0x5f, 0x45,
	0x56, 0x45, 0x4e, 0x54, 0x52, 0x0b, 0x73, 0x68, 0x69, 0x70, 0x69, 0x6e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x35, 0x0a, 0x0d, 0x61, 0x74, 0x74, 0x72, 0x5f, 0x61, 0x63, 0x63, 0x5f, 0x6c, 0x69,
	0x73, 0x74, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61,
	0x73, 0x74, 0x2e, 0x4b, 0x45, 0x59, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x52, 0x0b, 0x61, 0x74, 0x74,
	0x72, 0x41, 0x63, 0x63, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x39, 0x0a, 0x0f, 0x61, 0x74, 0x74, 0x72,
	0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x4b, 0x45, 0x59, 0x56,
	0x41, 0x4c, 0x55, 0x45, 0x52, 0x0d, 0x61, 0x74, 0x74, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x4c,
	0x69, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x6e, 0x6f, 0x64, 0x65,
	0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x2e, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x4e, 0x4f, 0x44, 0x45, 0x52, 0x0a, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x6e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x66, 0x66, 0x69,
	0x63, 0x69, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x09, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x66,
	0x66, 0x69, 0x63, 0x69, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x65, 0x72, 0x73,
	0x6f, 0x6e, 0x73, 0x68, 0x69, 0x70, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x62,
	0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x50, 0x45, 0x52, 0x53, 0x4f, 0x4e, 0x5f, 0x53, 0x48,
	0x49, 0x50, 0x5f, 0x49, 0x4e, 0x5f, 0x50, 0x41, 0x47, 0x45, 0x52, 0x0a, 0x70, 0x65, 0x72, 0x73,
	0x6f, 0x6e, 0x73, 0x68, 0x69, 0x70, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66,
}

var (
	file_EVENT_BASE_proto_rawDescOnce sync.Once
	file_EVENT_BASE_proto_rawDescData = file_EVENT_BASE_proto_rawDesc
)

func file_EVENT_BASE_proto_rawDescGZIP() []byte {
	file_EVENT_BASE_proto_rawDescOnce.Do(func() {
		file_EVENT_BASE_proto_rawDescData = protoimpl.X.CompressGZIP(file_EVENT_BASE_proto_rawDescData)
	})
	return file_EVENT_BASE_proto_rawDescData
}

var file_EVENT_BASE_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_EVENT_BASE_proto_goTypes = []any{
	(*EVENT_BASE)(nil),          // 0: belfast.EVENT_BASE
	(*SHIP_IN_EVENT)(nil),       // 1: belfast.SHIP_IN_EVENT
	(*KEYVALUE)(nil),            // 2: belfast.KEYVALUE
	(*EVENT_NODE)(nil),          // 3: belfast.EVENT_NODE
	(*PERSON_SHIP_IN_PAGE)(nil), // 4: belfast.PERSON_SHIP_IN_PAGE
}
var file_EVENT_BASE_proto_depIdxs = []int32{
	1, // 0: belfast.EVENT_BASE.shipinevent:type_name -> belfast.SHIP_IN_EVENT
	2, // 1: belfast.EVENT_BASE.attr_acc_list:type_name -> belfast.KEYVALUE
	2, // 2: belfast.EVENT_BASE.attr_count_list:type_name -> belfast.KEYVALUE
	3, // 3: belfast.EVENT_BASE.eventnodes:type_name -> belfast.EVENT_NODE
	4, // 4: belfast.EVENT_BASE.personship:type_name -> belfast.PERSON_SHIP_IN_PAGE
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_EVENT_BASE_proto_init() }
func file_EVENT_BASE_proto_init() {
	if File_EVENT_BASE_proto != nil {
		return
	}
	file_SHIP_IN_EVENT_proto_init()
	file_KEYVALUE_proto_init()
	file_EVENT_NODE_proto_init()
	file_PERSON_SHIP_IN_PAGE_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_EVENT_BASE_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*EVENT_BASE); i {
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
			RawDescriptor: file_EVENT_BASE_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_EVENT_BASE_proto_goTypes,
		DependencyIndexes: file_EVENT_BASE_proto_depIdxs,
		MessageInfos:      file_EVENT_BASE_proto_msgTypes,
	}.Build()
	File_EVENT_BASE_proto = out.File
	file_EVENT_BASE_proto_rawDesc = nil
	file_EVENT_BASE_proto_goTypes = nil
	file_EVENT_BASE_proto_depIdxs = nil
}
