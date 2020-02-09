package list

import (
	"strconv"

	"github.com/UniversityRadioYork/bifrost-go/core"
	"github.com/UniversityRadioYork/bifrost-go/message"
)

// RsCountL is the COUNTL response word.
const RsCountL = "COUNTL"

// CountLResponse announces the number of items in a list.
type CountLResponse int

// Message converts CountLResponse c to a message with tag tag.
func (c CountLResponse) Message(tag string) *message.Message {
	return message.New(tag, RsCountL).AddArgs(strconv.Itoa(int(c)))
}

// ParseCountLResponse tries to parse an arbitrary message as a COUNTL response.
func ParseCountLResponse(m *message.Message) (CountLResponse, error) {
	var err error
	if err = core.CheckWord(RsCountL, m); err != nil {
		return 0, err
	}

	var cstr string
	if cstr, err = core.OneArg(m); err != nil {
		return 0, err
	}

	var cint int
	cint, err = strconv.Atoi(cstr)
	return CountLResponse(cint), err
}
