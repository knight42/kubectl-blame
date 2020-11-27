package cmd

import "testing"

func TestYAMLMarshal(t *testing.T) {
	o := M{
		"list-in-map": L{
			"nil-list",
			(L)(nil),
			"nil-map",
			(M)(nil),
			"empty-list",
			L{},
			"empty-map",
			M{},
			"map-in-list",
			M{
				"fl-oat":             2.71828,
				"fl/oat":             3.14159,
				"example.com/fl-oat": float64(1) / float64(3),
			},
		},
		"bool": L{
			"yes", "Yes", "YES", "yES",
			"No", "NO", "no", "nO",
			"tRue", "TRUE", "true", "True",
			"false", "False", "FALSE", "fALsE",
			true,
			false,
		},
		"empty-list": L{},
		"empty-map":  M{},
		"nil-list":   (L)(nil),
		"nil-map":    (M)(nil),
		"nested-map": M{
			"yes":    true,
			"yEs":    true,
			"no":     false,
			"nO":     false,
			"true":   true,
			"truE":   true,
			"false":  false,
			"fALse":  false,
			"time":   "2020-11-06T06:29:56Z",
			"string": "Running",

			"multiple-line-string":                     "a\nlong\nstring",
			"multiple-line-string-with-newline":        "a\nlong\nstring\n",
			"multiple-line-string-with-invisible-char": "a\nlon\tg\nstring\n",
			"m1":                            M{"m2": M{"m3": "a\nlong\nstring"}},
			"string-with-newline":           "xzc\n",
			"string-with-multiple-newlines": "xzc\n\n\n",

			"single-quote":        `single-'quote'`,
			"escape-single-quote": `'single\'-'quote'`,
			"double-quote":        `"double"-quote`,
			"escape-double-quote": `d"ouble-\"double\""-quote`,

			"map":  M{"k1": 1, "k2": "v2"},
			"list": L{"0", 1, 2.0},
		},
	}
	_ = o
	//
	// expected, err := yaml.Marshal(o)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//
	// m := Marshaller{}
	// got, err := m.MarshalUnstructuredObject(o)
	// if err != nil {
	// 	t.Fatalf("%+v\n", err)
	// }
	// assert.Equal(t, string(expected), string(got))
}
