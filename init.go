package main

import (
	"os"
	"path/filepath"
	"strings"
)

func build_tree(parent *Node) (*Node, error) {

	if parent == nil {
		project_root, err := get_config_value(CFG_ROOT_FIELD)

		if err != nil {
			return nil, err
		}

		files, err := os.ReadDir(filepath.Join(project_root, CFG_SEMESTER_DIR))
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !file.IsDir() {
				continue
			}

			info_file, _ := find_path("file", filepath.Join(project_root, CFG_SEMESTER_DIR, file.Name()), "info")

			id, _ := read_json_value(info_file, "id")

			node := &Node{}

			node.set_title(file.Name())
			node.set_group(CFG_DEPTH_GROUP[0])
			node.set_path(filepath.Join(project_root, CFG_SEMESTER_DIR, file.Name()))
			node.set_id(id)
			node.set_parent(parent)

			Nodes = append(Nodes, node)

			build_tree(node)
		}

		return nil, nil
	}

	parent_depth := CFG_GROUP_DEPTH[parent.get_group()]
	child_group := CFG_DEPTH_GROUP[parent_depth+1]

	children_directory, err := find_path("directory", parent.get_path(), child_group)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(children_directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {

		var title, id, group string

		if !file.IsDir() {

			title = strings.Split(file.Name(), filepath.Ext(file.Name()))[0]
			group = child_group

		} else {

			info_file, _ := find_path("file", filepath.Join(children_directory, file.Name()), "info")

			title, _ = read_json_value(info_file, "title")
			id, _ = read_json_value(info_file, "id")
			group, _ = read_json_value(info_file, "group")
		}

		node := &Node{}

		node.set_title(title)
		node.set_group(group)
		node.set_path(filepath.Join(children_directory, file.Name()))
		node.set_id(id)
		node.set_parent(parent)

		Nodes = append(Nodes, node)

		build_tree(node)
	}

	return nil, nil
}
