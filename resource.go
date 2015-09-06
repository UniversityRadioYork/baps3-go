package bifrost

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

// TODO(CaptainHayashi): Do we need all this machinery?
// []Resources are practically only emitted by converting ResourceNodes!

// Resourcifier is the interface for things that can be converted to resource
// lists.
type Resourcifier interface {
	// Resourcify converts a Resourcifier into a list of resources.
	Resourcify(path []string) []Resource
}

// ToResource converts an item and its location in the tree to a list of resources.
// If the item is a Resourcifier, Resourcify() is called on it.
// Struct fields may be annotated with a `res` tag giving the name the
// corresponding child should take in the resource.
func ToResource(path []string, item interface{}) []Resource {
	// First, see if item can do the work for us.
	switch item := item.(type) {
	case BifrostType:
		return []Resource{{path, item}}
	case Resourcifier:
		return item.Resourcify(path)
	default:
		return toResourceReflect(path, reflect.ValueOf(item), reflect.TypeOf(item))
	}
}

func toResourceReflect(path []string, val reflect.Value, typ reflect.Type) []Resource {
	switch val.Kind() {
	case reflect.Ptr:
		// Don't call toResourceReflect here; otherwise, we'll forget
		// to check to see if it's a Resourcifier.
		return ToResource(path, reflect.Indirect(val).Interface())
	case reflect.Map:
		return mapToResource(path, val, typ)
	case reflect.Struct:
		return structToResource(path, val, typ)
	case reflect.Array, reflect.Slice:
		return sliceToResource(path, val, typ)
	default:
		return []Resource{{path: path, value: ToBifrostType(val.Interface())}}
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
		if fieldt.PkgPath != "" || fieldt.Anonymous {
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
	res := []Resource{{path, BifrostTypeDirectory{numChildren: len}}}

	for i := 0; i < len; i++ {
		fieldv := val.Index(i)
		res = append(res, ToResource(append(path, strconv.Itoa(i)), fieldv.Interface())...)
	}

	return res
}

func mapToResource(path []string, val reflect.Value, typ reflect.Type) []Resource {
	// This is similar to sliceToResource, but now we're indexing over keys
	// too.
	len := val.Len()

	res := []Resource{{path, BifrostTypeDirectory{numChildren: len}}}

	for _, kval := range val.MapKeys() {
		// We just stringify keys to turn them into bits of path.
		// This should be sufficient for 99.9% of conditions.
		kstr := fmt.Sprint(kval.Interface())

		res = append(res, ToResource(append(path, kstr), val.MapIndex(kval).Interface())...)
	}

	return res
}
