package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Resp struct {
	reader *bufio.Reader
}

type Writer struct {
	writer io.Writer
}

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte() // read the input byte-by-byte and store in a variable.
		if err != nil {
			return nil, 0, err
		}
		n += 1                 // n keeps track of the bytes written, cuz only one byte is writtern in every loop.
		line = append(line, b) // line gathers everything b sends in each iteration.
		// break out of the loop once we reached the last time \r appears.
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	// return everything from the first byte to /r, number of bytes and error.
	return line[:len(line)-2], n, nil
}

// returns the length of the input, bytes and error
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, err
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.array = make([]Value, length)
	for i := range length {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.array[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "array"

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, length)

	r.reader.Read(bulk)

	v.bulk = string(bulk)

	r.readLine()

	return v, nil
}

func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.MarshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "error":
		return v.marshalError()
	case "null":
		return v.marshalNull()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) MarshalArray() []byte {
	length := len(v.array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')

	for i := range length {
		bytes = append(bytes, v.array[i].Marshal()...)
	}
	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
