package bifrost

import (
	"fmt"
	"strings"
)

// splitPath splits a resource path separated by /'s into segments, ignoring leading/trailing /'s and
// empty segments
func splitPath(path string) []string {
	splitPath := []string{}
	var currentSegment string
	for _, c := range path {
		if c == '/' {
			if currentSegment != "" {
				splitPath = append(splitPath, currentSegment)
				currentSegment = ""
			}
		} else {
			currentSegment += string(c)
		}
	}
	if currentSegment != "" {
		splitPath = append(splitPath, currentSegment) // Whatever we have left
	}
	return splitPath
}

type Response struct {
	Status Status
	Node   ResourceNoder
}

type ResourceNoder interface {
	NRead(prefix, relpath []string) (ResourceNoder, error)
	NWrite(prefix, relpath []string, value BifrostType) error
	NDelete(prefix, relpath []string) error
	NAdd(prefix, relpath []string, v ResourceNoder) error
}

type ResourceNode struct {
}

func Add(r ResourceNoder, path string, n ResourceNoder) error {
	splitPath := splitPath(path)
	return r.NAdd([]string{}, splitPath, n)
}

func Read(r ResourceNoder, path string) Response {
	splitPath := splitPath(path)
	node, err := r.NRead([]string{}, splitPath)
	status := Status{StatusOk, ""}
	if err != nil {
		status = Status{StatusError, err.Error()}
	}
	return Response{
		status,
		node,
	}
}

func Write(r ResourceNoder, path, value string) Response {
	// TODO(CaptainHayashi): non-string arguments?
	// This will need to split the scalar stuff out of ToResponse so it works with BifrostTypes.

	s := BifrostTypeString(value)

	splitPath := splitPath(path)
	err := r.NWrite([]string{}, splitPath, s)
	status := Status{StatusOk, ""}
	if err != nil {
		status = Status{StatusError, err.Error()}
	}
	return Response{
		status,
		nil,
	}
}

func (r ResourceNode) NRead(_, _ []string) ([]Resource, error) {
	return nil, fmt.Errorf("THIS SHOULDNT HAPPEN")
}

type DirectoryResourceNode map[string]ResourceNoder

func NewDirectoryResourceNode(children map[string]ResourceNoder) DirectoryResourceNode {
	return DirectoryResourceNode(children)
}

func (n DirectoryResourceNode) NAdd(prefix, relpath []string, v ResourceNoder) error {
	switch len(relpath) {
	case 0: // Error, trying to add a node at this node's path TODO: maybe this shouldn't error?
		return fmt.Errorf("A node aready exists at that path")
	case 1:
		// We are adding the node under this one.
		newPrefix := append(prefix, relpath[0])
		if _, exists := n[relpath[0]]; exists {
			// A node already exists under that name.
			return fmt.Errorf("Path %s already exists", strings.Join(newPrefix, "/"))
		} else {
			n[relpath[0]] = v // Add the child
			return nil
		}
	default: // Traverse!
		newPrefix := append(prefix, relpath[0])
		if node, ok := n[relpath[0]]; ok {
			return node.NAdd(newPrefix, relpath[1:], v)
		} else { // Nothing here, add a new directory
			newNode := NewDirectoryResourceNode(make(map[string]ResourceNoder))
			n[relpath[0]] = newNode
			return newNode.NAdd(newPrefix, relpath[1:], v)
		}
	}
}

func (n DirectoryResourceNode) NRead(prefix, relpath []string) (ResourceNoder, error) {
	if len(relpath) == 0 { // This is the resource being Read
		return n, nil
	} else {
		newPrefix := append(prefix, relpath[0])
		if node, ok := n[relpath[0]]; ok {
			return node.NRead(newPrefix, relpath[1:])
		} else { // Nothing here, error time!
			// TODO(wlcx): error types
			return nil, fmt.Errorf("Path %s does not exist", strings.Join(newPrefix, "/"))
		}
	}
}

func (n DirectoryResourceNode) NWrite(prefix, relpath []string, value BifrostType) error {
	if len(relpath) == 0 { // This is the resource being Read
		return fmt.Errorf("can't read a directory")
	}
	newPrefix := append(prefix, relpath[0])
	if node, ok := n[relpath[0]]; ok {
		return node.NWrite(newPrefix, relpath[1:], value)
	}
	// Nothing here, error time!
	// TODO(wlcx): error types
	return fmt.Errorf("Path %s does not exist", strings.Join(newPrefix, "/"))
}

func (n DirectoryResourceNode) NDelete(prefix, relpath []string) error {
	return nil
}

type EntryResourceNode struct {
	ResourceNode
	Value BifrostType
}

func NewEntryResourceNode(v BifrostType) *EntryResourceNode {
	return &EntryResourceNode{
		ResourceNode{},
		v,
	}
}

func (n EntryResourceNode) NAdd(prefix, relpath []string, v ResourceNoder) error {
	// Trying to add something but we've hit a leaf node - stop. Error time.
	return fmt.Errorf("Path %s already exists", strings.Join(append(prefix, relpath[0]), "/"))
}

func (n EntryResourceNode) NRead(prefix, relpath []string) (ResourceNoder, error) {
	if len(relpath) != 0 { // Bad request, this is not a directory
		return nil, fmt.Errorf("Path %s does not exist", prefix)
	} else {
		return n, nil
	}
}

func (n EntryResourceNode) NWrite(prefix, relpath []string, value BifrostType) error {
	return nil
}

func (n EntryResourceNode) NDelete(prefix, relpath []string) error {
	return nil
}
