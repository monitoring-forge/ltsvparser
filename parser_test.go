package ltsvparser

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

type testCase struct {
	Name   string
	Input  string
	Keys   []string
	Values []string
	Hits   []bool
}

type testCaseOptions func(*testCase)

func withName(name string) testCaseOptions {
	return func(tc *testCase) {
		tc.Name = name
	}
}

func withInput(input string) testCaseOptions {
	return func(tc *testCase) {
		tc.Input = input
	}
}

func withKeys(keys []string) testCaseOptions {
	return func(tc *testCase) {
		tc.Keys = keys
	}
}

func withValues(values []string) testCaseOptions {
	return func(tc *testCase) {
		tc.Values = values
	}
}

func withHits(hits []bool) testCaseOptions {
	return func(tc *testCase) {
		tc.Hits = hits
	}
}

func buildTestCase(opts ...testCaseOptions) testCase {
	tc := testCase{
		Name:   "Simple",
		Input:  "user:kazeburo\tage:43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age", "weight"},
		Values: []string{"kazeburo", "43", "55.9"},
		Hits:   []bool{true, true, true},
	}
	for _, opt := range opts {
		opt(&tc)
	}
	return tc
}

var parseTests = []testCase{
	buildTestCase(),
	buildTestCase(
		withName("Simple order"),
		withKeys([]string{"user", "weight", "age"}),
		withValues([]string{"kazeburo", "55.9", "43"})),
	buildTestCase(
		withName("Simple order 2"),
		withKeys([]string{"height", "user", "age"}),
		withValues([]string{"163.1", "kazeburo", "43"})),
	buildTestCase(
		withName("Empty"),
		withInput("user:kazeburo\tage:\theight:-\tweight:55.9"),
		withKeys([]string{"user", "age", "height"}),
		withValues([]string{"kazeburo", "", "-"})),
	buildTestCase(
		withName("Not exists"),
		withKeys([]string{"user", "age2", "height"}),
		withValues([]string{"kazeburo", "", "163.1"}),
		withHits([]bool{true, false, true})),
	buildTestCase(
		withName("Not exit : in middle"),
		withInput("user:kazeburo\tage\theight:163.1\tweight:55.9"),
		withKeys([]string{"user", "age", "height"}),
		withValues([]string{"kazeburo", "", "163.1"})),
	buildTestCase(
		withName("only one"),
		withInput("user:kazeburo"),
		withKeys([]string{"user"}),
		withValues([]string{"kazeburo"}),
		withHits([]bool{true})),
	buildTestCase(
		withName("not exist : at last"),
		withInput("user:kazeburo\tage"),
		withKeys([]string{"user", "age"}),
		withValues([]string{"kazeburo", ""}),
		withHits([]bool{true, true})),
	buildTestCase(
		withName("parse not ignore last"),
		withInput("user:kazeburo\tage:"),
		withKeys([]string{"user", "age"}),
		withValues([]string{"kazeburo", ""}),
		withHits([]bool{true, true})),
	buildTestCase(
		withName("parse end with tab"),
		withInput("user:kazeburo\t"),
		withKeys([]string{"user"}),
		withValues([]string{"kazeburo"}),
		withHits([]bool{true})),
	buildTestCase(
		withName("Simple Ir"),
		withInput("\tuser:kazeburo\t\tage::43\theight:163.1\tweight:55.9"),
		withKeys([]string{"user", "age", "weight", ""}),
		withValues([]string{"kazeburo", ":43", "55.9", ""}),
		withHits([]bool{true, true, true, false})),
	buildTestCase(
		withName("Simple Ir 2"),
		withInput("\tuser:kazeburo\t:\tage::43\theight:163.1\tweight:55.9"),
		withKeys([]string{"user", "age", "weight", ""}),
		withValues([]string{"kazeburo", ":43", "55.9", ""}),
		withHits([]bool{true, true, true, true})),
	buildTestCase(
		withName("hyphen"),
		withInput("referer:-\tuser:kazeburo\t:\tage::43\theight:163.1\tweight:55.9"),
		withKeys([]string{"referer", "user", "age", "weight"}),
		withValues([]string{"-", "kazeburo", ":43", "55.9"}),
		withHits([]bool{true, true, true, true})),
}

func TestEach(t *testing.T) {
	for _, pt := range parseTests {
		keys := make([][]byte, 0)
		for _, k := range pt.Keys {
			keys = append(keys, []byte(k))
		}
		values := make([]string, len(keys))
		hits := make([]bool, len(keys))
		err := Each([]byte(pt.Input), func(i int, v []byte) error {
			values[i] = string(v)
			hits[i] = true
			return nil
		}, keys...)
		if err != nil {
			t.Error(pt.Name, err)
		}
		if diff := deep.Equal(pt.Values, values); diff != nil {
			t.Error("values missmatch", pt.Name, diff)
		}
		if diff := deep.Equal(pt.Hits, hits); diff != nil {
			t.Error("hits missmatch", pt.Name, diff)
		}
	}
}

func TestEachError(t *testing.T) {
	count := 0
	err := Each(
		[]byte("user:kazeburo\tage:43\theight:163.1\tweight:55.9"),
		func(i int, v []byte) error {
			count = count + 1
			return fmt.Errorf("test")
		},
		[]byte("user"), []byte("age"),
	)
	if err == nil {
		t.Error("error should not be null")
	}
	if count != 1 {
		t.Errorf("cb called once %d", count)
	}
}

func TestEachCancel(t *testing.T) {
	count := 0
	err := Each(
		[]byte("user:kazeburo\tage:43\theight:163.1\tweight:55.9"),
		func(i int, v []byte) error {
			count = count + 1
			return Cancel
		},
		[]byte("user"),
		[]byte("age"),
	)
	if err != nil {
		t.Error("error should be null")
	}
	if count != 1 {
		t.Errorf("cb called once %d", count)
	}
}

func TestEachCancelWithNewCanceler(t *testing.T) {
	count := 0
	cancel := &Canceler{}
	err := Each(
		[]byte("user:kazeburo\tage:43\theight:163.1\tweight:55.9"),
		func(i int, v []byte) error {
			count = count + 1
			return cancel
		},
		[]byte("user"),
		[]byte("age"),
	)
	if err != nil {
		t.Error("error should be null")
	}
	if count != 1 {
		t.Errorf("cb called once %d", count)
	}
}
