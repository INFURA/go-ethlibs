package jsonrpc

import (
	"github.com/INFURA/go-ethlibs/eth"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeParams(t *testing.T) {

	type testCase struct {
		Description string
		Expected    Params
		Input       []interface{}
	}

	type Object struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}

	testCases := []testCase{
		{
			Description: "single string",
			Expected:    MustParams("newHeads"),
			Input:       []interface{}{"newHeads"},
		},
		{
			Description: "null",
			Expected:    nil,
			Input:       nil,
		},
		{
			Description: "string, bool",
			Expected:    MustParams("latest", true),
			Input:       []interface{}{"latest", true},
		},
		{
			Description: "complex object",
			// Don't use a map[string]interface here, make a new Object so we know
			// that it;s not just comparing the same object to itself
			Expected: MustParams(&Object{Foo: "foo", Bar: 42}),
			Input:    []interface{}{&Object{Foo: "foo", Bar: 42}},
		},
	}

	for _, testCase := range testCases {
		actual, err := MakeParams(testCase.Input...)
		assert.NoError(t, err, "should not fail")
		assert.Equal(t, testCase.Expected, actual, "%#v", testCase)
	}
}

func TestParams_DecodeInto(t *testing.T) {

	type testCase struct {
		Description string
		Expected    []interface{}
		Input       Params
		Test        func(tc *testCase) ([]interface{}, error)
	}

	type Object struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}

	testCases := []testCase{
		//{
		//	Description: "single string",
		//	Expected:    []interface{}{"foo"},
		//	Input:       MustParams("foo"),
		//	Test: func(tc *testCase) ([]interface{}, error) {
		//		var str string
		//		err := tc.Input.UnmarshalInto(&str)
		//		return []interface{}{str}, err
		//	},
		//},
		{
			Description: "string and bool",
			Expected:    []interface{}{"foo", true},
			Input:       MustParams("foo", true),
			Test: func(tc *testCase) ([]interface{}, error) {
				var str string
				var b bool
				err := tc.Input.UnmarshalInto(&str, &b)
				return []interface{}{str, b}, err
			},
		},
		{
			Description: "complex object",
			Expected:    []interface{}{&Object{Foo: "foo", Bar: 42}},
			Input:       MustParams(&Object{Foo: "foo", Bar: 42}),
			Test: func(tc *testCase) ([]interface{}, error) {
				var obj Object
				err := tc.Input.UnmarshalInto(&obj)
				return []interface{}{&obj}, err
			},
		},
		{
			Description: "decode a subset of params",
			Expected:    []interface{}{"latest"},
			Input:       MustParams("latest", true),
			Test: func(tc *testCase) ([]interface{}, error) {
				var str string
				err := tc.Input.UnmarshalInto(&str)
				return []interface{}{str}, err
			},
		},
		{
			Description: "receiver's type is a struct",
			Expected:    []interface{}{eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456"), BlockHash: (*eth.Data32)(nil), Address: []eth.Address(nil), Topics: [][]eth.Data32(nil)}},
			Input:       MustParams(&eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456")}),
			Test: func(tc *testCase) ([]interface{}, error) {
				var rec eth.LogFilter
				err := tc.Input.UnmarshalInto(&rec)
				return []interface{}{rec}, err
			},
		},
		{
			Description: "multiple element in params",
			Expected:    []interface{}{eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456")}, eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x5678"), ToBlock: eth.MustBlockNumberOrTag("0x1234")}},
			Input:       MustParams(&eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456")}, &eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x5678"), ToBlock: eth.MustBlockNumberOrTag("0x1234")}),
			Test: func(tc *testCase) ([]interface{}, error) {
				var rec1 eth.LogFilter
				var rec2 eth.LogFilter
				err := tc.Input.UnmarshalInto(&rec1, &rec2)
				return []interface{}{rec1, rec2}, err
			},
		},
	}

	for _, testCase := range testCases {
		actual, err := testCase.Test(&testCase)
		assert.NoError(t, err, "should not fail")
		assert.Equal(t, testCase.Expected, actual, "%#v", testCase)
	}

	// Lets do a decode single test case here too
	multiple := MustParams("str", 42, true)
	expected := []interface{}{"str", 42, true}
	var str string
	var num int
	var b bool

	assert.NoError(t, multiple.UnmarshalSingleParam(0, &str), "should not fail")
	assert.NoError(t, multiple.UnmarshalSingleParam(1, &num), "should not fail")
	assert.NoError(t, multiple.UnmarshalSingleParam(2, &b), "should not fail")

	assert.Equal(t, expected[0], str)
	assert.Equal(t, expected[1], num)
	assert.Equal(t, expected[2], b)

	// this should fail, not enough params
	object := Object{}
	assert.Error(t, multiple.UnmarshalSingleParam(3, &object), "should have failed")
}

func TestParams_DecodeInto_Fail(t *testing.T) {

	type testCase struct {
		Description string
		Expected    []interface{}
		Input       Params
		Test        func(tc *testCase) ([]interface{}, error)
	}

	testCases := []testCase{
		{
			Description: "params null",
			Expected:    nil,
			Input:       nil,
			Test: func(tc *testCase) ([]interface{}, error) {
				var str string
				err := tc.Input.UnmarshalInto(str)
				return nil, err
			},
		},
		{
			Description: "len(p)<len(rec)",
			Expected:    []interface{}{},
			Input:       MustParams("foo"),
			Test: func(tc *testCase) ([]interface{}, error) {
				var str string
				var b bool
				err := tc.Input.UnmarshalInto(&str, &b)
				return []interface{}{}, err
			},
		},
		//{
		//	Description: "",
		//	Expected:    []interface{}{eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456")}, eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x5678"), ToBlock: eth.MustBlockNumberOrTag("0x1234")}},
		//	Input:       MustParams(&eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("3456789"), ToBlock: eth.MustBlockNumberOrTag("0x3456")}, &eth.LogFilter{FromBlock: eth.MustBlockNumberOrTag("0x5678"), ToBlock: eth.MustBlockNumberOrTag("0x1234")}),
		//	Test: func(tc *testCase) ([]interface{}, error) {
		//		var rec eth.LogFilter
		//		err := tc.Input.UnmarshalInto(&rec)
		//		return []interface{}{}, err
		//	},
		//},

	}

	for _, testCase := range testCases {
		actual, err := testCase.Test(&testCase)
		if testCase.Input != nil {
			assert.Error(t, err, "should fail")
		}
		assert.Equal(t, testCase.Expected, actual, "%#v", testCase)
	}

}

//func TestParams_parsePositionalArguments_Fail(t *testing.T) {
//
//	type testCase struct {
//		Description string
//		args        []reflect.Value
//		err         error
//		rawArgs     json.RawMessage
//		types       []reflect.Type
//		//Test        func(tc *testCase) ([]interface{}, error)
//	}
//
//	testCases := []testCase{{
//		Description: "err == nil",
//		args:        nil,
//		err:         err,
//		rawArgs:     json.RawMessage{{"[asdfghj]"}},
//		types:       []reflect.Type{erty},
//	},
//	}
//
//	for _, testCase := range testCases {
//		actual, err := ParsePositionalArguments(testCase.rawArgs, testCase.types)
//		assert.NoError(t, err, "should not fail")
//		assert.Equal(t, testCase.args, actual, "%#v", testCase)
//	}
//
//}
