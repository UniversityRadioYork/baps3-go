package bifrost

import (
	"fmt"
	"strconv"
)

type BifrostType interface {
	String() string
	ResourceBody() (string, string)
}

type BifrostTypeString string

func (t BifrostTypeString) String() string {
	return "STRING " + string(t)
}
func (t BifrostTypeString) ResourceBody() (string, string) {
	return "string", string(t)
}

type BifrostTypeInt int

func (t BifrostTypeInt) String() string {
	return "INT " + strconv.Itoa(int(t))
}
func (t BifrostTypeInt) ResourceBody() (string, string) {
	return "int", strconv.Itoa(int(t))
}

// BifrostTypeEnum is a value in a set of possible values
// An example would be state - Playing, Stopped, Ejected
type BifrostTypeEnum struct {
	current   string
	available []string
}

func (t BifrostTypeEnum) String() string {
	return "I AM AN ENUM"
}
func (t BifrostTypeEnum) ResourceBody() (string, string) {
	// TODO(CaptainHayashi): correct?
	return "enum", t.current
}

type BifrostTypeDirectory struct {
	numChildren int
}

func (t BifrostTypeDirectory) String() string {
	return "DIRECTORY " + strconv.Itoa(t.numChildren)
}
func (t BifrostTypeDirectory) ResourceBody() (string, string) {
	return "directory", strconv.Itoa(t.numChildren)
}

// ToBifrostType converts a value to a BifrostType.
// It will return a BifrostTypeInt if the value is an integer, and
// BifrostTypeString otherwise (converting the value to a string).
func ToBifrostType(v interface{}) BifrostType {
	switch v := v.(type) {
	case int:
		return BifrostTypeInt(v)
	case string:
		return BifrostTypeString(v)
	default:
		return BifrostTypeString(fmt.Sprint(v))
	}
}
