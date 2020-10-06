// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.6.1
// source: match.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type PGN struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GoldPlayer   string            `protobuf:"bytes,1,opt,name=gold_player,json=goldPlayer,proto3" json:"gold_player,omitempty"`
	SilverPlayer string            `protobuf:"bytes,2,opt,name=silver_player,json=silverPlayer,proto3" json:"silver_player,omitempty"`
	Pgn          string            `protobuf:"bytes,7,opt,name=pgn,proto3" json:"pgn,omitempty"`
	Steps        []uint32          `protobuf:"varint,3,rep,packed,name=steps,proto3" json:"steps,omitempty"`
	Annotations  []*PGN_Annotation `protobuf:"bytes,5,rep,name=annotations,proto3" json:"annotations,omitempty"`
	Result       int32             `protobuf:"varint,6,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *PGN) Reset() {
	*x = PGN{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PGN) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PGN) ProtoMessage() {}

func (x *PGN) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PGN.ProtoReflect.Descriptor instead.
func (*PGN) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{0}
}

func (x *PGN) GetGoldPlayer() string {
	if x != nil {
		return x.GoldPlayer
	}
	return ""
}

func (x *PGN) GetSilverPlayer() string {
	if x != nil {
		return x.SilverPlayer
	}
	return ""
}

func (x *PGN) GetPgn() string {
	if x != nil {
		return x.Pgn
	}
	return ""
}

func (x *PGN) GetSteps() []uint32 {
	if x != nil {
		return x.Steps
	}
	return nil
}

func (x *PGN) GetAnnotations() []*PGN_Annotation {
	if x != nil {
		return x.Annotations
	}
	return nil
}

func (x *PGN) GetResult() int32 {
	if x != nil {
		return x.Result
	}
	return 0
}

type Match struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string        `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Players []string      `protobuf:"bytes,2,rep,name=players,proto3" json:"players,omitempty"`
	Results *Match_Result `protobuf:"bytes,4,opt,name=results,proto3" json:"results,omitempty"`
	Games   []*Match_Game `protobuf:"bytes,5,rep,name=games,proto3" json:"games,omitempty"`
}

func (x *Match) Reset() {
	*x = Match{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Match) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Match) ProtoMessage() {}

func (x *Match) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Match.ProtoReflect.Descriptor instead.
func (*Match) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{1}
}

func (x *Match) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Match) GetPlayers() []string {
	if x != nil {
		return x.Players
	}
	return nil
}

func (x *Match) GetResults() *Match_Result {
	if x != nil {
		return x.Results
	}
	return nil
}

func (x *Match) GetGames() []*Match_Game {
	if x != nil {
		return x.Games
	}
	return nil
}

type PGN_Annotation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Step    string             `protobuf:"bytes,2,opt,name=step,proto3" json:"step,omitempty"`
	Comment string             `protobuf:"bytes,3,opt,name=comment,proto3" json:"comment,omitempty"`
	Policy  map[uint32]float32 `protobuf:"bytes,4,rep,name=policy,proto3" json:"policy,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"fixed32,2,opt,name=value,proto3"`
}

func (x *PGN_Annotation) Reset() {
	*x = PGN_Annotation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PGN_Annotation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PGN_Annotation) ProtoMessage() {}

func (x *PGN_Annotation) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PGN_Annotation.ProtoReflect.Descriptor instead.
func (*PGN_Annotation) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{0, 0}
}

func (x *PGN_Annotation) GetStep() string {
	if x != nil {
		return x.Step
	}
	return ""
}

func (x *PGN_Annotation) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

func (x *PGN_Annotation) GetPolicy() map[uint32]float32 {
	if x != nil {
		return x.Policy
	}
	return nil
}

type PGN_Step struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Step       uint32          `protobuf:"varint,1,opt,name=step,proto3" json:"step,omitempty"`
	Annotation *PGN_Annotation `protobuf:"bytes,2,opt,name=annotation,proto3" json:"annotation,omitempty"`
}

func (x *PGN_Step) Reset() {
	*x = PGN_Step{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PGN_Step) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PGN_Step) ProtoMessage() {}

func (x *PGN_Step) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PGN_Step.ProtoReflect.Descriptor instead.
func (*PGN_Step) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{0, 1}
}

func (x *PGN_Step) GetStep() uint32 {
	if x != nil {
		return x.Step
	}
	return 0
}

func (x *PGN_Step) GetAnnotation() *PGN_Annotation {
	if x != nil {
		return x.Annotation
	}
	return nil
}

type Match_Result struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Wins []uint32 `protobuf:"varint,1,rep,packed,name=wins,proto3" json:"wins,omitempty"`
}

func (x *Match_Result) Reset() {
	*x = Match_Result{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Match_Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Match_Result) ProtoMessage() {}

func (x *Match_Result) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Match_Result.ProtoReflect.Descriptor instead.
func (*Match_Result) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{1, 0}
}

func (x *Match_Result) GetWins() []uint32 {
	if x != nil {
		return x.Wins
	}
	return nil
}

type Match_Game struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GoldPlayer         uint32 `protobuf:"varint,1,opt,name=gold_player,json=goldPlayer,proto3" json:"gold_player,omitempty"`
	SilverPlayer       uint32 `protobuf:"varint,2,opt,name=silver_player,json=silverPlayer,proto3" json:"silver_player,omitempty"`
	GoldPlayerRating   uint32 `protobuf:"varint,4,opt,name=gold_player_rating,json=goldPlayerRating,proto3" json:"gold_player_rating,omitempty"`
	SilverPlayerRating uint32 `protobuf:"varint,5,opt,name=silver_player_rating,json=silverPlayerRating,proto3" json:"silver_player_rating,omitempty"`
	Rated              bool   `protobuf:"varint,6,opt,name=rated,proto3" json:"rated,omitempty"`
	Pgn                *PGN   `protobuf:"bytes,3,opt,name=pgn,proto3" json:"pgn,omitempty"`
}

func (x *Match_Game) Reset() {
	*x = Match_Game{}
	if protoimpl.UnsafeEnabled {
		mi := &file_match_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Match_Game) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Match_Game) ProtoMessage() {}

func (x *Match_Game) ProtoReflect() protoreflect.Message {
	mi := &file_match_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Match_Game.ProtoReflect.Descriptor instead.
func (*Match_Game) Descriptor() ([]byte, []int) {
	return file_match_proto_rawDescGZIP(), []int{1, 1}
}

func (x *Match_Game) GetGoldPlayer() uint32 {
	if x != nil {
		return x.GoldPlayer
	}
	return 0
}

func (x *Match_Game) GetSilverPlayer() uint32 {
	if x != nil {
		return x.SilverPlayer
	}
	return 0
}

func (x *Match_Game) GetGoldPlayerRating() uint32 {
	if x != nil {
		return x.GoldPlayerRating
	}
	return 0
}

func (x *Match_Game) GetSilverPlayerRating() uint32 {
	if x != nil {
		return x.SilverPlayerRating
	}
	return 0
}

func (x *Match_Game) GetRated() bool {
	if x != nil {
		return x.Rated
	}
	return false
}

func (x *Match_Game) GetPgn() *PGN {
	if x != nil {
		return x.Pgn
	}
	return nil
}

var File_match_proto protoreflect.FileDescriptor

var file_match_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbc, 0x03,
	0x0a, 0x03, 0x50, 0x47, 0x4e, 0x12, 0x1f, 0x0a, 0x0b, 0x67, 0x6f, 0x6c, 0x64, 0x5f, 0x70, 0x6c,
	0x61, 0x79, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x67, 0x6f, 0x6c, 0x64,
	0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x69, 0x6c, 0x76, 0x65, 0x72,
	0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73,
	0x69, 0x6c, 0x76, 0x65, 0x72, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x70,
	0x67, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x70, 0x67, 0x6e, 0x12, 0x18, 0x0a,
	0x05, 0x73, 0x74, 0x65, 0x70, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0d, 0x42, 0x02, 0x10, 0x01,
	0x52, 0x05, 0x73, 0x74, 0x65, 0x70, 0x73, 0x12, 0x31, 0x0a, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x50,
	0x47, 0x4e, 0x2e, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x1a, 0xaa, 0x01, 0x0a, 0x0a, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x74, 0x65, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x73, 0x74, 0x65, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x12,
	0x33, 0x0a, 0x06, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x50, 0x47, 0x4e, 0x2e, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x70, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x1a, 0x39, 0x0a, 0x0b, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x4b, 0x0a, 0x04, 0x53, 0x74, 0x65, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x74, 0x65, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x73, 0x74, 0x65, 0x70, 0x12, 0x2f, 0x0a, 0x0a, 0x61,
	0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x50, 0x47, 0x4e, 0x2e, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x0a, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xfc, 0x02, 0x0a,
	0x05, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73,
	0x12, 0x27, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x52, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x12, 0x21, 0x0a, 0x05, 0x67, 0x61, 0x6d,
	0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68,
	0x2e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x05, 0x67, 0x61, 0x6d, 0x65, 0x73, 0x1a, 0x20, 0x0a, 0x06,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x04, 0x77, 0x69, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0d, 0x42, 0x02, 0x10, 0x01, 0x52, 0x04, 0x77, 0x69, 0x6e, 0x73, 0x1a, 0xda,
	0x01, 0x0a, 0x04, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x67, 0x6f, 0x6c, 0x64, 0x5f,
	0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x67, 0x6f,
	0x6c, 0x64, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x69, 0x6c, 0x76,
	0x65, 0x72, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x0c, 0x73, 0x69, 0x6c, 0x76, 0x65, 0x72, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x2c, 0x0a,
	0x12, 0x67, 0x6f, 0x6c, 0x64, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x72, 0x61, 0x74,
	0x69, 0x6e, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x10, 0x67, 0x6f, 0x6c, 0x64, 0x50,
	0x6c, 0x61, 0x79, 0x65, 0x72, 0x52, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x30, 0x0a, 0x14, 0x73,
	0x69, 0x6c, 0x76, 0x65, 0x72, 0x5f, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x72, 0x61, 0x74,
	0x69, 0x6e, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x12, 0x73, 0x69, 0x6c, 0x76, 0x65,
	0x72, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x52, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x14, 0x0a,
	0x05, 0x72, 0x61, 0x74, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x72, 0x61,
	0x74, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x03, 0x70, 0x67, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x04, 0x2e, 0x50, 0x47, 0x4e, 0x52, 0x03, 0x70, 0x67, 0x6e, 0x42, 0x21, 0x5a, 0x1f, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x6a, 0x7a, 0x61, 0x66, 0x66,
	0x2f, 0x62, 0x6f, 0x74, 0x5f, 0x7a, 0x6f, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_match_proto_rawDescOnce sync.Once
	file_match_proto_rawDescData = file_match_proto_rawDesc
)

func file_match_proto_rawDescGZIP() []byte {
	file_match_proto_rawDescOnce.Do(func() {
		file_match_proto_rawDescData = protoimpl.X.CompressGZIP(file_match_proto_rawDescData)
	})
	return file_match_proto_rawDescData
}

var file_match_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_match_proto_goTypes = []interface{}{
	(*PGN)(nil),            // 0: PGN
	(*Match)(nil),          // 1: Match
	(*PGN_Annotation)(nil), // 2: PGN.Annotation
	(*PGN_Step)(nil),       // 3: PGN.Step
	nil,                    // 4: PGN.Annotation.PolicyEntry
	(*Match_Result)(nil),   // 5: Match.Result
	(*Match_Game)(nil),     // 6: Match.Game
}
var file_match_proto_depIdxs = []int32{
	2, // 0: PGN.annotations:type_name -> PGN.Annotation
	5, // 1: Match.results:type_name -> Match.Result
	6, // 2: Match.games:type_name -> Match.Game
	4, // 3: PGN.Annotation.policy:type_name -> PGN.Annotation.PolicyEntry
	2, // 4: PGN.Step.annotation:type_name -> PGN.Annotation
	0, // 5: Match.Game.pgn:type_name -> PGN
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_match_proto_init() }
func file_match_proto_init() {
	if File_match_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_match_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PGN); i {
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
		file_match_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Match); i {
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
		file_match_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PGN_Annotation); i {
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
		file_match_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PGN_Step); i {
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
		file_match_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Match_Result); i {
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
		file_match_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Match_Game); i {
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
			RawDescriptor: file_match_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_match_proto_goTypes,
		DependencyIndexes: file_match_proto_depIdxs,
		MessageInfos:      file_match_proto_msgTypes,
	}.Build()
	File_match_proto = out.File
	file_match_proto_rawDesc = nil
	file_match_proto_goTypes = nil
	file_match_proto_depIdxs = nil
}
