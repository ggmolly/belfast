// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_11752.proto

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

type SC_11752 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Active          *uint32 `protobuf:"varint,1,req,name=active" json:"active,omitempty"`
	ReturnLv        *uint32 `protobuf:"varint,2,opt,name=return_lv,json=returnLv" json:"return_lv,omitempty"`
	ReturnTime      *uint32 `protobuf:"varint,3,opt,name=return_time,json=returnTime" json:"return_time,omitempty"`
	ShipNumber      *uint32 `protobuf:"varint,4,opt,name=ship_number,json=shipNumber" json:"ship_number,omitempty"`
	LastOfflineTime *uint32 `protobuf:"varint,5,opt,name=last_offline_time,json=lastOfflineTime" json:"last_offline_time,omitempty"`
	Pt              *uint32 `protobuf:"varint,6,opt,name=pt" json:"pt,omitempty"`
	SignCnt         *uint32 `protobuf:"varint,7,opt,name=sign_cnt,json=signCnt" json:"sign_cnt,omitempty"`
	SignLastTime    *uint32 `protobuf:"varint,8,opt,name=sign_last_time,json=signLastTime" json:"sign_last_time,omitempty"`
	PtStage         *uint32 `protobuf:"varint,9,opt,name=pt_stage,json=ptStage" json:"pt_stage,omitempty"`
}

func (x *SC_11752) Reset() {
	*x = SC_11752{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_11752_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_11752) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_11752) ProtoMessage() {}

func (x *SC_11752) ProtoReflect() protoreflect.Message {
	mi := &file_SC_11752_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_11752.ProtoReflect.Descriptor instead.
func (*SC_11752) Descriptor() ([]byte, []int) {
	return file_SC_11752_proto_rawDescGZIP(), []int{0}
}

func (x *SC_11752) GetActive() uint32 {
	if x != nil && x.Active != nil {
		return *x.Active
	}
	return 0
}

func (x *SC_11752) GetReturnLv() uint32 {
	if x != nil && x.ReturnLv != nil {
		return *x.ReturnLv
	}
	return 0
}

func (x *SC_11752) GetReturnTime() uint32 {
	if x != nil && x.ReturnTime != nil {
		return *x.ReturnTime
	}
	return 0
}

func (x *SC_11752) GetShipNumber() uint32 {
	if x != nil && x.ShipNumber != nil {
		return *x.ShipNumber
	}
	return 0
}

func (x *SC_11752) GetLastOfflineTime() uint32 {
	if x != nil && x.LastOfflineTime != nil {
		return *x.LastOfflineTime
	}
	return 0
}

func (x *SC_11752) GetPt() uint32 {
	if x != nil && x.Pt != nil {
		return *x.Pt
	}
	return 0
}

func (x *SC_11752) GetSignCnt() uint32 {
	if x != nil && x.SignCnt != nil {
		return *x.SignCnt
	}
	return 0
}

func (x *SC_11752) GetSignLastTime() uint32 {
	if x != nil && x.SignLastTime != nil {
		return *x.SignLastTime
	}
	return 0
}

func (x *SC_11752) GetPtStage() uint32 {
	if x != nil && x.PtStage != nil {
		return *x.PtStage
	}
	return 0
}

var File_SC_11752_proto protoreflect.FileDescriptor

var file_SC_11752_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x31, 0x31, 0x37, 0x35, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x22, 0x99, 0x02, 0x0a, 0x08, 0x53, 0x43,
	0x5f, 0x31, 0x31, 0x37, 0x35, 0x32, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12, 0x1b,
	0x0a, 0x09, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x5f, 0x6c, 0x76, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x08, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x4c, 0x76, 0x12, 0x1f, 0x0a, 0x0b, 0x72,
	0x65, 0x74, 0x75, 0x72, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0a, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x73, 0x68, 0x69, 0x70, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0a, 0x73, 0x68, 0x69, 0x70, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a,
	0x11, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x4f, 0x66,
	0x66, 0x6c, 0x69, 0x6e, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x70, 0x74, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x70, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x69, 0x67,
	0x6e, 0x5f, 0x63, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x73, 0x69, 0x67,
	0x6e, 0x43, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x5f, 0x6c, 0x61, 0x73,
	0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x73, 0x69,
	0x67, 0x6e, 0x4c, 0x61, 0x73, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x74,
	0x5f, 0x73, 0x74, 0x61, 0x67, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x70, 0x74,
	0x53, 0x74, 0x61, 0x67, 0x65, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_SC_11752_proto_rawDescOnce sync.Once
	file_SC_11752_proto_rawDescData = file_SC_11752_proto_rawDesc
)

func file_SC_11752_proto_rawDescGZIP() []byte {
	file_SC_11752_proto_rawDescOnce.Do(func() {
		file_SC_11752_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_11752_proto_rawDescData)
	})
	return file_SC_11752_proto_rawDescData
}

var file_SC_11752_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_11752_proto_goTypes = []any{
	(*SC_11752)(nil), // 0: belfast.SC_11752
}
var file_SC_11752_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_SC_11752_proto_init() }
func file_SC_11752_proto_init() {
	if File_SC_11752_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_SC_11752_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_11752); i {
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
			RawDescriptor: file_SC_11752_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_11752_proto_goTypes,
		DependencyIndexes: file_SC_11752_proto_depIdxs,
		MessageInfos:      file_SC_11752_proto_msgTypes,
	}.Build()
	File_SC_11752_proto = out.File
	file_SC_11752_proto_rawDesc = nil
	file_SC_11752_proto_goTypes = nil
	file_SC_11752_proto_depIdxs = nil
}
