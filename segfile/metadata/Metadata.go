// automatically generated, do not modify

package metadata

import (
	flatbuffers "github.com/kadirahq/flatbuffers/go"
)
type Metadata struct {
	_tab flatbuffers.Table
}

func GetRootAsMetadata(buf []byte, offset flatbuffers.UOffsetT) *Metadata {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Metadata{}
	x.Init(buf, n + offset)
	return x
}

func (rcv *Metadata) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Metadata) Segs() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Metadata) MutateSegs(n int64) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		 return rcv._tab.MutateInt64(o + rcv._tab.Pos, n)
	}
	return false
}

func (rcv *Metadata) Size() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Metadata) MutateSize(n int64) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		 return rcv._tab.MutateInt64(o + rcv._tab.Pos, n)
	}
	return false
}

func (rcv *Metadata) Used() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Metadata) MutateUsed(n int64) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		 return rcv._tab.MutateInt64(o + rcv._tab.Pos, n)
	}
	return false
}

func MetadataStart(builder *flatbuffers.Builder) { builder.StartObject(3) }
func MetadataAddSegs(builder *flatbuffers.Builder, segs int64) { builder.PrependInt64Slot(0, segs, 0) }
func MetadataAddSize(builder *flatbuffers.Builder, size int64) { builder.PrependInt64Slot(1, size, 0) }
func MetadataAddUsed(builder *flatbuffers.Builder, used int64) { builder.PrependInt64Slot(2, used, 0) }
func MetadataEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT { return builder.EndObject() }
