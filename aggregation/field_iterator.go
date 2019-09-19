package aggregation

import (
	"math"

	"github.com/lindb/lindb/pkg/bit"
	"github.com/lindb/lindb/pkg/collections"
	"github.com/lindb/lindb/pkg/encoding"
	"github.com/lindb/lindb/pkg/stream"
	"github.com/lindb/lindb/series"
	"github.com/lindb/lindb/series/field"
)

type fieldIterator struct {
	id        uint16
	name      string
	fieldType field.Type

	familyStartTime int64
	startSlot       int

	length int
	idx    int
	its    []series.PrimitiveIterator
}

func newFieldIterator(id uint16, name string, fieldType field.Type,
	familyStartTime int64, startSlot int, its []series.PrimitiveIterator) series.FieldIterator {
	return &fieldIterator{
		id:              id,
		name:            name,
		fieldType:       fieldType,
		familyStartTime: familyStartTime,
		startSlot:       startSlot,
		its:             its,
		length:          len(its),
	}
}

func (it *fieldIterator) FieldMeta() field.Meta {
	return field.Meta{ID: it.id, Name: it.name, Type: it.fieldType}
}

func (it *fieldIterator) HasNext() bool {
	return it.idx < it.length
}

func (it *fieldIterator) Next() series.PrimitiveIterator {
	if it.idx >= it.length {
		return nil
	}
	primitiveIt := it.its[it.idx]
	it.idx++
	return primitiveIt
}

//FIXME stone1100 need refactor
func (it *fieldIterator) Bytes() ([]byte, error) {
	writer := stream.NewBufferWriter(nil)

	writer.PutVarint64(it.familyStartTime)

	for it.HasNext() {
		primitiveIt := it.Next()
		encoder := encoding.NewTSDEncoder(it.startSlot)
		idx := 0
		for primitiveIt.HasNext() {
			slot, value := primitiveIt.Next()
			for slot > idx {
				encoder.AppendTime(bit.Zero)
				idx++
			}
			encoder.AppendTime(bit.One)
			encoder.AppendValue(math.Float64bits(value))
			idx++
		}
		data, err := encoder.Bytes()
		if err != nil {
			return nil, err
		}
		writer.PutUInt16(primitiveIt.FieldID())
		writer.PutVarint32(int32(len(data)))
		writer.PutBytes(data)
	}
	return writer.Bytes()
}

func (it *fieldIterator) SegmentStartTime() int64 {
	return it.familyStartTime
}

// primitiveIterator represents primitive iterator using array
type primitiveIterator struct {
	id uint16
	it collections.FloatArrayIterator
}

// newPrimitiveIterator create primitive iterator using array
func newPrimitiveIterator(id uint16, values collections.FloatArray) series.PrimitiveIterator {
	it := &primitiveIterator{
		id: id,
	}
	if values != nil {
		it.it = values.Iterator()
	}
	return it
}

// ID returns the primitive field id
func (it *primitiveIterator) FieldID() uint16 {
	return it.id
}

// HasNext returns if the iteration has more data points
func (it *primitiveIterator) HasNext() bool {
	if it.it == nil {
		return false
	}
	return it.it.HasNext()
}

// Next returns the data point in the iteration
func (it *primitiveIterator) Next() (timeSlot int, value float64) {
	if it.it == nil {
		return -1, 0
	}
	return it.it.Next()
}
