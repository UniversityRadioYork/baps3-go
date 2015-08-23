package bifrost

import (
	"fmt"
	"reflect"
	"strconv"
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

// Message flattens a Resource into a Bifrost RES, given the tag of the read
// generating it.
//
// TODO(CaptainHayashi): does this belong elsewhere?
func (r *Resource) Message(tag string) *Message {
	vtype, val := r.value.ResourceBody()
	return NewMessage(RsRes).AddArg(tag).AddArg("/" + strings.Join(r.path, "/")).AddArg(vtype).AddArg(val)
}

// ToResource converts an item and its location in the tree to a list of resources.
// Struct fields may be annotated with a `res` tag giving the name the
// corresponding child should take in the resource.
func ToResource(path []string, item interface{}) []Resource {
	val := reflect.ValueOf(item)
	typ := reflect.TypeOf(item)

	switch val.Kind() {
	case reflect.Struct:
		return structToResource(path, val, typ)
	case reflect.Array, reflect.Slice:
		return sliceToResource(path, val, typ)
	case reflect.Int:
		// TODO(CaptainHayashi): catch more integers here?
		return []Resource{{path: path, value: BifrostTypeInt(item.(int))}}
	default:
		// TODO(CaptainHayashi): enums?
		return []Resource{{path: path, value: BifrostTypeString(fmt.Sprint(item))}}
	}
}

func structToResource(path []string, val reflect.Value, typ reflect.Type) []Resource {
	nf := val.NumField()
	af := nf

	// First, reserve space for the incoming directory.
	// We'll fix the inner value later.
	res := []Resource{{path: path, value: nil}}

	// Now, recursively work out the fields.
	for i := 0; i < nf; i++ {
		fieldt := typ.Field(i)

		// We can't announce fields that aren't exported.
		// If this one isn't, knock one off the available fields and ignore it.
		if fieldt.PkgPath != "" {
			af--
			continue
		}

		// Work out the resource name from the field name/tag.
		tag := fieldt.Tag.Get("res")
		if tag == "" {
			tag = fieldt.Name
		}

		// Now, recursively emit and collate each resource.
		fieldv := val.Field(i)
		res = append(res, ToResource(append(path, tag), fieldv.Interface())...)
	}

	// Now fill in the final available fields count
	res[0].value = BifrostTypeDirectory{numChildren: af}

	return res
}

func sliceToResource(path []string, val reflect.Value, typ reflect.Type) []Resource {
	len := val.Len()

	// As before, but now with a list and indexes.
	// TODO(CaptainHayashi): modelling a list as a directory
	res := []Resource{{path, BifrostTypeDirectory{numChildren: len}}}

	for i := 0; i < len; i++ {
		fieldv := val.Index(i)
		res = append(res, ToResource(append(path, strconv.Itoa(i)), fieldv.Interface())...)
	}

	return res
}

type Response struct {
	Status    Status
	Resources []Resource
}

type ResourceNoder interface {
	NRead(prefix, relpath []string) ([]Resource, error)
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
	resps, err := r.NRead([]string{}, splitPath)
	status := Status{StatusOk, ""}
	if err != nil {
		status = Status{StatusError, err.Error()}
	}
	return Response{
		status,
		resps,
	}
}

func (r ResourceNode) NRead(_, _ []string) ([]Resource, error) {
	return nil, fmt.Errorf("THIS SHOULDNT HAPPEN")
}

type DirectoryResourceNode struct {
	ResourceNode // We'll include this for completeness :-)
	children     map[string]ResourceNoder
}

func NewDirectoryResourceNode(children map[string]ResourceNoder) *DirectoryResourceNode {
	return &DirectoryResourceNode{
		ResourceNode{},
		children,
	}
}

func (n *DirectoryResourceNode) NAdd(prefix, relpath []string, v ResourceNoder) error {
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
			return node.NAdd(newPrefix, relpath[1:], v)
		} else { // Nothing here, add a new directory
			newNode := NewDirectoryResourceNode(make(map[string]ResourceNoder))
			n.children[relpath[0]] = newNode
			return newNode.NAdd(newPrefix, relpath[1:], v)
		}
	}
}

func (n *DirectoryResourceNode) NRead(prefix, relpath []string) ([]Resource, error) {
	if len(relpath) == 0 { // This is the resource being Read
		childResources := []Resource{}
		for childNodeName, childNode := range n.children {
			r, err := childNode.NRead(append(prefix, childNodeName), []string{})
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
			return node.NRead(newPrefix, relpath[1:])
		} else { // Nothing here, error time!
			// TODO(wlcx): error types
			return nil, fmt.Errorf("Path %s does not exist", strings.Join(newPrefix, "/"))
		}
	}
}

func (n *DirectoryResourceNode) NWrite(prefix, relpath []string, value BifrostType) error {
	return nil
}

func (n *DirectoryResourceNode) NDelete(prefix, relpath []string) error {
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

func (n *EntryResourceNode) NAdd(prefix, relpath []string, v ResourceNoder) error {
	// Trying to add something but we've hit a leaf node - stop. Error time.
	return fmt.Errorf("Path %s already exists", strings.Join(append(prefix, relpath[0]), "/"))
}

func (n *EntryResourceNode) NRead(prefix, relpath []string) ([]Resource, error) {
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

func (n *EntryResourceNode) NWrite(prefix, relpath []string, value BifrostType) error {
	return nil
}

func (n *EntryResourceNode) NDelete(prefix, relpath []string) error {
	return nil
}
