package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
	"sigs.k8s.io/structured-merge-diff/v4/value"
	"sigs.k8s.io/yaml"
)

func getInfoOr(n *Node, defVal string) string {
	for n != nil {
		if n.Info != nil {
			return n.Info.String()
		}
		n = n.Parent
	}
	return defVal
}

// fieldsToSet creates a set paths from an input trie of fields
func fieldsToSet(f metav1.FieldsV1) (s fieldpath.Set, err error) {
	err = s.FromJSON(bytes.NewReader(f.Raw))
	return s, err
}

func fieldListID(fl value.FieldList) string {
	strs := make([]string, len(fl))
	for i, k := range fl {
		strs[i] = fmt.Sprintf("%v=%v", k.Name, value.ToString(k.Value))
	}
	return strings.Join(strs, ",")
}

func toYAMLString(s string) string {
	data, _ := yaml.Marshal(s)
	return string(data[:len(data)-1]) // remove the trailing newline
}

func onlyContainsNewline(s string) (onlyNewline bool) {
	sawNewline := false
	for _, ch := range s {
		if ch == '\n' {
			sawNewline = true
		} else if unicode.IsControl(ch) {
			return false
		}
	}
	return sawNewline == true
}

func toYAMLStringValueln(prefix, s string, lvl int) string {
	if onlyContainsNewline(s) {
		var b strings.Builder
		if strings.HasSuffix(s, "\n\n") {
			b.WriteString("|+\n")
		} else if strings.HasSuffix(s, "\n") {
			b.WriteString("|\n")
		} else {
			b.WriteString("|-\n")
		}
		s = strings.TrimSuffix(s, "\n")
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			if len(line) > 0 {
				b.WriteString(prefix)
				writeIndent(&b, lvl)
				b.WriteString(line)
			}
			b.WriteRune('\n')
		}
		return b.String()
	}

	return toYAMLString(s) + "\n"
}

func writeString(w io.Writer, v string) error {
	_, err := w.Write([]byte(v))
	return err
}

func writeBytes(w io.Writer, v ...byte) error {
	_, err := w.Write(v)
	return err
}

func formatBasicType(v interface{}) string {
	switch actual := v.(type) {
	case int:
		return strconv.FormatInt(int64(actual), 10)
	case int8:
		return strconv.FormatInt(int64(actual), 10)
	case int16:
		return strconv.FormatInt(int64(actual), 10)
	case int32:
		return strconv.FormatInt(int64(actual), 10)
	case int64:
		return strconv.FormatInt(actual, 10)
	case uint:
		return strconv.FormatUint(uint64(actual), 10)
	case uint8:
		return strconv.FormatUint(uint64(actual), 10)
	case uint16:
		return strconv.FormatUint(uint64(actual), 10)
	case uint32:
		return strconv.FormatUint(uint64(actual), 10)
	case uint64:
		return strconv.FormatUint(actual, 10)
	case bool:
		return strconv.FormatBool(actual)
	case float32:
		return strconv.FormatFloat(float64(actual), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(actual, 'f', -1, 64)
	default:
		panic("unreachable")
	}
}

func writeIndent(w io.Writer, lvl int) error {
	length := lvl * 2
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = ' '
	}
	_, err := w.Write(data)
	return err
}

func appendSpace(s string, totalLen int) string {
	var b strings.Builder
	b.Grow(totalLen)
	b.WriteString(s)
	n := totalLen - len(s)
	for i := 0; i < n; i++ {
		b.WriteByte(' ')
	}
	return b.String()
}

func fieldListMatchObject(fl value.FieldList, o map[string]interface{}) bool {
	for _, field := range fl {
		i, exist := o[field.Name]
		if !exist {
			return false
		}
		gotVal := value.NewValueInterface(i)
		if value.Compare(field.Value, gotVal) != 0 {
			return false
		}
	}
	return true
}
