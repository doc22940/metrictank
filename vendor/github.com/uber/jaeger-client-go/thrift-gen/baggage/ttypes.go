// Autogenerated by Thrift Compiler (0.9.3)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package baggage

import (
	"bytes"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

var GoUnusedProtection__ int

// Attributes:
//  - BaggageKey
//  - MaxValueLength
type BaggageRestriction struct {
	BaggageKey     string `thrift:"baggageKey,1,required" json:"baggageKey"`
	MaxValueLength int32  `thrift:"maxValueLength,2,required" json:"maxValueLength"`
}

func NewBaggageRestriction() *BaggageRestriction {
	return &BaggageRestriction{}
}

func (p *BaggageRestriction) GetBaggageKey() string {
	return p.BaggageKey
}

func (p *BaggageRestriction) GetMaxValueLength() int32 {
	return p.MaxValueLength
}
func (p *BaggageRestriction) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	var issetBaggageKey bool = false
	var issetMaxValueLength bool = false

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.readField1(iprot); err != nil {
				return err
			}
			issetBaggageKey = true
		case 2:
			if err := p.readField2(iprot); err != nil {
				return err
			}
			issetMaxValueLength = true
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	if !issetBaggageKey {
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("Required field BaggageKey is not set"))
	}
	if !issetMaxValueLength {
		return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("Required field MaxValueLength is not set"))
	}
	return nil
}

func (p *BaggageRestriction) readField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.BaggageKey = v
	}
	return nil
}

func (p *BaggageRestriction) readField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.MaxValueLength = v
	}
	return nil
}

func (p *BaggageRestriction) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("BaggageRestriction"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *BaggageRestriction) writeField1(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("baggageKey", thrift.STRING, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:baggageKey: ", p), err)
	}
	if err := oprot.WriteString(string(p.BaggageKey)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.baggageKey (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:baggageKey: ", p), err)
	}
	return err
}

func (p *BaggageRestriction) writeField2(oprot thrift.TProtocol) (err error) {
	if err := oprot.WriteFieldBegin("maxValueLength", thrift.I32, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:maxValueLength: ", p), err)
	}
	if err := oprot.WriteI32(int32(p.MaxValueLength)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.maxValueLength (2) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:maxValueLength: ", p), err)
	}
	return err
}

func (p *BaggageRestriction) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BaggageRestriction(%+v)", *p)
}
