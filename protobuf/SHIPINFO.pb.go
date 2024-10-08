// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: SHIPINFO.proto

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

type SHIPINFO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                  *uint32           `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	TemplateId          *uint32           `protobuf:"varint,2,req,name=template_id,json=templateId" json:"template_id,omitempty"`
	Level               *uint32           `protobuf:"varint,3,req,name=level" json:"level,omitempty"`
	Exp                 *uint32           `protobuf:"varint,4,req,name=exp" json:"exp,omitempty"`
	EquipInfoList       []*EQUIPSKIN_INFO `protobuf:"bytes,5,rep,name=equip_info_list,json=equipInfoList" json:"equip_info_list,omitempty"`
	Energy              *uint32           `protobuf:"varint,6,req,name=energy" json:"energy,omitempty"`
	State               *SHIPSTATE        `protobuf:"bytes,7,req,name=state" json:"state,omitempty"`
	IsLocked            *uint32           `protobuf:"varint,8,req,name=is_locked,json=isLocked" json:"is_locked,omitempty"`
	TransformList       []*TRANSFORM_INFO `protobuf:"bytes,9,rep,name=transform_list,json=transformList" json:"transform_list,omitempty"`
	SkillIdList         []*SHIPSKILL      `protobuf:"bytes,10,rep,name=skill_id_list,json=skillIdList" json:"skill_id_list,omitempty"`
	Intimacy            *uint32           `protobuf:"varint,11,req,name=intimacy" json:"intimacy,omitempty"`
	Proficiency         *uint32           `protobuf:"varint,12,req,name=proficiency" json:"proficiency,omitempty"`
	StrengthList        []*STRENGTH_INFO  `protobuf:"bytes,13,rep,name=strength_list,json=strengthList" json:"strength_list,omitempty"`
	CreateTime          *uint32           `protobuf:"varint,14,req,name=create_time,json=createTime" json:"create_time,omitempty"`
	SkinId              *uint32           `protobuf:"varint,15,req,name=skin_id,json=skinId" json:"skin_id,omitempty"`
	Propose             *uint32           `protobuf:"varint,16,req,name=propose" json:"propose,omitempty"`
	Name                *string           `protobuf:"bytes,17,opt,name=name" json:"name,omitempty"`
	ChangeNameTimestamp *uint32           `protobuf:"varint,18,opt,name=change_name_timestamp,json=changeNameTimestamp" json:"change_name_timestamp,omitempty"`
	Commanderid         *uint32           `protobuf:"varint,19,req,name=commanderid" json:"commanderid,omitempty"`
	MaxLevel            *uint32           `protobuf:"varint,20,req,name=max_level,json=maxLevel" json:"max_level,omitempty"`
	BluePrintFlag       *uint32           `protobuf:"varint,21,req,name=blue_print_flag,json=bluePrintFlag" json:"blue_print_flag,omitempty"`
	CommonFlag          *uint32           `protobuf:"varint,22,opt,name=common_flag,json=commonFlag" json:"common_flag,omitempty"`
	ActivityNpc         *uint32           `protobuf:"varint,23,req,name=activity_npc,json=activityNpc" json:"activity_npc,omitempty"`
	MetaRepairList      []uint32          `protobuf:"varint,24,rep,name=meta_repair_list,json=metaRepairList" json:"meta_repair_list,omitempty"`
	CoreList            []*SHIPCOREINFO   `protobuf:"bytes,25,rep,name=core_list,json=coreList" json:"core_list,omitempty"`
	Spweapon            *SPWEAPONINFO     `protobuf:"bytes,26,opt,name=spweapon" json:"spweapon,omitempty"`
}

func (x *SHIPINFO) Reset() {
	*x = SHIPINFO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_SHIPINFO_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SHIPINFO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SHIPINFO) ProtoMessage() {}

func (x *SHIPINFO) ProtoReflect() protoreflect.Message {
	mi := &file_SHIPINFO_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SHIPINFO.ProtoReflect.Descriptor instead.
func (*SHIPINFO) Descriptor() ([]byte, []int) {
	return file_SHIPINFO_proto_rawDescGZIP(), []int{0}
}

func (x *SHIPINFO) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *SHIPINFO) GetTemplateId() uint32 {
	if x != nil && x.TemplateId != nil {
		return *x.TemplateId
	}
	return 0
}

func (x *SHIPINFO) GetLevel() uint32 {
	if x != nil && x.Level != nil {
		return *x.Level
	}
	return 0
}

func (x *SHIPINFO) GetExp() uint32 {
	if x != nil && x.Exp != nil {
		return *x.Exp
	}
	return 0
}

func (x *SHIPINFO) GetEquipInfoList() []*EQUIPSKIN_INFO {
	if x != nil {
		return x.EquipInfoList
	}
	return nil
}

func (x *SHIPINFO) GetEnergy() uint32 {
	if x != nil && x.Energy != nil {
		return *x.Energy
	}
	return 0
}

func (x *SHIPINFO) GetState() *SHIPSTATE {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *SHIPINFO) GetIsLocked() uint32 {
	if x != nil && x.IsLocked != nil {
		return *x.IsLocked
	}
	return 0
}

func (x *SHIPINFO) GetTransformList() []*TRANSFORM_INFO {
	if x != nil {
		return x.TransformList
	}
	return nil
}

func (x *SHIPINFO) GetSkillIdList() []*SHIPSKILL {
	if x != nil {
		return x.SkillIdList
	}
	return nil
}

func (x *SHIPINFO) GetIntimacy() uint32 {
	if x != nil && x.Intimacy != nil {
		return *x.Intimacy
	}
	return 0
}

func (x *SHIPINFO) GetProficiency() uint32 {
	if x != nil && x.Proficiency != nil {
		return *x.Proficiency
	}
	return 0
}

func (x *SHIPINFO) GetStrengthList() []*STRENGTH_INFO {
	if x != nil {
		return x.StrengthList
	}
	return nil
}

func (x *SHIPINFO) GetCreateTime() uint32 {
	if x != nil && x.CreateTime != nil {
		return *x.CreateTime
	}
	return 0
}

func (x *SHIPINFO) GetSkinId() uint32 {
	if x != nil && x.SkinId != nil {
		return *x.SkinId
	}
	return 0
}

func (x *SHIPINFO) GetPropose() uint32 {
	if x != nil && x.Propose != nil {
		return *x.Propose
	}
	return 0
}

func (x *SHIPINFO) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *SHIPINFO) GetChangeNameTimestamp() uint32 {
	if x != nil && x.ChangeNameTimestamp != nil {
		return *x.ChangeNameTimestamp
	}
	return 0
}

func (x *SHIPINFO) GetCommanderid() uint32 {
	if x != nil && x.Commanderid != nil {
		return *x.Commanderid
	}
	return 0
}

func (x *SHIPINFO) GetMaxLevel() uint32 {
	if x != nil && x.MaxLevel != nil {
		return *x.MaxLevel
	}
	return 0
}

func (x *SHIPINFO) GetBluePrintFlag() uint32 {
	if x != nil && x.BluePrintFlag != nil {
		return *x.BluePrintFlag
	}
	return 0
}

func (x *SHIPINFO) GetCommonFlag() uint32 {
	if x != nil && x.CommonFlag != nil {
		return *x.CommonFlag
	}
	return 0
}

func (x *SHIPINFO) GetActivityNpc() uint32 {
	if x != nil && x.ActivityNpc != nil {
		return *x.ActivityNpc
	}
	return 0
}

func (x *SHIPINFO) GetMetaRepairList() []uint32 {
	if x != nil {
		return x.MetaRepairList
	}
	return nil
}

func (x *SHIPINFO) GetCoreList() []*SHIPCOREINFO {
	if x != nil {
		return x.CoreList
	}
	return nil
}

func (x *SHIPINFO) GetSpweapon() *SPWEAPONINFO {
	if x != nil {
		return x.Spweapon
	}
	return nil
}

var File_SHIPINFO_proto protoreflect.FileDescriptor

var file_SHIPINFO_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x48, 0x49, 0x50, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x1a, 0x14, 0x45, 0x51, 0x55, 0x49, 0x50,
	0x53, 0x4b, 0x49, 0x4e, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0f, 0x53, 0x48, 0x49, 0x50, 0x53, 0x54, 0x41, 0x54, 0x45, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x14, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x46, 0x4f, 0x52, 0x4d, 0x5f, 0x49, 0x4e, 0x46, 0x4f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x53, 0x48, 0x49, 0x50, 0x53, 0x4b, 0x49, 0x4c,
	0x4c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x53, 0x54, 0x52, 0x45, 0x4e, 0x47, 0x54,
	0x48, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x53, 0x48,
	0x49, 0x50, 0x43, 0x4f, 0x52, 0x45, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x12, 0x53, 0x50, 0x57, 0x45, 0x41, 0x50, 0x4f, 0x4e, 0x49, 0x4e, 0x46, 0x4f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xce, 0x07, 0x0a, 0x08, 0x53, 0x48, 0x49, 0x50, 0x49, 0x4e, 0x46,
	0x4f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0a, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65,
	0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x78, 0x70, 0x18,
	0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x65, 0x78, 0x70, 0x12, 0x3f, 0x0a, 0x0f, 0x65, 0x71,
	0x75, 0x69, 0x70, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x45, 0x51,
	0x55, 0x49, 0x50, 0x53, 0x4b, 0x49, 0x4e, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0d, 0x65, 0x71,
	0x75, 0x69, 0x70, 0x49, 0x6e, 0x66, 0x6f, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x65,
	0x6e, 0x65, 0x72, 0x67, 0x79, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x06, 0x65, 0x6e, 0x65,
	0x72, 0x67, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x07, 0x20, 0x02,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53, 0x48, 0x49,
	0x50, 0x53, 0x54, 0x41, 0x54, 0x45, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1b, 0x0a,
	0x09, 0x69, 0x73, 0x5f, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x18, 0x08, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x08, 0x69, 0x73, 0x4c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x12, 0x3e, 0x0a, 0x0e, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x09, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x54, 0x52, 0x41,
	0x4e, 0x53, 0x46, 0x4f, 0x52, 0x4d, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x0d, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x36, 0x0a, 0x0d, 0x73, 0x6b,
	0x69, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x0a, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53, 0x48, 0x49, 0x50,
	0x53, 0x4b, 0x49, 0x4c, 0x4c, 0x52, 0x0b, 0x73, 0x6b, 0x69, 0x6c, 0x6c, 0x49, 0x64, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x69, 0x6d, 0x61, 0x63, 0x79, 0x18, 0x0b,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x69, 0x6e, 0x74, 0x69, 0x6d, 0x61, 0x63, 0x79, 0x12, 0x20,
	0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x63, 0x69, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x0c, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x63, 0x69, 0x65, 0x6e, 0x63, 0x79,
	0x12, 0x3b, 0x0a, 0x0d, 0x73, 0x74, 0x72, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x5f, 0x6c, 0x69, 0x73,
	0x74, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73,
	0x74, 0x2e, 0x53, 0x54, 0x52, 0x45, 0x4e, 0x47, 0x54, 0x48, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x52,
	0x0c, 0x73, 0x74, 0x72, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1f, 0x0a,
	0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0e, 0x20, 0x02,
	0x28, 0x0d, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x17,
	0x0a, 0x07, 0x73, 0x6b, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x0f, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x06, 0x73, 0x6b, 0x69, 0x6e, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x65, 0x18, 0x10, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x32, 0x0a, 0x15, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x12,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x13, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x65, 0x72, 0x69, 0x64, 0x18, 0x13, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0b,
	0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x65, 0x72, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x6d,
	0x61, 0x78, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x14, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x08,
	0x6d, 0x61, 0x78, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x26, 0x0a, 0x0f, 0x62, 0x6c, 0x75, 0x65,
	0x5f, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x15, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x0d, 0x62, 0x6c, 0x75, 0x65, 0x50, 0x72, 0x69, 0x6e, 0x74, 0x46, 0x6c, 0x61, 0x67,
	0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x18,
	0x16, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x46, 0x6c, 0x61,
	0x67, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x5f, 0x6e, 0x70,
	0x63, 0x18, 0x17, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74,
	0x79, 0x4e, 0x70, 0x63, 0x12, 0x28, 0x0a, 0x10, 0x6d, 0x65, 0x74, 0x61, 0x5f, 0x72, 0x65, 0x70,
	0x61, 0x69, 0x72, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x18, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x0e,
	0x6d, 0x65, 0x74, 0x61, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x32,
	0x0a, 0x09, 0x63, 0x6f, 0x72, 0x65, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x19, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53, 0x48, 0x49, 0x50,
	0x43, 0x4f, 0x52, 0x45, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x08, 0x63, 0x6f, 0x72, 0x65, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x31, 0x0a, 0x08, 0x73, 0x70, 0x77, 0x65, 0x61, 0x70, 0x6f, 0x6e, 0x18, 0x1a,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x62, 0x65, 0x6c, 0x66, 0x61, 0x73, 0x74, 0x2e, 0x53,
	0x50, 0x57, 0x45, 0x41, 0x50, 0x4f, 0x4e, 0x49, 0x4e, 0x46, 0x4f, 0x52, 0x08, 0x73, 0x70, 0x77,
	0x65, 0x61, 0x70, 0x6f, 0x6e, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66,
}

var (
	file_SHIPINFO_proto_rawDescOnce sync.Once
	file_SHIPINFO_proto_rawDescData = file_SHIPINFO_proto_rawDesc
)

func file_SHIPINFO_proto_rawDescGZIP() []byte {
	file_SHIPINFO_proto_rawDescOnce.Do(func() {
		file_SHIPINFO_proto_rawDescData = protoimpl.X.CompressGZIP(file_SHIPINFO_proto_rawDescData)
	})
	return file_SHIPINFO_proto_rawDescData
}

var file_SHIPINFO_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_SHIPINFO_proto_goTypes = []any{
	(*SHIPINFO)(nil),       // 0: belfast.SHIPINFO
	(*EQUIPSKIN_INFO)(nil), // 1: belfast.EQUIPSKIN_INFO
	(*SHIPSTATE)(nil),      // 2: belfast.SHIPSTATE
	(*TRANSFORM_INFO)(nil), // 3: belfast.TRANSFORM_INFO
	(*SHIPSKILL)(nil),      // 4: belfast.SHIPSKILL
	(*STRENGTH_INFO)(nil),  // 5: belfast.STRENGTH_INFO
	(*SHIPCOREINFO)(nil),   // 6: belfast.SHIPCOREINFO
	(*SPWEAPONINFO)(nil),   // 7: belfast.SPWEAPONINFO
}
var file_SHIPINFO_proto_depIdxs = []int32{
	1, // 0: belfast.SHIPINFO.equip_info_list:type_name -> belfast.EQUIPSKIN_INFO
	2, // 1: belfast.SHIPINFO.state:type_name -> belfast.SHIPSTATE
	3, // 2: belfast.SHIPINFO.transform_list:type_name -> belfast.TRANSFORM_INFO
	4, // 3: belfast.SHIPINFO.skill_id_list:type_name -> belfast.SHIPSKILL
	5, // 4: belfast.SHIPINFO.strength_list:type_name -> belfast.STRENGTH_INFO
	6, // 5: belfast.SHIPINFO.core_list:type_name -> belfast.SHIPCOREINFO
	7, // 6: belfast.SHIPINFO.spweapon:type_name -> belfast.SPWEAPONINFO
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_SHIPINFO_proto_init() }
func file_SHIPINFO_proto_init() {
	if File_SHIPINFO_proto != nil {
		return
	}
	file_EQUIPSKIN_INFO_proto_init()
	file_SHIPSTATE_proto_init()
	file_TRANSFORM_INFO_proto_init()
	file_SHIPSKILL_proto_init()
	file_STRENGTH_INFO_proto_init()
	file_SHIPCOREINFO_proto_init()
	file_SPWEAPONINFO_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_SHIPINFO_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SHIPINFO); i {
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
			RawDescriptor: file_SHIPINFO_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_SHIPINFO_proto_goTypes,
		DependencyIndexes: file_SHIPINFO_proto_depIdxs,
		MessageInfos:      file_SHIPINFO_proto_msgTypes,
	}.Build()
	File_SHIPINFO_proto = out.File
	file_SHIPINFO_proto_rawDesc = nil
	file_SHIPINFO_proto_goTypes = nil
	file_SHIPINFO_proto_depIdxs = nil
}
