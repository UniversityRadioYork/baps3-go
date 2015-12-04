package message

import (
	"strings"
	"unicode"
)

const (
	// Message word constants.

	// - Requests

	// RqRead denotes a 'read' request message.
	RqRead = "read"

	// RqWrite denotes a 'write' request message.
	RqWrite = "write"

	// RqDelete denotes a 'delete' request message.
	RqDelete = "delete"

	// - Responses

	// RsRes denotes a message with the 'RES' response.
	RsRes = "RES"

	// RsUpdate denotes a message with the 'RES' response.
	RsUpdate = "UPDATE"

	// RsAck denotes a message with the 'ACK' response.
	RsAck = "ACK"

	// RsOhai denotes a message with the 'OHAI' response.
	RsOhai = "OHAI"

	// AckOk denotes an ACK message with the 'OK' type.
	AckOk = "OK"

	// AckWhat denotes an ACK message with the 'WHAT' type.
	AckWhat = "WHAT"

	// AckFail denotes an ACK message with the 'FAIL' type.
	AckFail = "FAIL"
)

type Message []string

// Read constructs a 'read' request command, with tag and path to be read.
func Read(tag, path string) Message {
	return Message{RqRead, tag, path}
}

// Write constructs a 'write' request command, with tag, path to be written
// to and value to write.
func Write(tag, path, value string) Message {
	return Message{RqWrite, tag, path, value}
}

// Delete constructs a 'delete' request command, with tag and path to be deleted.
func Delete(tag, path string) Message {
	return Message{RqDelete, tag, path}
}

// Res constructs a 'RES' response command, with tag, path, type of value and
// actual value of said path.
func Res(tag, path, val_type, value string) Message {
	return Message{RsRes, tag, path, val_type, value}
}

// Update constructs an 'UPDATE' response command, with path that's been
// updated and the path's new value with its type.
func Update(path, val_type, value string) Message {
	return Message{RsUpdate, path, val_type, value}
}

// Ack constructs an 'ACK' response command, with the type of ACK and message,
// followed by the original request command.
func Ack(ack_type, msg string, orig_cmd Message) Message {
	resp := Message{RsAck, ack_type, msg}
	return append(resp, orig_cmd...)
}

func escapeArgument(input string) string {
	return "'" + strings.Replace(input, "'", `'\''`, -1) + "'"
}

// Pack outputs the given Message as raw bytes representing a BAPS3 message.
// These bytes can be sent down a TCP connection to a BAPS3 server, providing
// they are terminated using a line-feed character.
func (m Message) Pack() []byte {
	outstr := m[0]
	for _, a := range m[1:] {
		// Escape arg if needed
		for _, c := range a {
			if c < unicode.MaxASCII && (unicode.IsSpace(c) || strings.ContainsRune(`'"\`, c)) {
				a = escapeArgument(a)
				break
			}
		}
		outstr += " " + a
	}
	outstr += "\n"
	return []byte(outstr)
}

// Converts a message into a string representation. Note that it doesn't escape
// the arguments, so is likely only useful for logging and debugging.
func (m Message) String() string {
	outstr := m[0]
	for _, s := range m[1:] {
		outstr += " " + s
	}
	return outstr
}
