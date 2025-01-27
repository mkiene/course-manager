package main

import (
	"log"
)

func main() {

	if _, err := build_tree(nil); err != nil {
		log.Fatal(err)
		return
	}

	handle_input()

	for _, node := range Nodes {
		add_children_to_input_file(node)
	}
}
