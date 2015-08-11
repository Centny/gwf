// Code generated by protoc-gen-go.
// source: hrv.proto
// DO NOT EDIT!

/*
Package hrv is a generated protocol buffer package.

It is generated from these files:
	hrv.proto

It has these top-level messages:
	Res
*/
package hrv

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type Res struct {
	Code             *int32 `protobuf:"varint,1,req,name=code" json:"code,omitempty"`
	Data             []byte `protobuf:"bytes,2,req,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Res) Reset()         { *m = Res{} }
func (m *Res) String() string { return proto.CompactTextString(m) }
func (*Res) ProtoMessage()    {}

func (m *Res) GetCode() int32 {
	if m != nil && m.Code != nil {
		return *m.Code
	}
	return 0
}

func (m *Res) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
}
