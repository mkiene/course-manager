package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/mkiene/huh"
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
				Value(&choices[0]).
				Validate(func(s string) error {
					for _, node := range Nodes {
						if node.get_depth() != CFG_GROUP_DEPTH[group] {
							continue
						}
						if CFG_GROUP_DEPTH[group] > 0 {
							current_parent_title, err := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1])
							if err != nil {
								return err
							}
							if node.get_parent().get_title() != current_parent_title {
								continue
							}
						}
						if s == node.get_title() {
							return fmt.Errorf("%v already exists: '%v'", group, node.get_title())
						}
					}

					return nil
				}),

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

					t.Child(bold_style.Render(fmt.Sprintf("new %v: %v", group, display_title)))

					return tree_style.Render(t.String())
				}, &choices),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Create new %v ''?", group)).
				TitleFunc(func() string {
					return fmt.Sprintf("Create new %v '%v'?", group, choices[0])
				}, &choices).
				Value(&confirm),
		),
	).WithTheme(form_theme).
		WithLayout(huh.LayoutStack).
		WithProgramOptions(tea.WithAltScreen())

	err := form.Run()

	if err != nil {
		return nil, err
	}

	if confirm {
		node, err := create_node(group, choices[0])
		if err != nil {
			return nil, err
		}

		if node.get_parent() != nil {
			add_children_to_input_file(node.get_parent())
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
	var group_nodes_objects []*Node

	if group_depth < 1 {
		for _, node := range Nodes {
			if node.get_group() == group {
				group_nodes = append(group_nodes, node.get_title())
				group_nodes_objects = append(group_nodes_objects, node)
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
			group_nodes_objects = append(group_nodes_objects, child)
		}

		if len(group_nodes) < 1 {
			return fmt.Errorf("no children to select")
		}
	}

	var choices [1]string

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewSelect[string]().
				Title(fmt.Sprintf("Choose a %v", group)).
				Options(huh.NewOptions(group_nodes...)...).
				Value(&choices[0]),

			huh.NewNote().
				Title("Preview").
				DescriptionFunc(func() string {
					root_title, _ := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1])
					root_title = fmt.Sprintf("%v: %v", CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1], root_title)
					if CFG_GROUP_DEPTH[group] < 1 {
						root_title = "root"
					}
					t := tree.New().Root(root_title)

					for _, node := range group_nodes_objects {
						if node.get_title() == choices[0] {
							branch, err := show_branch(node)
							if err != nil {
								return ""
							}
							t.Child(bold_style.Render(branch.String()))
						} else {
							t.Child(fmt.Sprintf("%v: %v", group, node.get_title()))
						}
					}

					return tree_style.Render(t.String())
				}, &choices),
		),
	).WithTheme(form_theme).
		WithLayout(huh.LayoutStack).
		WithProgramOptions(tea.WithAltScreen())

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

func node_deletion_form(group string) error {

	if !valid_node_group(group) {
		return fmt.Errorf("invalid group")
	}

	var group_nodes []string

	current_parent_title := ""

	if CFG_GROUP_DEPTH[group] > 0 {
		value, err := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[CFG_GROUP_DEPTH[group]-1])
		if err != nil {
			return err
		}
		current_parent_title = value
	}

	for _, node := range Nodes {
		if node.get_group() != group {
			continue
		}
		if CFG_GROUP_DEPTH[group] > 0 && node.get_parent().get_title() != current_parent_title {
			continue
		}

		group_nodes = append(group_nodes, node.get_title())
	}

	if len(group_nodes) < 1 {
		fmt.Printf("no nodes in group '%v' exist.\n", group)
		return nil
	}

	var choices [1]string

	var confirmed [2]bool

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewSelect[string]().Title(fmt.Sprintf("Choose a %v to remove", group)).
				Options(huh.NewOptions(group_nodes...)...).
				Value(&choices[0]),

			huh.NewNote().DescriptionFunc(func() string {

				var node *Node
				parent_title := "root"

				for _, listnode := range Nodes {
					if listnode.get_title() != choices[0] {
						continue
					}
					if CFG_GROUP_DEPTH[group] > 0 && listnode.get_parent().get_title() != current_parent_title {
						continue
					}
					node = listnode
				}

				if node.get_parent() != nil {
					parent_title = node.get_parent().get_title()
				}

				tree := tree.New().Root(parent_title)

				for _, zero := range Nodes {
					if zero.get_group() != group {
						continue
					}

					if zero.get_parent() != node.get_parent() {
						continue
					}

					if zero.get_id() == node.get_id() {
						tr, err := show_branch(zero)
						if err != nil {
							return "ERROR"
						}

						tree.Child(bold_style.Render(tr.String()))
					} else {
						tree.Child(fmt.Sprintf("%v: %v", group, zero.get_title()))
					}
				}

				return fmt.Sprintln(tree_style.Render(tree.String()))

			}, &choices),

			huh.NewConfirm().TitleFunc(func() string {
				return fmt.Sprintf("Remove '%v'?", choices[0])
			}, &choices).Value(&confirmed[0]).
				Description("This action cannot be undone."),
		),

		huh.NewGroup(

			huh.NewConfirm().
				Title("---").
				TitleFunc(func() string {
					return fmt.Sprintf("Are you sure you want to remove '%v'?", choices[0])
				}, &choices).Value(&confirmed[1]).
				DescriptionFunc(func() string {
					return "This action cannot be undone."
				}, &choices),
		).WithHideFunc(func() bool {
			if confirmed[0] {
				return false
			}

			return true
		}),
	).WithTheme(form_theme).
		WithLayout(huh.LayoutStack).
		WithProgramOptions(tea.WithAltScreen())

	if err := form.Run(); err != nil {
		return err
	}

	if confirmed[0] && confirmed[1] {
		for _, node := range Nodes {
			if node.get_group() == group && node.get_title() == choices[0] {
				err := remove_from_parent_input_file(node)
				if err != nil {
					return err
				}

				os.RemoveAll(node.get_path())

				break
			}
		}

		fmt.Printf("Removed %v '%v'.\n", group, choices[0])
	}

	return nil
}
