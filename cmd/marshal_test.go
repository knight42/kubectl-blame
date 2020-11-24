package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

type M = map[string]interface{}
type L = []interface{}

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
	now := metav1.NewTime(time.Unix(1606150365, 0).UTC())
	s1 := fieldpath.NewSet(
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
	)

	s2 := fieldpath.NewSet(
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

	f1, _ := s1.ToJSON()
	f2, _ := s2.ToJSON()
	f3, _ := s3.ToJSON()

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
			},
			Labels: map[string]string{
				"app":     "bar",
				"version": "v1",
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
m1 (Update 2020-11-23 16:52:45 +0000)   labels:
m1 (Update 2020-11-23 16:52:45 +0000)     app: bar
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
                                      status: {}
`
	data, err := MarshalMetaObject(pod)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, string(data))
}
