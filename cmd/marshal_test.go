package cmd

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

type M = map[string]interface{}

func TestFieldListMatchObject(t *testing.T) {
	testCases := []struct {
		name        string
		fieldList   value.FieldList
		object      M
		shouldMatch bool
	}{
		{
			name: "match",
			fieldList: value.FieldList{
				value.Field{Name: "containerPort", Value: value.NewValueInterface(80)},
				value.Field{Name: "protocol", Value: value.NewValueInterface("TCP")},
			},
			object: M{
				"containerPort": 80,
				"protocol":      "TCP",
				"name":          "foo",
			},
			shouldMatch: true,
		},
		{
			name: "mismatch",
			fieldList: value.FieldList{
				value.Field{Name: "containerPort", Value: value.NewValueInterface(80)},
				value.Field{Name: "protocol", Value: value.NewValueInterface("TCP")},
			},
			object: M{
				"containerPort": 8080,
				"protocol":      "TCP",
				"name":          "foo",
			},
			shouldMatch: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := fieldListMatchObject(tc.fieldList, tc.object)
			if got != tc.shouldMatch {
				t.Errorf("Unexpected result: got=%v, expected=%v", got, tc.shouldMatch)
			}
		})
	}
}

func TestMarshaller_MarshalMetaObject(t *testing.T) {
	r := require.New(t)
	now := metav1.NewTime(time.Unix(1606150365, 0).UTC())
	s1 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("metadata", "finalizers"),
		fieldpath.MakePathOrDie("metadata", "finalizers",
			value.NewValueInterface("service.kubernetes.io/load-balancer-cleanup"),
		),
		fieldpath.MakePathOrDie("metadata", "labels"),
		fieldpath.MakePathOrDie("metadata", "labels", "app"),
		fieldpath.MakePathOrDie("metadata", "ownerReferences"),
		fieldpath.MakePathOrDie("metadata", "ownerReferences",
			fieldpath.KeyByFields("uid", "72594682-7b8d-4d52-bb84-8cab3cd2e16f")),
		fieldpath.MakePathOrDie("ownerReferences",
			fieldpath.KeyByFields("uid", "72594682-7b8d-4d52-bb84-8cab3cd2e16f"),
			"kind",
		),
		fieldpath.MakePathOrDie("ownerReferences",
			fieldpath.KeyByFields("uid", "72594682-7b8d-4d52-bb84-8cab3cd2e16f"),
			"apiVersion",
		),
		fieldpath.MakePathOrDie("spec", "serviceAccountName"),
	)

	s2 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("metadata", "finalizers",
			value.NewValueInterface("service.kubernetes.io/foo"),
		),
		fieldpath.MakePathOrDie("spec", "serviceAccountName"),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1")),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"), "image"),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"), "ports"),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "TCP")),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "TCP"),
			"containerPort"),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "TCP"),
			"protocol"),
	)

	s3 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "UDP")),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "UDP"),
			"containerPort"),
		fieldpath.MakePathOrDie("spec", "containers",
			fieldpath.KeyByFields("name", "c1"),
			"ports",
			fieldpath.KeyByFields("containerPort", 53, "protocol", "UDP"),
			"protocol"),
	)

	s4 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("metadata", "labels", "s4"),
	)

	f1, _ := s1.ToJSON()
	f2, _ := s2.ToJSON()
	f3, _ := s3.ToJSON()
	f4, _ := s4.ToJSON()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:   "m1",
					Operation: metav1.ManagedFieldsOperationUpdate,
					Time:      &now,
					FieldsV1:  &metav1.FieldsV1{Raw: f1},
				},
				{
					Manager:   "m2",
					Operation: metav1.ManagedFieldsOperationUpdate,
					Time:      &now,
					FieldsV1:  &metav1.FieldsV1{Raw: f2},
				},
				{
					Manager:   "m3",
					Operation: metav1.ManagedFieldsOperationUpdate,
					Time:      &now,
					FieldsV1:  &metav1.FieldsV1{Raw: f3},
				},
				{
					Manager:   "m4",
					Operation: metav1.ManagedFieldsOperationUpdate,
					FieldsV1:  &metav1.FieldsV1{Raw: f4},
				},
			},
			Labels: map[string]string{
				"app":     "bar",
				"version": "v1",
				"s4":      "v",
			},
			Finalizers: []string{
				"service.kubernetes.io/load-balancer-cleanup",
				"service.kubernetes.io/foo",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					UID:  "72594682-7b8d-4d52-bb84-8cab3cd2e16f",
					Kind: "ReplicaSet",
					Name: "bar-xxxx",
				},
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "foo",
			Containers: []corev1.Container{
				{
					Name:  "c1",
					Image: "image:latest",
					Ports: []corev1.ContainerPort{
						{
							Name:          "tcp",
							ContainerPort: 53,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "udp",
							ContainerPort: 53,
							Protocol:      corev1.ProtocolUDP,
						},
					},
				},
			},
		},
	}

	const expected = `                                      metadata:
                                        creationTimestamp: null
m1 (Update 2020-11-23 16:52:45 +0000)   finalizers:
m1 (Update 2020-11-23 16:52:45 +0000)   - service.kubernetes.io/load-balancer-cleanup
m2 (Update 2020-11-23 16:52:45 +0000)   - service.kubernetes.io/foo
m1 (Update 2020-11-23 16:52:45 +0000)   labels:
m1 (Update 2020-11-23 16:52:45 +0000)     app: bar
m4 (Update                          )     s4: v
m1 (Update 2020-11-23 16:52:45 +0000)     version: v1
m1 (Update 2020-11-23 16:52:45 +0000)   ownerReferences:
m1 (Update 2020-11-23 16:52:45 +0000)   - apiVersion: ""
m1 (Update 2020-11-23 16:52:45 +0000)     kind: ReplicaSet
m1 (Update 2020-11-23 16:52:45 +0000)     name: bar-xxxx
m1 (Update 2020-11-23 16:52:45 +0000)     uid: 72594682-7b8d-4d52-bb84-8cab3cd2e16f
                                      spec:
                                        containers:
m2 (Update 2020-11-23 16:52:45 +0000)   - image: image:latest
m2 (Update 2020-11-23 16:52:45 +0000)     name: c1
m2 (Update 2020-11-23 16:52:45 +0000)     ports:
m2 (Update 2020-11-23 16:52:45 +0000)     - containerPort: 53
m2 (Update 2020-11-23 16:52:45 +0000)       name: tcp
m2 (Update 2020-11-23 16:52:45 +0000)       protocol: TCP
m3 (Update 2020-11-23 16:52:45 +0000)     - containerPort: 53
m3 (Update 2020-11-23 16:52:45 +0000)       name: udp
m3 (Update 2020-11-23 16:52:45 +0000)       protocol: UDP
m2 (Update 2020-11-23 16:52:45 +0000)     resources: {}
m1 (Update 2020-11-23 16:52:45 +0000) /
m2 (Update 2020-11-23 16:52:45 +0000)   serviceAccountName: foo
                                      status: {}
`
	data, err := MarshalMetaObject(pod, TimeFormatFull, nil)
	r.NoError(err)
	if diff := cmp.Diff(expected, string(data)); len(diff) > 0 {
		t.Errorf("unexpected diff (-want +got): %s", diff)
	}
}

func TestBuildTree(t *testing.T) {
	s1 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("metadata", "finalizers"),
		fieldpath.MakePathOrDie("metadata", "finalizers",
			value.NewValueInterface("service.kubernetes.io/load-balancer-cleanup"),
		),
	)
	s2 := fieldpath.NewSet(
		fieldpath.MakePathOrDie("metadata", "finalizers",
			value.NewValueInterface("service.kubernetes.io/foo"),
		),
	)

	f1, _ := s1.ToJSON()
	f2, _ := s2.ToJSON()

	m := Marshaller{
		now:        time.Now(),
		timeFormat: TimeFormatNone,
	}
	r := require.New(t)
	root, err := m.buildTree([]metav1.ManagedFieldsEntry{
		{
			Manager:   "m1",
			Operation: metav1.ManagedFieldsOperationApply,
			FieldsV1: &metav1.FieldsV1{
				Raw: f1,
			},
		},
		{
			Manager:   "m2",
			Operation: metav1.ManagedFieldsOperationUpdate,
			FieldsV1: &metav1.FieldsV1{
				Raw: f2,
			},
		},
	}, 0, 0, 0)
	r.NoError(err)

	leaf1 := &Node{
		Managers: []ManagerInfo{
			{
				Manager:   "m1",
				Operation: string(metav1.ManagedFieldsOperationApply),
			},
		},
	}
	leaf2 := &Node{
		Managers: []ManagerInfo{
			{
				Manager:   "m2",
				Operation: string(metav1.ManagedFieldsOperationUpdate),
			},
		},
	}
	node1 := &Node{
		Values: map[string]*ValueWithNode{
			`"service.kubernetes.io/load-balancer-cleanup"`: {
				Value: value.NewValueInterface("service.kubernetes.io/load-balancer-cleanup"),
				Node:  leaf1,
			},
			`"service.kubernetes.io/foo"`: {
				Value: value.NewValueInterface("service.kubernetes.io/foo"),
				Node:  leaf2,
			},
		},
		Managers: []ManagerInfo{
			{
				Manager:   "m1",
				Operation: string(metav1.ManagedFieldsOperationApply),
			},
		},
	}
	leaf1.Parent, leaf2.Parent = node1, node1

	node2 := &Node{
		Fields: map[string]*Node{
			"finalizers": node1,
		},
	}
	node1.Parent = node2

	expected := &Node{
		Fields: map[string]*Node{
			"metadata": node2,
		},
	}
	node2.Parent = expected

	r.Equal(expected, root)
}
