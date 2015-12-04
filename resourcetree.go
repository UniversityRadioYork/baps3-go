package bifrost

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// splitPath splits a resource path separated by /'s into segments
// (ignoring leading/trailing /'s and empty segments)
func splitPath(path string) []string {
	f := func(c rune) bool {
		return c == '/'
	}
	return strings.FieldsFunc(path, f)
}

type Response struct {
	Status Status
	Path   []string
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
		splitPath,
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
		splitPath,
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

func (n DirectoryResourceNode) Resourcify(path []string) []Resource {
	return ToResource(path, map[string]ResourceNoder(n))
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

func (n EntryResourceNode) Resourcify(path []string) []Resource {
	return ToResource(path, n.Value)
}

// Nodifier is the interface for things that can be converted to resource
// nodes.
type Nodifier interface {
	// Nodify converts this value into a ResourceNoder.
	Nodify() ResourceNoder
}

// ToNode converts an arbitrary value into a resource node.
// If the item is a ResourceNoder, it is ignored.
// If the item is a Nodifier, Nodify() is called on it.
// Struct fields may be annotated with a `res` tag giving the name the
// corresponding child should take in the resource.
func ToNode(item interface{}) ResourceNoder {
	// First, see if item can do the work for us.
	switch item := item.(type) {
	case ResourceNoder:
		return item
	case Nodifier:
		return item.Nodify()
	default:
		return toNodeReflect(reflect.ValueOf(item), reflect.TypeOf(item))
	}
}

func toNodeReflect(val reflect.Value, typ reflect.Type) ResourceNoder {
	switch val.Kind() {
	case reflect.Ptr:
		// Don't call toNodeReflect here; otherwise, we'll forget
		// to check to see if it's a Node or Nodifier.
		return ToNode(reflect.Indirect(val).Interface())
	case reflect.Map:
		return mapToNode(val, typ)
	case reflect.Struct:
		return structToNode(val, typ)
	case reflect.Array, reflect.Slice:
		return sliceToNode(val, typ)
	default:
		return NewEntryResourceNode(ToBifrostType(val.Interface()))
	}
}

func structToNode(val reflect.Value, typ reflect.Type) ResourceNoder {
	nf := val.NumField()

	children := map[string]ResourceNoder{}

	// Now, work out the fields.
	for i := 0; i < nf; i++ {
		fieldt := typ.Field(i)

		// We can't announce fields that aren't exported.
		// If this one isn't, knock one off the available fields and ignore it.
		if fieldt.PkgPath != "" || fieldt.Anonymous {
			continue
		}

		// Work out the resource name from the field name/tag.
		tag := fieldt.Tag.Get("res")
		if tag == "" {
			tag = fieldt.Name
		}

		// Now, recursively emit and collate each resource.
		fieldv := val.Field(i)
		children[tag] = ToNode(fieldv.Interface())
	}

	// Now package the children into a DirectoryResourceNode.
	return NewDirectoryResourceNode(children)
}

func sliceToNode(val reflect.Value, typ reflect.Type) ResourceNoder {
	len := val.Len()

	// As before, but now with a list and indexes.
	children := map[string]ResourceNoder{}

	for i := 0; i < len; i++ {
		fieldv := val.Index(i)
		children[strconv.Itoa(i)] = ToNode(fieldv.Interface())
	}

	// Now package the children into a DirectoryResourceNode.
	return NewDirectoryResourceNode(children)
}

func mapToNode(val reflect.Value, typ reflect.Type) ResourceNoder {
	// This is similar to sliceToResource, but now we're indexing over keys
	// too.
	children := map[string]ResourceNoder{}

	for _, kval := range val.MapKeys() {
		// We just stringify keys to turn them into bits of path.
		// This should be sufficient for 99.9% of conditions.
		kstr := fmt.Sprint(kval.Interface())

		children[kstr] = ToNode(val.MapIndex(kval).Interface())
	}

	// Now package the children into a DirectoryResourceNode.
	return NewDirectoryResourceNode(children)
}
