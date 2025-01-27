package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type Node struct {
	Group    string
	Title    string
	Path     string
	Id       string
	Parent   *Node
	Children []*Node
}

var Nodes []*Node

func (n *Node) get_field_value_by_name(field_name string) interface{} {
	v := reflect.ValueOf(n).Elem()

	field := v.FieldByName(field_name)
	if !field.IsValid() {
		panic(fmt.Sprintf("field '%s' does not exist in node struct", field_name))
	}

	return field.Interface()
}

func (n *Node) get_group() string { return n.Group }
func (n *Node) set_group(group string) error {

	if valid_node_group(group) {
		n.Group = strings.ToLower(group)
		return nil
	}

	return fmt.Errorf("invalid group: %v", group)
}

func (n *Node) get_title() string            { return n.Title }
func (n *Node) set_title(title string) error { n.Title = title; return nil }

func (n *Node) get_path() string           { return n.Path }
func (n *Node) set_path(path string) error { n.Path = path; return nil }

func (n *Node) get_id() string { return n.Id }
func (n *Node) set_id(id string) {
	if id == "" {
		n.Id = uuid.NewString()
	} else {
		n.Id = id
	}
}

func (n *Node) get_depth() int { return CFG_GROUP_DEPTH[n.Group] }

func (n *Node) get_parent() *Node     { return n.Parent }
func (n *Node) get_children() []*Node { return n.Children }

func (n *Node) set_parent(parent *Node) error {

	if parent == nil {
		return nil
	}

	if n.get_depth() != parent.get_depth()+1 {
		return fmt.Errorf("depth mismatch between child and parent")
	}

	contained := false

	for _, child := range parent.Children {
		if child.get_id() == n.get_id() {
			contained = true
		}
	}

	if !contained {
		parent.Children = append(parent.Children, n)
	}

	n.Parent = parent

	return nil
}

func create_node(group, title string) (*Node, error) {

	root_dir, err := get_config_value(CFG_ROOT_FIELD)
	if err != nil {
		return nil, err
	}

	if !valid_node_group(group) {
		return nil, fmt.Errorf("invalid group: %v", group)
	}

	base, err := get_base_directory(root_dir, group)
	if err != nil {
		return nil, err
	}

	group_path, err := find_path("directory", base, group)
	if err != nil {
		return nil, err
	}

	var node_path string

	if CFG_GROUP_DEPTH[group] == len(CFG_GROUP_DEPTH)-1 {
		node_path = filepath.Join(group_path, title+CFG_NOTE_FILETYPE)
	} else {
		node_path = filepath.Join(group_path, title)
	}

	node, err := initialize_node(group, title, node_path)
	if err != nil {
		return nil, err
	}

	template_path, err := find_path("file", filepath.Join(root_dir, CFG_TEMPLATE_DIR), group)
	if err != nil {
		return nil, err
	}

	if err := apply_template(template_path, node); err != nil {
		return nil, err
	}

	set_config_value(CFG_CURRENT_NODE_PREFIX+group, title)

	for i := CFG_GROUP_DEPTH[group] + 1; i < len(CFG_GROUP_DEPTH); i++ {
		set_config_value(CFG_CURRENT_NODE_PREFIX+CFG_DEPTH_GROUP[i], "")
	}

	Nodes = append(Nodes, node)

	populate_note_fields(node)

	return node, nil
}

func valid_node_group(group string) bool {
	_, exists := CFG_GROUP_DEPTH[strings.ToLower(group)]
	return exists
}

func get_base_directory(root_dir, group string) (string, error) {

	prefix := CFG_CURRENT_NODE_PREFIX
	group_depth := CFG_GROUP_DEPTH[group]
	parent_group := CFG_DEPTH_GROUP[group_depth-1]

	if CFG_GROUP_DEPTH[group] < 1 {
		return filepath.Join(root_dir, CFG_SEMESTER_DIR), nil
	}
	parent_title, err := get_config_value(prefix + parent_group)
	if err != nil {
		return "", err
	}
	return find_path("directory", root_dir, parent_title)
}

func initialize_node(group, title, node_path string) (*Node, error) {

	node := &Node{}
	if err := node.set_group(group); err != nil {
		return nil, err
	}
	if err := node.set_title(title); err != nil {
		return nil, err
	}
	if err := node.set_path(node_path); err != nil {
		return nil, err
	}

	if CFG_GROUP_DEPTH[group] > 0 {
		parent_group := CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1]
		config_current_parent, err := get_config_value(CFG_CURRENT_NODE_PREFIX + parent_group)
		if err != nil {
			return nil, err
		}

		var parent *Node

		for _, node := range Nodes {
			if node.get_title() == config_current_parent {
				parent = node
			}
		}

		if parent != nil {
			if err := node.set_parent(parent); err != nil {
				return nil, err
			}
		}
	}

	node.set_id("")
	return node, nil
}

func get_struct_field_names(data interface{}) []string {
	t := reflect.TypeOf(data)

	if t.Kind() != reflect.Struct {
		panic("Provided input is not a struct")
	}

	var field_names []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		field_names = append(field_names, field.Name)
	}

	return field_names
}
