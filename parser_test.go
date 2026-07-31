package ltsvparser

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

var parseTests = []struct {
	Name   string
	Input  string
	Keys   []string
	Values []string
	Hits   []bool
}{
	{
		Name:   "Simple",
		Input:  "user:kazeburo\tage:43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age", "weight"},
		Values: []string{"kazeburo", "43", "55.9"},
		Hits:   []bool{true, true, true},
	},
	{
		Name:   "Simple order",
		Input:  "user:kazeburo\tage:43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "weight", "age"},
		Values: []string{"kazeburo", "55.9", "43"},
		Hits:   []bool{true, true, true},
	},
	{
		Name:   "Simple order 2",
		Input:  "user:kazeburo\tage:43\theight:163.1\tweight:55.9",
		Keys:   []string{"height", "user", "age"},
		Values: []string{"163.1", "kazeburo", "43"},
		Hits:   []bool{true, true, true},
	},
	{
		Name:   "Empty",
		Input:  "user:kazeburo\tage:\theight:-\tweight:55.9",
		Keys:   []string{"user", "age", "height"},
		Values: []string{"kazeburo", "", "-"},
		Hits:   []bool{true, true, true},
	},
	{
		Name:   "Not exists",
		Input:  "user:kazeburo\tage:43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age2", "height"},
		Values: []string{"kazeburo", "", "163.1"},
		Hits:   []bool{true, false, true},
	},
	{
		Name:   "Not exit : in middle",
		Input:  "user:kazeburo\tage\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age", "height"},
		Values: []string{"kazeburo", "", "163.1"},
		Hits:   []bool{true, true, true},
	},
	{
		Name:   "only one",
		Input:  "user:kazeburo",
		Keys:   []string{"user"},
		Values: []string{"kazeburo"},
		Hits:   []bool{true},
	},
	{
		Name:   "not exist : at last",
		Input:  "user:kazeburo\tage",
		Keys:   []string{"user", "age"},
		Values: []string{"kazeburo", ""},
		Hits:   []bool{true, true},
	},
	{
		Name:   "parse not ignore last",
		Input:  "user:kazeburo\tage:",
		Keys:   []string{"user", "age"},
		Values: []string{"kazeburo", ""},
		Hits:   []bool{true, true},
	},
	{
		Name:   "parse end with tab",
		Input:  "user:kazeburo\t",
		Keys:   []string{"user"},
		Values: []string{"kazeburo"},
		Hits:   []bool{true},
	},
	{
		Name:   "Simple Ir",
		Input:  "\tuser:kazeburo\t\tage::43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age", "weight", ""},
		Values: []string{"kazeburo", ":43", "55.9", ""},
		Hits:   []bool{true, true, true, false},
	},
	{
		Name:   "Simple Ir 2",
		Input:  "\tuser:kazeburo\t:\tage::43\theight:163.1\tweight:55.9",
		Keys:   []string{"user", "age", "weight", ""},
		Values: []string{"kazeburo", ":43", "55.9", ""},
		Hits:   []bool{true, true, true, true},
	},
	{
		Name:   "hyphen",
		Input:  "referer:-\tuser:kazeburo\t:\tage::43\theight:163.1\tweight:55.9",
		Keys:   []string{"referer", "user", "age", "weight", ""},
		Values: []string{"-", "kazeburo", ":43", "55.9", ""},
		Hits:   []bool{true, true, true, true, true},
	},
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
