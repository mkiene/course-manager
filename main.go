package main

import (
	"fmt"
	"log"
)

func main() {

	if _, err := build_tree(nil); err != nil {
		log.Fatal(err)
		return
	}

	current_semester, _ := get_config_value(CFG_CURRENT_NODE_PREFIX + CFG_DEPTH_GROUP[0])

	found_current_semester := false

	for _, node := range Nodes {
		if node.get_depth() != 0 {
			continue
		}
		if node.get_title() == current_semester {
			found_current_semester = true
			if ok, errgroup := validate_currents(node); !ok {
				fmt.Printf("Invalid current %v. Please choose one:\n", errgroup)
				set_currents_form(errgroup)
			}
			break
		}
	}

	if !found_current_semester {
		fmt.Println("Unable to find current semester. Please choose one:")
		set_currents_form(CFG_DEPTH_GROUP[0])
		return
	}

	handle_input()
}
