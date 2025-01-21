package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {

	if _, err := build_tree(nil); err != nil {
		log.Fatal(err)
		return
	}

	for _, node := range Nodes {
		add_children_to_input_file(node)
	}

	handle_input()
}

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
		}
	}
}

func open_note(node *Node) error {

	CFG_NOTE_ARGUMENTS = append(CFG_NOTE_ARGUMENTS, node.get_path())

	cmd := exec.Command(CFG_EDITOR, CFG_NOTE_ARGUMENTS...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("Failed to open tex file: %s", err.Error())
	}

	return nil
}
