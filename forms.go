package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func node_creation_form(group string) (*Node, error) {

	if !valid_node_group(group) {
		return nil, fmt.Errorf("invalid node group '%v'", group)
	}

	var choices [1]string
	var confirm bool

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewInput().
				Title("Provide a title").
				Value(&choices[0]),

			huh.NewConfirm().
				TitleFunc(func() string {
					return fmt.Sprintf("Create new %v '%v'?", group, choices[0])
				}, &choices).
				Value(&confirm),
		),
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
