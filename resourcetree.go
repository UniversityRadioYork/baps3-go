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

type Resource struct {
	path  []string
	value BifrostType
}

func (r *Resource) String() string {
	return fmt.Sprintf("/%s %s", strings.Join(r.path, "/"), r.value.String())
}

type Response struct {
	Status    Status
	Resources []Resource
}

type ResourceNoder interface {
	read(prefix, relpath []string) ([]Resource, error)
	write(prefix, relpath []string, value BifrostType) error
	delete(prefix, relpath []string) error
	add(prefix, relpath []string, v ResourceNoder) error
}

type ResourceNode struct {
}

func Add(r ResourceNoder, path string, n ResourceNoder) error {
	splitPath := splitPath(path)
	return r.add([]string{}, splitPath, n)
}

func Read(r ResourceNoder, path string) Response {
	splitPath := splitPath(path)
	resps, err := r.read([]string{}, splitPath)
	status := Status{StatusOk, ""}
	if err != nil {
		status = Status{StatusError, err.Error()}
	}
	return Response{
		status,
		resps,
	}
}

func (r ResourceNode) read(_, _ []string) ([]Resource, error) {
	return nil, fmt.Errorf("THIS SHOULDNT HAPPEN")
}

type DirectoryResourceNode struct {
	ResourceNode // We'll include this for completeness :-)
	children     map[string]ResourceNoder
}

func NewDirectoryResourceNode() DirectoryResourceNode {
	return DirectoryResourceNode{
		ResourceNode{},
		make(map[string]ResourceNoder),
	}
}

func (n DirectoryResourceNode) add(prefix, relpath []string, v ResourceNoder) error {
	switch len(relpath) { // Error, trying to add a node to
	case 0:
		// Something
		return fmt.Errorf("IDKLOL")
	case 1:
		// We are adding the node under this one.
		newPrefix := append(prefix, relpath[0])
		if _, exists := n.children[relpath[0]]; exists {
			// A node already exists under that name.
			return fmt.Errorf("Path %s already exists", strings.Join(newPrefix, "/"))
		} else {
			n.children[relpath[0]] = v // Add the child
			return nil
		}
	default: // Traverse!
		newPrefix := append(prefix, relpath[0])
		if node, ok := n.children[relpath[0]]; ok {
			return node.add(newPrefix, relpath[1:], v)
		} else { // Nothing here, add a new directory
			newNode := NewDirectoryResourceNode()
			n.children[relpath[0]] = newNode
			return newNode.add(newPrefix, relpath[1:], v)
		}
	}
}

func (n DirectoryResourceNode) read(prefix, relpath []string) ([]Resource, error) {
	if len(relpath) == 0 { // This is the resource being Read
		childResources := []Resource{}
		for childNodeName, childNode := range n.children {
			r, err := childNode.read(append(prefix, childNodeName), []string{})
			if err != nil {
				return nil, err
			}
			for _, res := range r {
				childResources = append(childResources, res)
			}
		}
		return append(
			childResources,
			Resource{
				prefix,
				BifrostTypeDirectory{
					len(childResources),
				},
			},
		), nil
	} else {
		newPrefix := append(prefix, relpath[0])
		if node, ok := n.children[relpath[0]]; ok {
			return node.read(newPrefix, relpath[1:])
		} else { // Nothing here, error time!
			// TODO(wlcx): error types
			return nil, fmt.Errorf("Path %s does not exist", strings.Join(newPrefix, "/"))
		}
	}
}

func (n DirectoryResourceNode) write(prefix, relpath []string, value BifrostType) error {
	return nil
}

func (n DirectoryResourceNode) delete(prefix, relpath []string) error {
	return nil
}

type EntryResourceNode struct {
	ResourceNode
	Value BifrostType
}

func NewEntryResourceNode(v BifrostType) EntryResourceNode {
	return EntryResourceNode{
		ResourceNode{},
		v,
	}
}

func (n EntryResourceNode) add(prefix, relpath []string, v ResourceNoder) error {
	// Trying to add something but we've hit a leaf node - stop. Error time.
	return fmt.Errorf("Path %s already exists", strings.Join(append(prefix, relpath[0]), "/"))
}

func (n EntryResourceNode) read(prefix, relpath []string) ([]Resource, error) {
	if len(relpath) != 0 { // Bad request, this is not a directory
		return nil, fmt.Errorf("Path %s does not exist", prefix)
	} else {
		return []Resource{
			Resource{
				prefix,
				n.Value,
			},
		}, nil
	}
}

func (n EntryResourceNode) write(prefix, relpath []string, value BifrostType) error {
	return nil
}

func (n EntryResourceNode) delete(prefix, relpath []string) error {
	return nil
}
