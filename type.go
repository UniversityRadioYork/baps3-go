package bifrost

import (
	"strconv"
)

type BifrostType interface {
	String() string
}

type BifrostTypeString string

func (t BifrostTypeString) String() string {
	return "STRING " + string(t)
}

type BifrostTypeInt int

func (t BifrostTypeInt) String() string {
	return "INT " + strconv.Itoa(int(t))
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

type BifrostTypeDirectory struct {
	numChildren int
}

func (t BifrostTypeDirectory) String() string {
	return "DIRECTORY " + strconv.Itoa(t.numChildren)
}

func ToBifrostType(val interface{}) BifrostType {
	switch val.(type) {
	case string:
		return BifrostTypeString(val.(string))
	default:
		return nil
	}
}
