// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sql/contentionpb/contention.proto

package contentionpb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import github_com_cockroachdb_cockroach_pkg_sql_catalog_descpb "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
import time "time"
import github_com_cockroachdb_cockroach_pkg_roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
import github_com_cockroachdb_cockroach_pkg_util_uuid "github.com/cockroachdb/cockroach/pkg/util/uuid"

import github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// IndexContentionEvents describes all of the available contention information
// about a single index.
type IndexContentionEvents struct {
	// TableID is the ID of the table experiencing contention.
	TableID github_com_cockroachdb_cockroach_pkg_sql_catalog_descpb.ID `protobuf:"varint,1,opt,name=table_id,json=tableId,proto3,casttype=github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb.ID" json:"table_id,omitempty"`
	// IndexID is the ID of the index experiencing contention.
	IndexID github_com_cockroachdb_cockroach_pkg_sql_catalog_descpb.IndexID `protobuf:"varint,2,opt,name=index_id,json=indexId,proto3,casttype=github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb.IndexID" json:"index_id,omitempty"`
	// NumContentionEvents is the number of contention events that have happened
	// on the index.
	NumContentionEvents uint64 `protobuf:"varint,3,opt,name=num_contention_events,json=numContentionEvents,proto3" json:"num_contention_events,omitempty"`
	// CumulativeContentionTime is the total duration that transactions touching
	// the index have spent contended.
	CumulativeContentionTime time.Duration `protobuf:"bytes,4,opt,name=cumulative_contention_time,json=cumulativeContentionTime,proto3,stdduration" json:"cumulative_contention_time"`
	// Events are all contention events on the index that we kept track of. Note
	// that some events could have been forgotten since we're keeping a limited
	// LRU cache of them.
	//
	// The events are ordered by the key.
	Events []SingleKeyContention `protobuf:"bytes,5,rep,name=events,proto3" json:"events"`
}

func (m *IndexContentionEvents) Reset()      { *m = IndexContentionEvents{} }
func (*IndexContentionEvents) ProtoMessage() {}
func (*IndexContentionEvents) Descriptor() ([]byte, []int) {
	return fileDescriptor_contention_38bb230587b267d1, []int{0}
}
func (m *IndexContentionEvents) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IndexContentionEvents) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *IndexContentionEvents) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexContentionEvents.Merge(dst, src)
}
func (m *IndexContentionEvents) XXX_Size() int {
	return m.Size()
}
func (m *IndexContentionEvents) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexContentionEvents.DiscardUnknown(m)
}

var xxx_messageInfo_IndexContentionEvents proto.InternalMessageInfo

// SingleKeyContention describes all of the available contention information for
// a single key.
type SingleKeyContention struct {
	// Key is the key that other transactions conflicted on.
	Key github_com_cockroachdb_cockroach_pkg_roachpb.Key `protobuf:"bytes,1,opt,name=key,proto3,casttype=github.com/cockroachdb/cockroach/pkg/roachpb.Key" json:"key,omitempty"`
	// Txns are all contending transactions that we kept track of. Note that some
	// transactions could have been forgotten since we're keeping a limited LRU
	// cache of them.
	Txns []SingleKeyContention_SingleTxnContention `protobuf:"bytes,2,rep,name=txns,proto3" json:"txns"`
}

func (m *SingleKeyContention) Reset()      { *m = SingleKeyContention{} }
func (*SingleKeyContention) ProtoMessage() {}
func (*SingleKeyContention) Descriptor() ([]byte, []int) {
	return fileDescriptor_contention_38bb230587b267d1, []int{1}
}
func (m *SingleKeyContention) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SingleKeyContention) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *SingleKeyContention) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SingleKeyContention.Merge(dst, src)
}
func (m *SingleKeyContention) XXX_Size() int {
	return m.Size()
}
func (m *SingleKeyContention) XXX_DiscardUnknown() {
	xxx_messageInfo_SingleKeyContention.DiscardUnknown(m)
}

var xxx_messageInfo_SingleKeyContention proto.InternalMessageInfo

// SingleTxnContention describes a single transaction that contended with the
// key.
type SingleKeyContention_SingleTxnContention struct {
	// TxnID is the contending transaction.
	TxnID github_com_cockroachdb_cockroach_pkg_util_uuid.UUID `protobuf:"bytes,2,opt,name=txn_ids,json=txnIds,proto3,customtype=github.com/cockroachdb/cockroach/pkg/util/uuid.UUID" json:"txn_ids"`
	// Count is the number of times the corresponding transaction was
	// encountered.
	Count uint64 `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
}

func (m *SingleKeyContention_SingleTxnContention) Reset() {
	*m = SingleKeyContention_SingleTxnContention{}
}
func (*SingleKeyContention_SingleTxnContention) ProtoMessage() {}
func (*SingleKeyContention_SingleTxnContention) Descriptor() ([]byte, []int) {
	return fileDescriptor_contention_38bb230587b267d1, []int{1, 0}
}
func (m *SingleKeyContention_SingleTxnContention) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SingleKeyContention_SingleTxnContention) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *SingleKeyContention_SingleTxnContention) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SingleKeyContention_SingleTxnContention.Merge(dst, src)
}
func (m *SingleKeyContention_SingleTxnContention) XXX_Size() int {
	return m.Size()
}
func (m *SingleKeyContention_SingleTxnContention) XXX_DiscardUnknown() {
	xxx_messageInfo_SingleKeyContention_SingleTxnContention.DiscardUnknown(m)
}

var xxx_messageInfo_SingleKeyContention_SingleTxnContention proto.InternalMessageInfo

func init() {
	proto.RegisterType((*IndexContentionEvents)(nil), "cockroach.sql.contentionpb.IndexContentionEvents")
	proto.RegisterType((*SingleKeyContention)(nil), "cockroach.sql.contentionpb.SingleKeyContention")
	proto.RegisterType((*SingleKeyContention_SingleTxnContention)(nil), "cockroach.sql.contentionpb.SingleKeyContention.SingleTxnContention")
}
func (m *IndexContentionEvents) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IndexContentionEvents) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.TableID != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintContention(dAtA, i, uint64(m.TableID))
	}
	if m.IndexID != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintContention(dAtA, i, uint64(m.IndexID))
	}
	if m.NumContentionEvents != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintContention(dAtA, i, uint64(m.NumContentionEvents))
	}
	dAtA[i] = 0x22
	i++
	i = encodeVarintContention(dAtA, i, uint64(github_com_gogo_protobuf_types.SizeOfStdDuration(m.CumulativeContentionTime)))
	n1, err := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.CumulativeContentionTime, dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if len(m.Events) > 0 {
		for _, msg := range m.Events {
			dAtA[i] = 0x2a
			i++
			i = encodeVarintContention(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *SingleKeyContention) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SingleKeyContention) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Key) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintContention(dAtA, i, uint64(len(m.Key)))
		i += copy(dAtA[i:], m.Key)
	}
	if len(m.Txns) > 0 {
		for _, msg := range m.Txns {
			dAtA[i] = 0x12
			i++
			i = encodeVarintContention(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *SingleKeyContention_SingleTxnContention) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SingleKeyContention_SingleTxnContention) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x12
	i++
	i = encodeVarintContention(dAtA, i, uint64(m.TxnID.Size()))
	n2, err := m.TxnID.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if m.Count != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintContention(dAtA, i, uint64(m.Count))
	}
	return i, nil
}

func encodeVarintContention(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *IndexContentionEvents) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.TableID != 0 {
		n += 1 + sovContention(uint64(m.TableID))
	}
	if m.IndexID != 0 {
		n += 1 + sovContention(uint64(m.IndexID))
	}
	if m.NumContentionEvents != 0 {
		n += 1 + sovContention(uint64(m.NumContentionEvents))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.CumulativeContentionTime)
	n += 1 + l + sovContention(uint64(l))
	if len(m.Events) > 0 {
		for _, e := range m.Events {
			l = e.Size()
			n += 1 + l + sovContention(uint64(l))
		}
	}
	return n
}

func (m *SingleKeyContention) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Key)
	if l > 0 {
		n += 1 + l + sovContention(uint64(l))
	}
	if len(m.Txns) > 0 {
		for _, e := range m.Txns {
			l = e.Size()
			n += 1 + l + sovContention(uint64(l))
		}
	}
	return n
}

func (m *SingleKeyContention_SingleTxnContention) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.TxnID.Size()
	n += 1 + l + sovContention(uint64(l))
	if m.Count != 0 {
		n += 1 + sovContention(uint64(m.Count))
	}
	return n
}

func sovContention(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozContention(x uint64) (n int) {
	return sovContention(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *IndexContentionEvents) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContention
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IndexContentionEvents: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IndexContentionEvents: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TableID", wireType)
			}
			m.TableID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TableID |= (github_com_cockroachdb_cockroach_pkg_sql_catalog_descpb.ID(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IndexID", wireType)
			}
			m.IndexID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.IndexID |= (github_com_cockroachdb_cockroach_pkg_sql_catalog_descpb.IndexID(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumContentionEvents", wireType)
			}
			m.NumContentionEvents = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumContentionEvents |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CumulativeContentionTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthContention
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.CumulativeContentionTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Events", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthContention
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Events = append(m.Events, SingleKeyContention{})
			if err := m.Events[len(m.Events)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipContention(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthContention
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SingleKeyContention) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContention
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SingleKeyContention: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SingleKeyContention: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthContention
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Key = append(m.Key[:0], dAtA[iNdEx:postIndex]...)
			if m.Key == nil {
				m.Key = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Txns", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthContention
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Txns = append(m.Txns, SingleKeyContention_SingleTxnContention{})
			if err := m.Txns[len(m.Txns)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipContention(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthContention
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SingleKeyContention_SingleTxnContention) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContention
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SingleTxnContention: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SingleTxnContention: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TxnID", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthContention
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TxnID.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Count", wireType)
			}
			m.Count = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContention
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Count |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipContention(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthContention
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipContention(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowContention
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowContention
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowContention
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthContention
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowContention
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipContention(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthContention = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowContention   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("sql/contentionpb/contention.proto", fileDescriptor_contention_38bb230587b267d1)
}

var fileDescriptor_contention_38bb230587b267d1 = []byte{
	// 520 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0x3f, 0x6f, 0xd3, 0x40,
	0x14, 0xb7, 0xf3, 0x5f, 0xd7, 0xb0, 0xb8, 0xad, 0x14, 0x32, 0xd8, 0xa1, 0x53, 0xa6, 0x33, 0x4a,
	0x99, 0xba, 0x20, 0xb9, 0x06, 0xc9, 0xaa, 0x60, 0x30, 0xe9, 0x82, 0x54, 0x45, 0xb6, 0xef, 0x70,
	0x4f, 0xb1, 0xef, 0xd2, 0xf8, 0xae, 0x72, 0xbe, 0x45, 0xc7, 0x8e, 0x7c, 0x9c, 0x8c, 0x1d, 0x40,
	0xaa, 0x18, 0x0c, 0x38, 0xdf, 0xa2, 0x13, 0xf2, 0xd9, 0xad, 0x83, 0x0a, 0x52, 0x61, 0xb1, 0xde,
	0xb3, 0xde, 0xfb, 0xfd, 0xbb, 0x07, 0x5e, 0x24, 0x17, 0x91, 0x19, 0x30, 0xca, 0x31, 0xe5, 0x84,
	0xd1, 0x85, 0xbf, 0xd5, 0xc0, 0xc5, 0x92, 0x71, 0xa6, 0x0d, 0x03, 0x16, 0xcc, 0x97, 0xcc, 0x0b,
	0xce, 0x61, 0x72, 0x11, 0xc1, 0xed, 0xe1, 0xe1, 0x5e, 0xc8, 0x42, 0x26, 0xc7, 0xcc, 0xa2, 0x2a,
	0x37, 0x86, 0x7a, 0xc8, 0x58, 0x18, 0x61, 0x53, 0x76, 0xbe, 0xf8, 0x64, 0x22, 0xb1, 0xf4, 0x6a,
	0xc4, 0x83, 0x2f, 0x4d, 0xb0, 0xef, 0x50, 0x84, 0xd3, 0xe3, 0x07, 0xac, 0x37, 0x97, 0x98, 0xf2,
	0x44, 0x43, 0xa0, 0xc7, 0x3d, 0x3f, 0xc2, 0x33, 0x82, 0x06, 0xea, 0x48, 0x1d, 0x3f, 0xb3, 0x9c,
	0x3c, 0x33, 0xba, 0xd3, 0xe2, 0x9f, 0x63, 0xdf, 0x65, 0xc6, 0x51, 0x48, 0xf8, 0xb9, 0xf0, 0x61,
	0xc0, 0x62, 0xf3, 0x41, 0x17, 0xf2, 0xeb, 0xda, 0x5c, 0xcc, 0x43, 0x53, 0x9a, 0xf2, 0xb8, 0x17,
	0xb1, 0xd0, 0x44, 0x38, 0x09, 0x16, 0x3e, 0x74, 0x6c, 0xb7, 0x2b, 0xa1, 0x1d, 0xa4, 0x11, 0xd0,
	0x23, 0x05, 0x7d, 0xc1, 0xd2, 0x90, 0x2c, 0xef, 0x0b, 0x16, 0x29, 0x49, 0xb2, 0xbc, 0xfe, 0x6f,
	0x96, 0x12, 0xc2, 0xed, 0x4a, 0x7c, 0x07, 0x69, 0x13, 0xb0, 0x4f, 0x45, 0x3c, 0xab, 0x43, 0x9b,
	0x61, 0xe9, 0x74, 0xd0, 0x1c, 0xa9, 0xe3, 0x96, 0xbb, 0x4b, 0x45, 0xfc, 0x28, 0x04, 0x0f, 0x0c,
	0x03, 0x11, 0x8b, 0xc8, 0xe3, 0xe4, 0x12, 0x6f, 0xaf, 0x72, 0x12, 0xe3, 0x41, 0x6b, 0xa4, 0x8e,
	0x77, 0x26, 0xcf, 0x61, 0x99, 0x31, 0xbc, 0xcf, 0x18, 0xda, 0x55, 0xc6, 0x56, 0x6f, 0x9d, 0x19,
	0xca, 0xf5, 0x77, 0x43, 0x75, 0x07, 0x35, 0x4c, 0x4d, 0x32, 0x25, 0x31, 0xd6, 0xde, 0x81, 0x4e,
	0xa5, 0xa3, 0x3d, 0x6a, 0x8e, 0x77, 0x26, 0x26, 0xfc, 0xfb, 0x23, 0xc3, 0x0f, 0x84, 0x86, 0x11,
	0x3e, 0xc1, 0xab, 0x1a, 0xc4, 0x6a, 0x15, 0x24, 0x6e, 0x05, 0x72, 0xd4, 0xba, 0xfe, 0x6c, 0x28,
	0x07, 0x5f, 0x1b, 0x60, 0xf7, 0x0f, 0xb3, 0xda, 0x5b, 0xd0, 0x9c, 0xe3, 0x95, 0x7c, 0xcf, 0xbe,
	0xf5, 0xea, 0x2e, 0x33, 0x5e, 0x3e, 0x29, 0x5e, 0x59, 0x2d, 0x7c, 0x78, 0x82, 0x57, 0x6e, 0x01,
	0xa0, 0x9d, 0x81, 0x16, 0x4f, 0x69, 0x32, 0x68, 0x48, 0xc9, 0xc7, 0xff, 0x28, 0xb9, 0xfa, 0x37,
	0x4d, 0xe9, 0x23, 0x1b, 0x12, 0x76, 0x78, 0xa5, 0xde, 0xcb, 0xff, 0x6d, 0x46, 0x3b, 0x03, 0x5d,
	0x9e, 0xd2, 0x19, 0x41, 0x89, 0x3c, 0x96, 0xbe, 0x65, 0x17, 0x4b, 0xdf, 0x32, 0xe3, 0xf0, 0x49,
	0x36, 0x04, 0x27, 0x91, 0x29, 0x04, 0x41, 0xf0, 0xf4, 0xd4, 0xb1, 0xf3, 0xcc, 0x68, 0x4f, 0x53,
	0xea, 0xd8, 0x6e, 0x87, 0xa7, 0xd4, 0x41, 0x89, 0xb6, 0x07, 0xda, 0x01, 0x13, 0x94, 0x57, 0x17,
	0x51, 0x36, 0x65, 0xa2, 0xe5, 0xd7, 0x82, 0xeb, 0x9f, 0xba, 0xb2, 0xce, 0x75, 0xf5, 0x26, 0xd7,
	0xd5, 0xdb, 0x5c, 0x57, 0x7f, 0xe4, 0xba, 0x7a, 0xb5, 0xd1, 0x95, 0x9b, 0x8d, 0xae, 0xdc, 0x6e,
	0x74, 0xe5, 0x63, 0x7f, 0xdb, 0xbc, 0xdf, 0x91, 0x37, 0x71, 0xf8, 0x2b, 0x00, 0x00, 0xff, 0xff,
	0xf4, 0xab, 0x85, 0x91, 0xdc, 0x03, 0x00, 0x00,
}
