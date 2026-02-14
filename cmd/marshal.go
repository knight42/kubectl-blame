package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

const timeLayout = "2006-01-02 15:04:05 -0700"

const (
	TimeFormatFull     = "full"
	TimeFormatRelative = "relative"
	TimeFormatNone     = "none"
)

var (
	bytesNull        = []byte("null\n")
	bytesEmptyObject = []byte("{}\n")
	bytesEmptyList   = []byte("[]\n")
)

type Marshaller struct {
	emptyInfo string

	now        time.Time
	timeFormat string
	colorizer  *Colorizer
}

type ManagerInfo struct {
	Manager   string
	Operation string
	Time      string
}

func (i *ManagerInfo) String() string {
	if len(i.Time) > 0 {
		return fmt.Sprintf("%s (%s %s) ", i.Manager, i.Operation, i.Time)
	}
	return fmt.Sprintf("%s %s ", i.Manager, i.Operation)
}

type KeyWithNode struct {
	Node *Node
	Key  value.FieldList
}

type ValueWithNode struct {
	Node  *Node
	Value value.Value
}

type Node struct {
	// field name -> child node
	Fields map[string]*Node
	// fieldList ID -> {fieldList, child node}
	Keys map[string]*KeyWithNode
	// value ID -> {value, child node}
	Values map[string]*ValueWithNode

	Parent *Node

	Managers []ManagerInfo
}

func (m *Marshaller) buildTree(managedFields []metav1.ManagedFieldsEntry, mgrMaxLength, opMaxLength, timeMaxLength int) (*Node, error) {
	root := &Node{}
	var timeFormatter func(*metav1.Time) string
	switch m.timeFormat {
	case TimeFormatFull:
		timeFormatter = func(t *metav1.Time) string {
			if t == nil {
				return ""
			}
			return t.Format(timeLayout)
		}
	case TimeFormatRelative:
		timeFormatter = func(t *metav1.Time) string {
			if t == nil {
				return ""
			}
			return duration.HumanDuration(m.now.Sub(t.Time))
		}
	case TimeFormatNone:
		timeFormatter = func(t *metav1.Time) string {
			return ""
		}
	default:
		return nil, fmt.Errorf("unknown time format: %s", m.timeFormat)
	}

	var infoLength int
	for _, field := range managedFields {
		manager := field.Manager
		operation := string(field.Operation)
		manager = appendSpace(manager, mgrMaxLength)
		operation = appendSpace(operation, opMaxLength)
		updatedAt := prependSpace(timeFormatter(field.Time), timeMaxLength)
		info := ManagerInfo{
			Manager:   manager,
			Operation: operation,
			Time:      updatedAt,
		}
		infoLength = len(info.String())

		s, err := fieldsToSet(*field.FieldsV1)
		if err != nil {
			return nil, err
		}

		var errList []error
		s.Iterate(func(path fieldpath.Path) {
			cur := root
			l := len(path)
			for i, ele := range path {
				isLeaf := i == l-1
				switch {
				case ele.FieldName != nil:
					name := *ele.FieldName
					if cur.Fields == nil {
						cur.Fields = map[string]*Node{}
					}
					if cur.Fields[name] == nil {
						cur.Fields[name] = &Node{Parent: cur}
					}
					cur = cur.Fields[name]
				case ele.Key != nil:
					name := fieldListID(*ele.Key)
					if cur.Keys == nil {
						cur.Keys = map[string]*KeyWithNode{}
					}
					if cur.Keys[name] == nil {
						cur.Keys[name] = &KeyWithNode{
							Node: &Node{Parent: cur},
							Key:  *ele.Key,
						}
					}
					cur = cur.Keys[name].Node
				case ele.Value != nil:
					name := value.ToString(*ele.Value)
					if cur.Values == nil {
						cur.Values = map[string]*ValueWithNode{}
					}
					if cur.Values[name] == nil {
						cur.Values[name] = &ValueWithNode{
							Value: *ele.Value,
							Node:  &Node{Parent: cur},
						}
					}
					cur = cur.Values[name].Node
				default:
					data, _ := json.Marshal(ele)
					errList = append(errList, fmt.Errorf("unknown element: %s", string(data)))
					continue
				}
				if isLeaf {
					cur.Managers = append(cur.Managers, info)
				}
			}
		})
		if len(errList) > 0 {
			return nil, errors.NewAggregate(errList)
		}
	}
	m.emptyInfo = strings.Repeat(" ", infoLength)
	return root, nil
}

func MarshalMetaObject(obj metav1.Object, timeFmt string, colorizer *Colorizer) ([]byte, error) {
	m := Marshaller{
		now:        time.Now(),
		timeFormat: timeFmt,
		colorizer:  colorizer,
	}
	return m.marshalMetaObject(obj)
}

func (m *Marshaller) marshalMetaObject(obj metav1.Object) ([]byte, error) {
	managedFields := obj.GetManagedFields()
	if len(managedFields) == 0 {
		return nil, fmt.Errorf(".metadata.managedFields is empty")
	}

	relativeTime := m.timeFormat == TimeFormatRelative
	var (
		mgrMaxLength, opMaxLength, timeMaxLength int
	)
	for _, field := range managedFields {
		if len(field.Manager) > mgrMaxLength {
			mgrMaxLength = len(field.Manager)
		}
		if len(field.Operation) > opMaxLength {
			opMaxLength = len(field.Operation)
		}

		if relativeTime && field.Time != nil {
			d := duration.HumanDuration(m.now.Sub(field.Time.Time))
			if len(d) > timeMaxLength {
				timeMaxLength = len(d)
			}
		}
	}
	if m.timeFormat == TimeFormatFull {
		timeMaxLength = len(timeLayout)
	}

	root, err := m.buildTree(managedFields, mgrMaxLength, opMaxLength, timeMaxLength)
	if err != nil {
		return nil, err
	}

	obj.SetManagedFields(nil)
	unsObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = m.marshalMapWithCtx(Context{
		Level:   0,
		NewLine: true,
		Node:    root,
	}, unsObj, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Marshaller) marshalMapWithCtx(ctx Context, o map[string]interface{}, w io.Writer) error {
	if o == nil {
		w.Write(bytesNull)
		return nil
	}
	if len(o) == 0 {
		w.Write(bytesEmptyObject)
		return nil
	}

	// make the result predictable
	keys := make([]string, 0, len(o))
	for key := range o {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	root := ctx.Node
	for i, key := range keys {
		val := o[key]
		child := root.Fields[key]
		if i == 0 && !ctx.NewLine {
			writeString(w, toYAMLString(key))
		} else {
			ok := child != nil
			if ok {
				info := getInfoOr(child, m.emptyInfo, m.colorizer)
				writeString(w, info)
			} else {
				info := getInfoOr(root, m.emptyInfo, m.colorizer)
				writeString(w, info)
			}
			writeIndent(w, ctx.Level)
			writeString(w, toYAMLString(key))
		}
		writeBytes(w, ':')

		if child == nil {
			child = root
		}
		switch actual := val.(type) {
		case map[string]interface{}:
			if len(actual) == 0 {
				writeBytes(w, ' ')
			} else {
				writeBytes(w, '\n')
			}
			m.marshalMapWithCtx(Context{
				NewLine: true,
				Level:   ctx.Level + 1,
				Node:    child,
			}, actual, w)
			continue
		case []interface{}:
			if len(actual) == 0 {
				writeBytes(w, ' ')
			} else {
				writeBytes(w, '\n')
			}
			m.marshalListWithCtx(ctx.WithNewLine(true).WithNode(child), actual, w)
			continue
		}

		writeBytes(w, ' ')
		switch actual := val.(type) {
		case string:
			writeString(w, toYAMLStringValueln(m.emptyInfo, actual, ctx.Level+1))
		case nil:
			w.Write(bytesNull)
		default:
			writeString(w, formatBasicType(val)+"\n")
		}
	}
	return nil
}

func (m *Marshaller) marshalListWithCtx(ctx Context, o []interface{}, w io.Writer) error {
	if o == nil {
		w.Write(bytesNull)
		return nil
	}
	if len(o) == 0 {
		w.Write(bytesEmptyList)
		return nil
	}

	root := ctx.Node
	prefix := getInfoOr(root, m.emptyInfo, m.colorizer)
	for i, val := range o {
		switch actual := val.(type) {
		case map[string]interface{}:
			child := root // fallback to root
			for _, s := range root.Keys {
				if fieldListMatchObject(s.Key, actual) {
					child = s.Node
					break
				}
			}
			mapPrefix := getInfoOr(child, m.emptyInfo, m.colorizer)
			if len(actual) > 0 {
				firstKey := firstSortedMapKey(actual)
				if fc := child.Fields[firstKey]; fc != nil {
					mapPrefix = getInfoOr(fc, m.emptyInfo, m.colorizer)
				}
			}
			writeString(w, mapPrefix)
			writeIndent(w, ctx.Level)
			writeBytes(w, '-', ' ')
			m.marshalMapWithCtx(Context{
				Level:   ctx.Level + 1,
				NewLine: false,
				Node:    child,
			}, actual, w)
			continue
		case []interface{}:
			if i != 0 || ctx.NewLine {
				writeString(w, prefix)
				writeIndent(w, ctx.Level)
				writeBytes(w, '-', ' ')
			}
			m.marshalListWithCtx(ctx.WithNewLine(false).WithLevel(ctx.Level+1), actual, w)
			continue
		}

		valPrefix := prefix
		if root.Values != nil {
			s := value.ToString(value.NewValueInterface(val))
			if v, ok := root.Values[s]; ok {
				valPrefix = getInfoOr(v.Node, m.emptyInfo, m.colorizer)
			}
		}
		writeString(w, valPrefix)
		writeIndent(w, ctx.Level)
		writeBytes(w, '-', ' ')
		switch actual := val.(type) {
		case string:
			writeString(w, toYAMLStringValueln(m.emptyInfo, actual, ctx.Level+1))
		case nil:
			w.Write(bytesNull)
		default:
			writeString(w, formatBasicType(val)+"\n")
		}
	}
	return nil
}
