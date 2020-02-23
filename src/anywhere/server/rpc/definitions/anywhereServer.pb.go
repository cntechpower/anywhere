// Code generated by protoc-gen-go. DO NOT EDIT.
// source: anywhere/server/rpc/definitions/anywhereServer.proto

package anywhereRpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{0}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Agent struct {
	AgentId              string   `protobuf:"bytes,1,opt,name=agentId,proto3" json:"agentId,omitempty"`
	AgentRemoteAddr      string   `protobuf:"bytes,2,opt,name=agentRemoteAddr,proto3" json:"agentRemoteAddr,omitempty"`
	AgentLastAckRcv      string   `protobuf:"bytes,3,opt,name=agentLastAckRcv,proto3" json:"agentLastAckRcv,omitempty"`
	AgentLastAckSend     string   `protobuf:"bytes,4,opt,name=agentLastAckSend,proto3" json:"agentLastAckSend,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Agent) Reset()         { *m = Agent{} }
func (m *Agent) String() string { return proto.CompactTextString(m) }
func (*Agent) ProtoMessage()    {}
func (*Agent) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{1}
}

func (m *Agent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Agent.Unmarshal(m, b)
}
func (m *Agent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Agent.Marshal(b, m, deterministic)
}
func (m *Agent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Agent.Merge(m, src)
}
func (m *Agent) XXX_Size() int {
	return xxx_messageInfo_Agent.Size(m)
}
func (m *Agent) XXX_DiscardUnknown() {
	xxx_messageInfo_Agent.DiscardUnknown(m)
}

var xxx_messageInfo_Agent proto.InternalMessageInfo

func (m *Agent) GetAgentId() string {
	if m != nil {
		return m.AgentId
	}
	return ""
}

func (m *Agent) GetAgentRemoteAddr() string {
	if m != nil {
		return m.AgentRemoteAddr
	}
	return ""
}

func (m *Agent) GetAgentLastAckRcv() string {
	if m != nil {
		return m.AgentLastAckRcv
	}
	return ""
}

func (m *Agent) GetAgentLastAckSend() string {
	if m != nil {
		return m.AgentLastAckSend
	}
	return ""
}

type Agents struct {
	Agent                []*Agent `protobuf:"bytes,1,rep,name=agent,proto3" json:"agent,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Agents) Reset()         { *m = Agents{} }
func (m *Agents) String() string { return proto.CompactTextString(m) }
func (*Agents) ProtoMessage()    {}
func (*Agents) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{2}
}

func (m *Agents) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Agents.Unmarshal(m, b)
}
func (m *Agents) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Agents.Marshal(b, m, deterministic)
}
func (m *Agents) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Agents.Merge(m, src)
}
func (m *Agents) XXX_Size() int {
	return xxx_messageInfo_Agents.Size(m)
}
func (m *Agents) XXX_DiscardUnknown() {
	xxx_messageInfo_Agents.DiscardUnknown(m)
}

var xxx_messageInfo_Agents proto.InternalMessageInfo

func (m *Agents) GetAgent() []*Agent {
	if m != nil {
		return m.Agent
	}
	return nil
}

type ProxyConfig struct {
	AgentId              string   `protobuf:"bytes,1,opt,name=agentId,proto3" json:"agentId,omitempty"`
	RemotePort           int64    `protobuf:"varint,2,opt,name=remotePort,proto3" json:"remotePort,omitempty"`
	LocalAddr            string   `protobuf:"bytes,3,opt,name=localAddr,proto3" json:"localAddr,omitempty"`
	IsWhiteListOn        bool     `protobuf:"varint,4,opt,name=isWhiteListOn,proto3" json:"isWhiteListOn,omitempty"`
	WhiteListIps         string   `protobuf:"bytes,5,opt,name=whiteListIps,proto3" json:"whiteListIps,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProxyConfig) Reset()         { *m = ProxyConfig{} }
func (m *ProxyConfig) String() string { return proto.CompactTextString(m) }
func (*ProxyConfig) ProtoMessage()    {}
func (*ProxyConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{3}
}

func (m *ProxyConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProxyConfig.Unmarshal(m, b)
}
func (m *ProxyConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProxyConfig.Marshal(b, m, deterministic)
}
func (m *ProxyConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProxyConfig.Merge(m, src)
}
func (m *ProxyConfig) XXX_Size() int {
	return xxx_messageInfo_ProxyConfig.Size(m)
}
func (m *ProxyConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ProxyConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ProxyConfig proto.InternalMessageInfo

func (m *ProxyConfig) GetAgentId() string {
	if m != nil {
		return m.AgentId
	}
	return ""
}

func (m *ProxyConfig) GetRemotePort() int64 {
	if m != nil {
		return m.RemotePort
	}
	return 0
}

func (m *ProxyConfig) GetLocalAddr() string {
	if m != nil {
		return m.LocalAddr
	}
	return ""
}

func (m *ProxyConfig) GetIsWhiteListOn() bool {
	if m != nil {
		return m.IsWhiteListOn
	}
	return false
}

func (m *ProxyConfig) GetWhiteListIps() string {
	if m != nil {
		return m.WhiteListIps
	}
	return ""
}

type AddProxyConfigInput struct {
	Config               *ProxyConfig `protobuf:"bytes,1,opt,name=Config,proto3" json:"Config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *AddProxyConfigInput) Reset()         { *m = AddProxyConfigInput{} }
func (m *AddProxyConfigInput) String() string { return proto.CompactTextString(m) }
func (*AddProxyConfigInput) ProtoMessage()    {}
func (*AddProxyConfigInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{4}
}

func (m *AddProxyConfigInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddProxyConfigInput.Unmarshal(m, b)
}
func (m *AddProxyConfigInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddProxyConfigInput.Marshal(b, m, deterministic)
}
func (m *AddProxyConfigInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddProxyConfigInput.Merge(m, src)
}
func (m *AddProxyConfigInput) XXX_Size() int {
	return xxx_messageInfo_AddProxyConfigInput.Size(m)
}
func (m *AddProxyConfigInput) XXX_DiscardUnknown() {
	xxx_messageInfo_AddProxyConfigInput.DiscardUnknown(m)
}

var xxx_messageInfo_AddProxyConfigInput proto.InternalMessageInfo

func (m *AddProxyConfigInput) GetConfig() *ProxyConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

type ListProxyConfigsOutput struct {
	Config               []*ProxyConfig `protobuf:"bytes,1,rep,name=Config,proto3" json:"Config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ListProxyConfigsOutput) Reset()         { *m = ListProxyConfigsOutput{} }
func (m *ListProxyConfigsOutput) String() string { return proto.CompactTextString(m) }
func (*ListProxyConfigsOutput) ProtoMessage()    {}
func (*ListProxyConfigsOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{5}
}

func (m *ListProxyConfigsOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListProxyConfigsOutput.Unmarshal(m, b)
}
func (m *ListProxyConfigsOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListProxyConfigsOutput.Marshal(b, m, deterministic)
}
func (m *ListProxyConfigsOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListProxyConfigsOutput.Merge(m, src)
}
func (m *ListProxyConfigsOutput) XXX_Size() int {
	return xxx_messageInfo_ListProxyConfigsOutput.Size(m)
}
func (m *ListProxyConfigsOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_ListProxyConfigsOutput.DiscardUnknown(m)
}

var xxx_messageInfo_ListProxyConfigsOutput proto.InternalMessageInfo

func (m *ListProxyConfigsOutput) GetConfig() []*ProxyConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

type RemoveProxyConfigInput struct {
	AgentId              string   `protobuf:"bytes,1,opt,name=agentId,proto3" json:"agentId,omitempty"`
	LocalAddr            string   `protobuf:"bytes,2,opt,name=localAddr,proto3" json:"localAddr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveProxyConfigInput) Reset()         { *m = RemoveProxyConfigInput{} }
func (m *RemoveProxyConfigInput) String() string { return proto.CompactTextString(m) }
func (*RemoveProxyConfigInput) ProtoMessage()    {}
func (*RemoveProxyConfigInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_0707d1479bfd60ca, []int{6}
}

func (m *RemoveProxyConfigInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveProxyConfigInput.Unmarshal(m, b)
}
func (m *RemoveProxyConfigInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveProxyConfigInput.Marshal(b, m, deterministic)
}
func (m *RemoveProxyConfigInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveProxyConfigInput.Merge(m, src)
}
func (m *RemoveProxyConfigInput) XXX_Size() int {
	return xxx_messageInfo_RemoveProxyConfigInput.Size(m)
}
func (m *RemoveProxyConfigInput) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveProxyConfigInput.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveProxyConfigInput proto.InternalMessageInfo

func (m *RemoveProxyConfigInput) GetAgentId() string {
	if m != nil {
		return m.AgentId
	}
	return ""
}

func (m *RemoveProxyConfigInput) GetLocalAddr() string {
	if m != nil {
		return m.LocalAddr
	}
	return ""
}

func init() {
	proto.RegisterType((*Empty)(nil), "anywhereRpc.Empty")
	proto.RegisterType((*Agent)(nil), "anywhereRpc.Agent")
	proto.RegisterType((*Agents)(nil), "anywhereRpc.Agents")
	proto.RegisterType((*ProxyConfig)(nil), "anywhereRpc.ProxyConfig")
	proto.RegisterType((*AddProxyConfigInput)(nil), "anywhereRpc.AddProxyConfigInput")
	proto.RegisterType((*ListProxyConfigsOutput)(nil), "anywhereRpc.ListProxyConfigsOutput")
	proto.RegisterType((*RemoveProxyConfigInput)(nil), "anywhereRpc.RemoveProxyConfigInput")
}

func init() {
	proto.RegisterFile("anywhere/server/rpc/definitions/anywhereServer.proto", fileDescriptor_0707d1479bfd60ca)
}

var fileDescriptor_0707d1479bfd60ca = []byte{
	// 447 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0x6d, 0x16, 0xda, 0xd1, 0x5b, 0x18, 0xe3, 0x56, 0x4c, 0xd6, 0x84, 0x50, 0x65, 0x78, 0xa8,
	0x78, 0x68, 0x51, 0x41, 0xbc, 0xa2, 0x08, 0xf1, 0x51, 0x54, 0x58, 0xe5, 0x22, 0xf1, 0x1c, 0x12,
	0x6f, 0xb3, 0xe8, 0xec, 0xc8, 0xf6, 0x3a, 0xfa, 0x73, 0xf8, 0x05, 0x3c, 0xf1, 0xff, 0x50, 0x5d,
	0xc2, 0xec, 0x24, 0x45, 0xec, 0xcd, 0x3e, 0xf7, 0xf4, 0xf4, 0x9c, 0xe3, 0xab, 0xc0, 0x8b, 0x54,
	0xae, 0xaf, 0xce, 0xb9, 0xe6, 0x63, 0xc3, 0xf5, 0x8a, 0xeb, 0xb1, 0x2e, 0xb2, 0x71, 0xce, 0x4f,
	0x85, 0x14, 0x56, 0x28, 0x69, 0xc6, 0xe5, 0x7c, 0xe1, 0xc6, 0xa3, 0x42, 0x2b, 0xab, 0xb0, 0x57,
	0xa2, 0xac, 0xc8, 0xe8, 0x3e, 0xb4, 0xdf, 0x5c, 0x14, 0x76, 0x4d, 0x7f, 0x44, 0xd0, 0x4e, 0xce,
	0xb8, 0xb4, 0x48, 0x60, 0x3f, 0xdd, 0x1c, 0xa6, 0x39, 0x89, 0x06, 0xd1, 0xb0, 0xcb, 0xca, 0x2b,
	0x0e, 0xe1, 0x9e, 0x3b, 0x32, 0x7e, 0xa1, 0x2c, 0x4f, 0xf2, 0x5c, 0x93, 0x3d, 0xc7, 0xa8, 0xc2,
	0x7f, 0x99, 0xb3, 0xd4, 0xd8, 0x24, 0xfb, 0xc6, 0xb2, 0x15, 0x89, 0x3d, 0xe6, 0x35, 0x8c, 0x4f,
	0xe1, 0xd0, 0x87, 0x16, 0x5c, 0xe6, 0xe4, 0x96, 0xa3, 0xd6, 0x70, 0x3a, 0x81, 0x8e, 0xb3, 0x68,
	0x70, 0x08, 0x6d, 0x37, 0x25, 0xd1, 0x20, 0x1e, 0xf6, 0x26, 0x38, 0xf2, 0x32, 0x8d, 0x1c, 0x87,
	0x6d, 0x09, 0xf4, 0x67, 0x04, 0xbd, 0xb9, 0x56, 0xdf, 0xd7, 0xaf, 0x95, 0x3c, 0x15, 0x67, 0xff,
	0x48, 0xf7, 0x08, 0x40, 0xbb, 0x04, 0x73, 0xa5, 0xad, 0x0b, 0x16, 0x33, 0x0f, 0xc1, 0x87, 0xd0,
	0x5d, 0xaa, 0x2c, 0x5d, 0xba, 0xdc, 0xdb, 0x34, 0xd7, 0x00, 0x3e, 0x81, 0xbb, 0xc2, 0x7c, 0x39,
	0x17, 0x96, 0xcf, 0x84, 0xb1, 0x27, 0xd2, 0x85, 0xb8, 0xcd, 0x42, 0x10, 0x29, 0xdc, 0xb9, 0x2a,
	0xaf, 0xd3, 0xc2, 0x90, 0xb6, 0x93, 0x09, 0x30, 0xfa, 0x0e, 0xfa, 0x49, 0x9e, 0x7b, 0x9e, 0xa7,
	0xb2, 0xb8, 0xb4, 0xf8, 0x0c, 0x3a, 0xdb, 0xab, 0xf3, 0xdd, 0x9b, 0x90, 0x20, 0xb3, 0x47, 0x67,
	0x7f, 0x78, 0xf4, 0x03, 0x1c, 0x6d, 0x34, 0xbd, 0x91, 0x39, 0xb9, 0xb4, 0x55, 0xad, 0xf8, 0xbf,
	0xb4, 0xe6, 0x70, 0xb4, 0x79, 0xde, 0x15, 0xaf, 0xf9, 0xda, 0x5d, 0x68, 0x50, 0xd8, 0x5e, 0xa5,
	0xb0, 0xc9, 0xaf, 0x18, 0x0e, 0x92, 0x60, 0x3f, 0xf1, 0x25, 0x74, 0x37, 0x86, 0xb7, 0x6b, 0x18,
	0xbe, 0xa9, 0x5b, 0xd2, 0xe3, 0x7e, 0xfd, 0x9d, 0x0d, 0x6d, 0xe1, 0x7b, 0x38, 0x08, 0x1b, 0xc3,
	0x41, 0x48, 0xac, 0xd7, 0x79, 0xdc, 0x20, 0x4f, 0x5b, 0xf8, 0x11, 0x0e, 0xab, 0x95, 0x35, 0x1a,
	0x79, 0x1c, 0x60, 0xcd, 0x2d, 0xd3, 0x16, 0x7e, 0x82, 0xfb, 0xb5, 0xd6, 0x30, 0xfc, 0x6d, 0x73,
	0xab, 0x3b, 0xec, 0xbd, 0x82, 0xfe, 0x4c, 0xa5, 0x7e, 0x98, 0xb7, 0x62, 0xc9, 0x1b, 0x1d, 0x36,
	0x0b, 0x24, 0xf0, 0x60, 0x91, 0x06, 0x7f, 0xf7, 0x59, 0xdd, 0x4c, 0xe2, 0x6b, 0xc7, 0x7d, 0x45,
	0x9e, 0xff, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x28, 0x2c, 0xd1, 0x60, 0x7d, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AnywhereServerClient is the client API for AnywhereServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AnywhereServerClient interface {
	ListAgent(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Agents, error)
	AddProxyConfig(ctx context.Context, in *AddProxyConfigInput, opts ...grpc.CallOption) (*Empty, error)
	ListProxyConfigs(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ListProxyConfigsOutput, error)
	RemoveProxyConfig(ctx context.Context, in *RemoveProxyConfigInput, opts ...grpc.CallOption) (*Empty, error)
	LoadProxyConfigFile(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	SaveProxyConfigToFile(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type anywhereServerClient struct {
	cc *grpc.ClientConn
}

func NewAnywhereServerClient(cc *grpc.ClientConn) AnywhereServerClient {
	return &anywhereServerClient{cc}
}

func (c *anywhereServerClient) ListAgent(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Agents, error) {
	out := new(Agents)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/ListAgent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *anywhereServerClient) AddProxyConfig(ctx context.Context, in *AddProxyConfigInput, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/AddProxyConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *anywhereServerClient) ListProxyConfigs(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ListProxyConfigsOutput, error) {
	out := new(ListProxyConfigsOutput)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/ListProxyConfigs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *anywhereServerClient) RemoveProxyConfig(ctx context.Context, in *RemoveProxyConfigInput, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/RemoveProxyConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *anywhereServerClient) LoadProxyConfigFile(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/LoadProxyConfigFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *anywhereServerClient) SaveProxyConfigToFile(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/anywhereRpc.AnywhereServer/SaveProxyConfigToFile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AnywhereServerServer is the server API for AnywhereServer service.
type AnywhereServerServer interface {
	ListAgent(context.Context, *Empty) (*Agents, error)
	AddProxyConfig(context.Context, *AddProxyConfigInput) (*Empty, error)
	ListProxyConfigs(context.Context, *Empty) (*ListProxyConfigsOutput, error)
	RemoveProxyConfig(context.Context, *RemoveProxyConfigInput) (*Empty, error)
	LoadProxyConfigFile(context.Context, *Empty) (*Empty, error)
	SaveProxyConfigToFile(context.Context, *Empty) (*Empty, error)
}

// UnimplementedAnywhereServerServer can be embedded to have forward compatible implementations.
type UnimplementedAnywhereServerServer struct {
}

func (*UnimplementedAnywhereServerServer) ListAgent(ctx context.Context, req *Empty) (*Agents, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAgent not implemented")
}
func (*UnimplementedAnywhereServerServer) AddProxyConfig(ctx context.Context, req *AddProxyConfigInput) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddProxyConfig not implemented")
}
func (*UnimplementedAnywhereServerServer) ListProxyConfigs(ctx context.Context, req *Empty) (*ListProxyConfigsOutput, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProxyConfigs not implemented")
}
func (*UnimplementedAnywhereServerServer) RemoveProxyConfig(ctx context.Context, req *RemoveProxyConfigInput) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveProxyConfig not implemented")
}
func (*UnimplementedAnywhereServerServer) LoadProxyConfigFile(ctx context.Context, req *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadProxyConfigFile not implemented")
}
func (*UnimplementedAnywhereServerServer) SaveProxyConfigToFile(ctx context.Context, req *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveProxyConfigToFile not implemented")
}

func RegisterAnywhereServerServer(s *grpc.Server, srv AnywhereServerServer) {
	s.RegisterService(&_AnywhereServer_serviceDesc, srv)
}

func _AnywhereServer_ListAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).ListAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/ListAgent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).ListAgent(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnywhereServer_AddProxyConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddProxyConfigInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).AddProxyConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/AddProxyConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).AddProxyConfig(ctx, req.(*AddProxyConfigInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnywhereServer_ListProxyConfigs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).ListProxyConfigs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/ListProxyConfigs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).ListProxyConfigs(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnywhereServer_RemoveProxyConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveProxyConfigInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).RemoveProxyConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/RemoveProxyConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).RemoveProxyConfig(ctx, req.(*RemoveProxyConfigInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnywhereServer_LoadProxyConfigFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).LoadProxyConfigFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/LoadProxyConfigFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).LoadProxyConfigFile(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnywhereServer_SaveProxyConfigToFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnywhereServerServer).SaveProxyConfigToFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/anywhereRpc.AnywhereServer/SaveProxyConfigToFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnywhereServerServer).SaveProxyConfigToFile(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _AnywhereServer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "anywhereRpc.AnywhereServer",
	HandlerType: (*AnywhereServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListAgent",
			Handler:    _AnywhereServer_ListAgent_Handler,
		},
		{
			MethodName: "AddProxyConfig",
			Handler:    _AnywhereServer_AddProxyConfig_Handler,
		},
		{
			MethodName: "ListProxyConfigs",
			Handler:    _AnywhereServer_ListProxyConfigs_Handler,
		},
		{
			MethodName: "RemoveProxyConfig",
			Handler:    _AnywhereServer_RemoveProxyConfig_Handler,
		},
		{
			MethodName: "LoadProxyConfigFile",
			Handler:    _AnywhereServer_LoadProxyConfigFile_Handler,
		},
		{
			MethodName: "SaveProxyConfigToFile",
			Handler:    _AnywhereServer_SaveProxyConfigToFile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "anywhere/server/rpc/definitions/anywhereServer.proto",
}
