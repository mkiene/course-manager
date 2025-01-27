package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/tree"
)

/*




































 */

func node_creation_form(group string) (*Node, error) {

	if !valid_node_group(group) {
		return nil, fmt.Errorf("invalid node group '%v'", group)
	}

	num_group_members := 0

	for _, node := range Nodes {
		if node.get_group() == group {
			num_group_members++
		}
	}

	var choices [1]string
	var confirm bool

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewInput().
				Title("Provide a title").
				Value(&choices[0]),

			huh.NewNote().Title("Preview").
				DescriptionFunc(func() string {

					t := tree.New().Root(CFG_DEPTH_GROUP[0])

					if CFG_GROUP_DEPTH[group] > 0 {
						title, err := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1])
						if err != nil {
							return ""
						}

						var parent *Node

						for _, node := range Nodes {
							if node.get_title() == title {
								parent = node
							}
						}

						if parent == nil {
							return "ERROR"
						}

						t.Root(parent.get_title())

						for _, child := range parent.get_children() {
							t.Child(fmt.Sprintf("%v: %v", group, child.get_title()))
						}

					} else {
						for _, node := range Nodes {
							if node.get_depth() == 0 {
								t.Child(fmt.Sprintf("%v: %v", group, node.get_title()))
							}
						}
					}

					display_title := "Untitled"

					if choices[0] != "" {
						display_title = choices[0]
					}

					t.Child(fmt.Sprintf("new %v: %v", group, display_title))

					return tree_style.Render(t.String())
				}, &choices).
				Height(num_group_members+5),

			huh.NewConfirm().
				TitleFunc(func() string {
					return fmt.Sprintf("Create new %v '%v'?", group, choices[0])
				}, &choices).
				Value(&confirm),
		).WithTheme(huh.ThemeBase()),
	)

	err := form.Run()

	if err != nil {
		return nil, err
	}

	if confirm {
		node, err := create_node(group, choices[0])
		if err != nil {
			return nil, err
		}

		return node, nil
	}

	return nil, fmt.Errorf("%v creation aborted", group)
}

func set_currents_form(group string) error {

	if !valid_node_group(group) {
		return fmt.Errorf("invalid group")
	}

	group_depth := CFG_GROUP_DEPTH[group]
	var group_nodes []string

	if group_depth < 1 {
		for _, node := range Nodes {
			if node.get_group() == group {
				group_nodes = append(group_nodes, node.get_title())
			}
		}
	} else {
		var current_parent *Node
		current_parent_title, err := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[group_depth-1])
		if err != nil {
			return err
		}
		for _, node := range Nodes {
			if node.get_title() == current_parent_title {
				current_parent = node
				break
			}
		}
		if current_parent == nil {
			return fmt.Errorf("couldn't find current parent")
		}
		for _, child := range current_parent.get_children() {
			if child.get_title() == "" {
				continue
			}
			group_nodes = append(group_nodes, child.get_title())
		}
	}

	var choices [1]string

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewSelect[string]().
				Title(fmt.Sprintf("Choose a %v", group)).
				Options(huh.NewOptions(group_nodes...)...).
				Value(&choices[0]),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	err = set_config_value(CFG_CURRENT_NODE_PREFIX+group, choices[0])

	if group_depth < len(CFG_GROUP_DEPTH)-1 {
		set_currents_form(CFG_DEPTH_GROUP[group_depth+1])
	}

	return nil
}
