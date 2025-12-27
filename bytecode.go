package main

import (
	"fmt"
	"io"
	"math"
	"os"
)

const (
	MagicHeader           = 0x4C4C4243
	VersionMajor    uint8 = 3
	VersionMinor    uint8 = 0
	VersionCombined       = (VersionMajor << 4) | (VersionMinor & 0x0F)

	ConstTypeNumber   = 0
	ConstTypeString   = 1
	ConstTypeFuncPtr  = 2
	ConstTypeBool     = 3
	ConstTypeNil      = 4
	ConstFlagSmallInt = 1 << 0
	ConstFlagShortStr = 1 << 1

	ArgTypeConst  = 0
	ArgTypeInt    = 1
	ArgTypeFloat  = 2
	ArgTypeString = 3
)

type BitWriter struct {
	writer io.Writer
	buffer byte
	bitPos uint8
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{writer: w}
}

func (bw *BitWriter) WriteBits(value uint64, bits uint8) error {
	for i := uint8(0); i < bits; i++ {
		bit := (value >> i) & 1
		bw.buffer |= byte(bit << bw.bitPos)
		bw.bitPos++

		if bw.bitPos == 8 {
			if _, err := bw.writer.Write([]byte{bw.buffer}); err != nil {
				return err
			}
			bw.buffer = 0
			bw.bitPos = 0
		}
	}
	return nil
}

func (bw *BitWriter) Flush() error {
	if bw.bitPos > 0 {
		_, err := bw.writer.Write([]byte{bw.buffer})
		bw.bitPos = 0
		bw.buffer = 0
		return err
	}
	return nil
}

func (bw *BitWriter) WriteUint32(val uint32) error {
	return bw.WriteBits(uint64(val), 32)
}

func (bw *BitWriter) WriteUint8(val uint8) error {
	return bw.WriteBits(uint64(val), 8)
}

func (bw *BitWriter) WriteVarUint(val uint32) error {
	for val >= 0x80 {
		if err := bw.WriteBits(uint64(val&0x7F)|0x80, 8); err != nil {
			return err
		}
		val >>= 7
	}
	return bw.WriteBits(uint64(val), 8)
}

func (bw *BitWriter) WriteVarUint16(val uint16) error {
	if val < 0x80 {
		return bw.WriteBits(uint64(val), 8)
	}
	if err := bw.WriteBits(uint64(val&0x7F)|0x80, 8); err != nil {
		return err
	}
	return bw.WriteBits(uint64(val>>7), 8)
}

func (bw *BitWriter) WriteVarInt(val int32) error {
	uval := uint32(val) << 1
	if val < 0 {
		uval = ^uval
	}
	return bw.WriteVarUint(uval)
}

type BitReader struct {
	reader io.Reader
	buffer byte
	bitPos uint8
	eof    bool
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{reader: r}
}

func (br *BitReader) ReadBits(bits uint8) (uint64, error) {
	var result uint64
	for i := uint8(0); i < bits; i++ {
		if br.bitPos == 0 && !br.eof {
			var buf [1]byte
			n, err := br.reader.Read(buf[:])
			if err != nil && err != io.EOF {
				return 0, err
			}
			if n == 0 {
				br.eof = true
				return 0, io.ErrUnexpectedEOF
			}
			br.buffer = buf[0]
		}

		bit := (br.buffer >> br.bitPos) & 1
		result |= uint64(bit) << i
		br.bitPos = (br.bitPos + 1) % 8
	}
	return result, nil
}

func (br *BitReader) ReadUint32() (uint32, error) {
	val, err := br.ReadBits(32)
	return uint32(val), err
}

func (br *BitReader) ReadUint8() (uint8, error) {
	val, err := br.ReadBits(8)
	return uint8(val), err
}

func (br *BitReader) ReadVarUint() (uint32, error) {
	var result uint32
	var shift uint
	for {
		b, err := br.ReadUint8()
		if err != nil {
			return 0, err
		}
		result |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, nil
}

func (br *BitReader) ReadVarUint16() (uint16, error) {
	first, err := br.ReadUint8()
	if err != nil {
		return 0, err
	}
	if first < 0x80 {
		return uint16(first), nil
	}
	second, err := br.ReadUint8()
	if err != nil {
		return 0, err
	}
	return uint16(first&0x7F) | (uint16(second) << 7), nil
}

type BytecodeWriter struct {
	bitWriter *BitWriter
}

func NewBytecodeWriter(w io.Writer) *BytecodeWriter {
	return &BytecodeWriter{
		bitWriter: NewBitWriter(w),
	}
}

func (bw *BytecodeWriter) WriteBytecode(instructions []Instruction, constants []Constant) error {
	if err := bw.bitWriter.WriteUint32(MagicHeader); err != nil {
		return err
	}

	if err := bw.bitWriter.WriteUint8(VersionCombined); err != nil {
		return err
	}

	if err := bw.bitWriter.WriteVarUint(uint32(len(constants))); err != nil {
		return err
	}

	if err := bw.bitWriter.WriteVarUint(uint32(len(instructions))); err != nil {
		return err
	}

	for _, c := range constants {
		switch c.Type {
		case "number":
			if val, ok := c.Value.(int); ok && val >= -64 && val <= 63 {
				if err := bw.bitWriter.WriteBits(uint64(ConstTypeNumber), 3); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteBits(1, 1); err != nil {
					return err
				}
				signedVal := int8(val)
				if err := bw.bitWriter.WriteBits(uint64(signedVal)&0x7F, 7); err != nil {
					return err
				}
			} else {
				if err := bw.bitWriter.WriteBits(uint64(ConstTypeNumber), 3); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteBits(0, 1); err != nil {
					return err
				}

				var fval float64
				if val, ok := c.Value.(float64); ok {
					fval = val
				} else if val, ok := c.Value.(int); ok {
					fval = float64(val)
				}

				bits := math.Float64bits(fval)
				for i := 0; i < 64; i++ {
					bit := (bits >> i) & 1
					if err := bw.bitWriter.WriteBits(bit, 1); err != nil {
						return err
					}
				}
			}

		case "string":
			str := c.Value.(string)
			if len(str) <= 255 {
				if err := bw.bitWriter.WriteBits(uint64(ConstTypeString), 3); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteBits(1, 1); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteBits(uint64(len(str)), 8); err != nil {
					return err
				}
			} else {
				if err := bw.bitWriter.WriteBits(uint64(ConstTypeString), 3); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteBits(0, 1); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteVarUint(uint32(len(str))); err != nil {
					return err
				}
			}

			for _, ch := range []byte(str) {
				if err := bw.bitWriter.WriteBits(uint64(ch), 8); err != nil {
					return err
				}
			}

		case "funcptr":
			if err := bw.bitWriter.WriteBits(uint64(ConstTypeFuncPtr), 3); err != nil {
				return err
			}
			var val uint32
			if v, ok := c.Value.(float64); ok {
				val = uint32(v)
			} else if v, ok := c.Value.(int); ok {
				val = uint32(v)
			}
			if err := bw.bitWriter.WriteVarUint(val); err != nil {
				return err
			}

		case "bool":
			if err := bw.bitWriter.WriteBits(uint64(ConstTypeBool), 3); err != nil {
				return err
			}
			var val uint64 = 0
			if c.Value == true {
				val = 1
			}
			if err := bw.bitWriter.WriteBits(val, 1); err != nil {
				return err
			}

		case "nil":
			if err := bw.bitWriter.WriteBits(uint64(ConstTypeNil), 3); err != nil {
				return err
			}
		}
	}

	for _, inst := range instructions {
		opcode := uint64(inst.Op) & 0x7F
		hasArg := inst.Arg != nil
		if hasArg {
			opcode |= 0x80
		}
		if err := bw.bitWriter.WriteBits(opcode, 8); err != nil {
			return err
		}

		if err := bw.bitWriter.WriteVarUint16(uint16(inst.Line)); err != nil {
			return err
		}

		if hasArg {
			var argType uint64

			switch arg := inst.Arg.(type) {
			case float64:
				if arg == float64(int32(arg)) {
					argType = ArgTypeInt
					val := int32(arg)
					if err := bw.bitWriter.WriteBits(argType, 2); err != nil {
						return err
					}
					if err := bw.bitWriter.WriteVarInt(val); err != nil {
						return err
					}
				} else {
					argType = ArgTypeFloat
					if err := bw.bitWriter.WriteBits(argType, 2); err != nil {
						return err
					}
					bits := math.Float64bits(arg)
					for i := 0; i < 64; i++ {
						bit := (bits >> i) & 1
						if err := bw.bitWriter.WriteBits(bit, 1); err != nil {
							return err
						}
					}
				}
				continue

			case int:
				argType = ArgTypeInt
				if err := bw.bitWriter.WriteBits(argType, 2); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteVarInt(int32(arg)); err != nil {
					return err
				}
				continue

			case string:
				argType = ArgTypeString
				if err := bw.bitWriter.WriteBits(argType, 2); err != nil {
					return err
				}
				if err := bw.bitWriter.WriteVarUint(uint32(len(arg))); err != nil {
					return err
				}
				for _, ch := range []byte(arg) {
					if err := bw.bitWriter.WriteBits(uint64(ch), 8); err != nil {
						return err
					}
				}
				continue

			default:
				if f, ok := arg.(float64); ok {
					argType = ArgTypeConst
					if err := bw.bitWriter.WriteBits(argType, 2); err != nil {
						return err
					}
					if err := bw.bitWriter.WriteVarUint(uint32(f)); err != nil {
						return err
					}
				}
			}
		}
	}

	return bw.bitWriter.Flush()
}

type BytecodeReader struct {
	bitReader *BitReader
}

func NewBytecodeReader(r io.Reader) *BytecodeReader {
	return &BytecodeReader{
		bitReader: NewBitReader(r),
	}
}

func (br *BytecodeReader) ReadBytecode() ([]Instruction, []Constant, error) {
	magic, err := br.bitReader.ReadUint32()
	if err != nil {
		return nil, nil, err
	}
	if magic != MagicHeader {
		return nil, nil, fmt.Errorf("invalid bytecode file: bad magic")
	}

	version, err := br.bitReader.ReadUint8()
	if err != nil {
		return nil, nil, err
	}
	major := version >> 4
	minor := version & 0x0F
	if major != VersionMajor {
		return nil, nil, fmt.Errorf("incompatible bytecode version: %d.%d", major, minor)
	}

	constantCount, err := br.bitReader.ReadVarUint()
	if err != nil {
		return nil, nil, err
	}
	instructionCount, err := br.bitReader.ReadVarUint()
	if err != nil {
		return nil, nil, err
	}

	constants := make([]Constant, constantCount)
	for i := range constants {
		constType, err := br.bitReader.ReadBits(3)
		if err != nil {
			return nil, nil, err
		}

		switch uint8(constType) {
		case ConstTypeNumber:
			isSmall, err := br.bitReader.ReadBits(1)
			if err != nil {
				return nil, nil, err
			}

			if isSmall == 1 {
				valBits, err := br.bitReader.ReadBits(7)
				if err != nil {
					return nil, nil, err
				}
				val := int8(valBits)
				if valBits&0x40 != 0 {
					val |= ^0x7F
				}
				constants[i] = Constant{Value: int(val), Type: "number"}
			} else {
				var bits uint64
				for i := 0; i < 64; i++ {
					bit, err := br.bitReader.ReadBits(1)
					if err != nil {
						return nil, nil, err
					}
					bits |= bit << i
				}
				val := math.Float64frombits(bits)
				constants[i] = Constant{Value: val, Type: "number"}
			}

		case ConstTypeString:
			isShort, err := br.bitReader.ReadBits(1)
			if err != nil {
				return nil, nil, err
			}

			var strLen uint32
			if isShort == 1 {
				lenBits, err := br.bitReader.ReadBits(8)
				if err != nil {
					return nil, nil, err
				}
				strLen = uint32(lenBits)
			} else {
				strLen, err = br.bitReader.ReadVarUint()
				if err != nil {
					return nil, nil, err
				}
			}

			strBytes := make([]byte, strLen)
			for j := range strBytes {
				ch, err := br.bitReader.ReadBits(8)
				if err != nil {
					return nil, nil, err
				}
				strBytes[j] = byte(ch)
			}
			constants[i] = Constant{Value: string(strBytes), Type: "string"}

		case ConstTypeFuncPtr:
			val, err := br.bitReader.ReadVarUint()
			if err != nil {
				return nil, nil, err
			}
			constants[i] = Constant{Value: float64(val), Type: "funcptr"}

		case ConstTypeBool:
			val, err := br.bitReader.ReadBits(1)
			if err != nil {
				return nil, nil, err
			}
			constants[i] = Constant{Value: val == 1, Type: "bool"}

		case ConstTypeNil:
			constants[i] = Constant{Value: nil, Type: "nil"}
		}
	}

	instructions := make([]Instruction, instructionCount)
	for i := range instructions {
		opcode, err := br.bitReader.ReadBits(8)
		if err != nil {
			return nil, nil, err
		}

		hasArg := (opcode & 0x80) != 0
		opcode &^= 0x80

		line, err := br.bitReader.ReadVarUint16()
		if err != nil {
			return nil, nil, err
		}

		var arg interface{}
		if hasArg {
			argType, err := br.bitReader.ReadBits(2)
			if err != nil {
				return nil, nil, err
			}

			switch argType {
			case ArgTypeConst:
				idx, err := br.bitReader.ReadVarUint()
				if err != nil {
					return nil, nil, err
				}
				arg = float64(idx)

			case ArgTypeInt:
				uval, err := br.bitReader.ReadVarUint()
				if err != nil {
					return nil, nil, err
				}
				val := int32(uval >> 1)
				if (uval & 1) != 0 {
					val = ^val
				}
				arg = float64(val)

			case ArgTypeFloat:
				var bits uint64
				for i := 0; i < 64; i++ {
					bit, err := br.bitReader.ReadBits(1)
					if err != nil {
						return nil, nil, err
					}
					bits |= bit << i
				}
				arg = math.Float64frombits(bits)

			case ArgTypeString:
				strLen, err := br.bitReader.ReadVarUint()
				if err != nil {
					return nil, nil, err
				}
				strBytes := make([]byte, strLen)
				for j := range strBytes {
					ch, err := br.bitReader.ReadBits(8)
					if err != nil {
						return nil, nil, err
					}
					strBytes[j] = byte(ch)
				}
				arg = string(strBytes)
			}
		}

		instructions[i] = Instruction{
			Op:   OpCode(opcode),
			Arg:  arg,
			Line: int(line),
		}
	}

	return instructions, constants, nil
}

func SaveBytecode(filename string, instructions []Instruction, constants []Constant) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := NewBytecodeWriter(file)
	return writer.WriteBytecode(instructions, constants)
}

func LoadBytecode(filename string) ([]Instruction, []Constant, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := NewBytecodeReader(file)
	return reader.ReadBytecode()
}
