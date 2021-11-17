package cmd

import (
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type metaObjectEncoder = func(object metav1.Object) ([]byte, error)

type encoder struct {
	count   int
	w       io.Writer
	marshal metaObjectEncoder
}

var separator = []byte("---\n")

func (e *encoder) Encode(obj metav1.Object) error {
	data, err := e.marshal(obj)
	if err != nil {
		return err
	}
	if e.count > 0 {
		_, _ = e.w.Write(separator)
	}
	_, _ = e.w.Write(data)
	e.count++
	return nil
}

func newEncoder(w io.Writer, marshal metaObjectEncoder) *encoder {
	return &encoder{w: w, marshal: marshal}
}
