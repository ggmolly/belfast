// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SC_10021.proto

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

type SC_10021 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result         *uint32       `protobuf:"varint,1,req,name=result" json:"result,omitempty"`
	Serverlist     []*SERVERINFO `protobuf:"bytes,2,rep,name=serverlist" json:"serverlist,omitempty"`
	AccountId      *uint32       `protobuf:"varint,3,req,name=account_id,json=accountId" json:"account_id,omitempty"`
	ServerTicket   *string       `protobuf:"bytes,4,req,name=server_ticket,json=serverTicket" json:"server_ticket,omitempty"`
	NoticeList     []*NOTICEINFO `protobuf:"bytes,5,rep,name=notice_list,json=noticeList" json:"notice_list,omitempty"`
	Device         *uint32       `protobuf:"varint,6,opt,name=device" json:"device,omitempty"`
	LimitServerIds []uint32      `protobuf:"varint,7,rep,name=limit_server_ids,json=limitServerIds" json:"limit_server_ids,omitempty"`
}

func (x *SC_10021) Reset() {
	*x = SC_10021{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SC_10021_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SC_10021) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SC_10021) ProtoMessage() {}

func (x *SC_10021) ProtoReflect() protoreflect.Message {
	mi := &file_SC_10021_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SC_10021.ProtoReflect.Descriptor instead.
func (*SC_10021) Descriptor() ([]byte, []int) {
	return file_SC_10021_proto_rawDescGZIP(), []int{0}
}

func (x *SC_10021) GetResult() uint32 {
	if x != nil && x.Result != nil {
		return *x.Result
	}
	return 0
}

func (x *SC_10021) GetServerlist() []*SERVERINFO {
	if x != nil {
		return x.Serverlist
	}
	return nil
}

func (x *SC_10021) GetAccountId() uint32 {
	if x != nil && x.AccountId != nil {
		return *x.AccountId
	}
	return 0
}

func (x *SC_10021) GetServerTicket() string {
	if x != nil && x.ServerTicket != nil {
		return *x.ServerTicket
	}
	return ""
}

func (x *SC_10021) GetNoticeList() []*NOTICEINFO {
	if x != nil {
		return x.NoticeList
	}
	return nil
}

func (x *SC_10021) GetDevice() uint32 {
	if x != nil && x.Device != nil {
		return *x.Device
	}
	return 0
}

func (x *SC_10021) GetLimitServerIds() []uint32 {
	if x != nil {
		return x.LimitServerIds
	}
	return nil
}

var File_SC_10021_proto protoreflect.FileDescriptor

var file_SC_10021_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x43, 0x5f, 0x31, 0x30, 0x30, 0x32, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x10, 0x53, 0x45, 0x52, 0x56, 0x45,
	0x52, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x4e, 0x4f, 0x54,
	0x49, 0x43, 0x45, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x93, 0x02,
	0x0a, 0x08, 0x53, 0x43, 0x5f, 0x31, 0x30, 0x30, 0x32, 0x31, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x33, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x6c, 0x69, 0x73, 0x74,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74,
	0x2e, 0x53, 0x45, 0x52, 0x56, 0x45, 0x52, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0a, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x09, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x5f, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x04, 0x20, 0x02, 0x28, 0x09, 0x52, 0x0c, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x34, 0x0a, 0x0b, 0x6e,
	0x6f, 0x74, 0x69, 0x63, 0x65, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x4e, 0x4f, 0x54, 0x49, 0x43,
	0x45, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0a, 0x6e, 0x6f, 0x74, 0x69, 0x63, 0x65, 0x4c, 0x69, 0x73,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x07, 0x20,
	0x03, 0x28, 0x0d, 0x52, 0x0e, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x49, 0x64, 0x73, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66,
}

var (
	file_SC_10021_proto_rawDescOnce sync.Once
	file_SC_10021_proto_rawDescData = file_SC_10021_proto_rawDesc
)

func file_SC_10021_proto_rawDescGZIP() []byte {
	file_SC_10021_proto_rawDescOnce.Do(func() {
		file_SC_10021_proto_rawDescData = protoimpl.X.CompressGZIP(file_SC_10021_proto_rawDescData)
	})
	return file_SC_10021_proto_rawDescData
}

var file_SC_10021_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SC_10021_proto_goTypes = []any{
	(*SC_10021)(nil),   // 0: belfast.SC_10021
	(*SERVERINFO)(nil), // 1: belfast.SERVERINFO
	(*NOTICEINFO)(nil), // 2: belfast.NOTICEINFO
}
var file_SC_10021_proto_depIdxs = []int32{
	1, // 0: belfast.SC_10021.serverlist:type_name -> belfast.SERVERINFO
	2, // 1: belfast.SC_10021.notice_list:type_name -> belfast.NOTICEINFO
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_SC_10021_proto_init() }
func file_SC_10021_proto_init() {
	if File_SC_10021_proto != nil {
		return
	}
	file_SERVERINFO_proto_init()
	file_NOTICEINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SC_10021_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SC_10021); i {
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
			RawDescriptor: file_SC_10021_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SC_10021_proto_goTypes,
		DependencyIndexes: file_SC_10021_proto_depIdxs,
		MessageInfos:      file_SC_10021_proto_msgTypes,
	}.Build()
	File_SC_10021_proto = out.File
	file_SC_10021_proto_rawDesc = nil
	file_SC_10021_proto_goTypes = nil
	file_SC_10021_proto_depIdxs = nil
}
