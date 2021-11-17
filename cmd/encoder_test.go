package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEncoder(t *testing.T) {
	r := require.New(t)
	var buf bytes.Buffer
	enc := newEncoder(&buf, func(object metav1.Object) ([]byte, error) {
		return []byte(object.GetName() + "\n"), nil
	})
	r.NoError(enc.Encode(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
	}))
	r.Equal("foo\n", buf.String())

	r.NoError(enc.Encode(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "bar"},
	}))
	r.Equal("foo\n---\nbar\n", buf.String())
}
