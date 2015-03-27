// Code generated by protoc-gen-go.
// source: sr.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	sr.proto

It has these top-level messages:
	KV
	Evn
*/
package pb

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type KV struct {
	Key              *string `protobuf:"bytes,1,req,name=key" json:"key,omitempty"`
	Val              *string `protobuf:"bytes,2,req,name=val" json:"val,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *KV) Reset()         { *m = KV{} }
func (m *KV) String() string { return proto.CompactTextString(m) }
func (*KV) ProtoMessage()    {}

func (m *KV) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *KV) GetVal() string {
	if m != nil && m.Val != nil {
		return *m.Val
	}
	return ""
}

type Evn struct {
	Uid              *string `protobuf:"bytes,1,req,name=uid" json:"uid,omitempty"`
	Name             *string `protobuf:"bytes,2,req,name=name" json:"name,omitempty"`
	Action           *string `protobuf:"bytes,3,req,name=action" json:"action,omitempty"`
	Time             *int64  `protobuf:"varint,4,req,name=time" json:"time,omitempty"`
	Type             *int32  `protobuf:"varint,5,req,name=type" json:"type,omitempty"`
	Kvs              []*KV   `protobuf:"bytes,6,rep,name=kvs" json:"kvs,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Evn) Reset()         { *m = Evn{} }
func (m *Evn) String() string { return proto.CompactTextString(m) }
func (*Evn) ProtoMessage()    {}

func (m *Evn) GetUid() string {
	if m != nil && m.Uid != nil {
		return *m.Uid
	}
	return ""
}

func (m *Evn) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Evn) GetAction() string {
	if m != nil && m.Action != nil {
		return *m.Action
	}
	return ""
}

func (m *Evn) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *Evn) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *Evn) GetKvs() []*KV {
	if m != nil {
		return m.Kvs
	}
	return nil
}

func init() {
}
