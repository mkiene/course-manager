package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss/tree"
)

func handle_input() {

	var args []string

	for i, arg := range os.Args {
		if i < 1 {
			continue
		}

		args = append(args, get_alias_group(arg))
	}

	if len(args) > 0 {

		switch args[0] {

		case "current":
			if len(args) > 1 {
				if valid_node_group(args[1]) {
					set_currents_form(args[1])
				}
			} else {
				err := set_currents_form(CFG_DEPTH_GROUP[0])
				if err != nil {
					log.Fatal(err)
				}
			}

		case "new":
			if len(args) > 1 {
				if valid_node_group(args[1]) {
					_, err := node_creation_form(args[1])
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}

		case "remove":
			if len(args) > 1 {
				if err := node_deletion_form(args[1]); err != nil {
					log.Fatal(err)
				}
			}

		case "lecture":
			current_lecture, err := get_config_value("current-lecture")
			if err != nil {
				log.Fatal(err)
				return
			}

			for _, node := range Nodes {
				if node.get_title() == current_lecture {
					open_note(node)
				}
			}

		case "tree":
			if len(args) > 1 {
				if valid_node_group(args[1]) {
					current_node_title, err := get_config_value(CFG_CURRENT_NODE_PREFIX + args[1])
					if err != nil {
						log.Fatal(err)
						return
					}
					tree := tree.New().Root(current_node_title)
					for _, node := range Nodes {
						if node.get_title() == current_node_title {
							tree.Child(show_branch(node))
							break
						}
					}
					fmt.Println(tree_style.Render(tree.String()))
				}
			} else {

				tree := tree.New().Root("root")

				for _, node := range Nodes {
					if node.get_depth() == 0 {
						tr, err := show_branch(node)
						if err != nil {
							log.Fatal(err)
							return
						}

						tree.Child(tr)
					}
				}

				fmt.Println(tree_style.Render(tree.String()))
			}
		}
	}

}

func show_branch(node *Node) (*tree.Tree, error) {

	t := tree.New().Root(fmt.Sprintf("%v: %v", node.get_group(), node.get_title()))

	for _, child := range node.Children {
		child_tree, err := show_branch(child)
		if err != nil {
			return nil, err
		}

		t.Child(child_tree)
	}

	return t, nil
}
