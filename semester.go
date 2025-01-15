package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/tree"
)

type Semester struct {
	Title    string
	Path     string
	Children []*Course
}

var Semesters []*Semester

func get_semesters(path string) error {

	files, err := os.ReadDir(path)

	if err != nil {
		return err
	}

	for _, file := range files {

		if semester, err := find_semester(file.Name()); semester != nil && err == nil || !file.IsDir() {
			continue
		}

		new_semester := &Semester{
			Title: file.Name(),
			Path:  filepath.Join(path, file.Name()),
		}

		Semesters = append(Semesters, new_semester)
	}

	return nil
}

func find_semester(title string) (*Semester, error) {

	for _, s := range Semesters {
		if s.Title == title {
			return s, nil
		}
	}

	return nil, fmt.Errorf("Unable to find semester '%s'.", title)
}

func create_semester(title string) (*Semester, error) {

	if found_semester, _ := find_semester(title); found_semester != nil {

		show_warning(fmt.Sprintf("Semester '%s' already exists! Creation aborted.", title))
		return found_semester, nil
	}

	data, err := os.ReadFile(filepath.Join(ROOT_DIR, SEMESTER_TEMPLATE_PATH))

	if err != nil {
		return nil, fmt.Errorf("Failed to read JSON file: %w", err)
	}

	var parsed_data map[string]interface{}

	if err := json.Unmarshal(data, &parsed_data); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %w", err)
	}

	os.Mkdir(filepath.Join(ROOT_DIR, SEMESTERS_DIR, title), 0755)

	if root, exists := parsed_data["root"]; exists {
		if root_map, ok := root.(map[string]interface{}); ok {
			if err := create_structure(filepath.Join(ROOT_DIR, SEMESTERS_DIR, title), root_map); err != nil {
				return nil, fmt.Errorf("Error creating directories: %w", err)
			}
		}
	}

	new_semester := &Semester{
		Title: title,
		Path:  filepath.Join(ROOT_DIR, SEMESTERS_DIR, title),
	}

	Semesters = append(Semesters, new_semester)

	set_current_semester(new_semester)

	return new_semester, nil
}

func create_semester_with_form() error {

	var title string
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(

			huh.NewInput().
				Title("Enter a Title").
				Placeholder("Untitled").
				Value(&title).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("Semester must have a name!")
					}
					return nil
				}),

			huh.NewNote().
				Title("Filetree").
				DescriptionFunc(
					func() string {

						t := tree.New().Root("Semesters")

						for _, s := range Semesters {
							t.Child("Semester: " + s.Title)
						}

						display_title := "Untitled"

						if title != "" {
							display_title = title
						}

						t.Child(bold_style.Render("New Semester: " + display_title))

						return tree_style.Render(t.String())
					}, &title,
				).Height(len(Semesters)+5),

			huh.NewConfirm().
				TitleFunc(
					func() string {

						display_title := "Untitled"

						if title != "" {
							display_title = title
						}

						return fmt.Sprintf("Create '%s'?", display_title)
					}, &title).
				Value(&confirmed),
		).Title("Creating a new Semester"),
	).WithTheme(huh.ThemeBase())

	form.Run()

	if confirmed {
		new_semester, err := create_semester(title)

		if err != nil {
			return err
		}

		if new_semester != nil {
			show_message(fmt.Sprintf("Created new semester '%v'!", title))
		}
	}

	return nil
}

func remove_semester_with_form() error {

	if len(Semesters) < 1 {
		return nil
	}

	var choices []string
	confirmed := false
	double_confirmed := false

	var options []string

	for _, s := range Semesters {
		options = append(options, s.Title)
	}

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewMultiSelect[string]().
				Title("Choose a Semester to delete").
				Options(huh.NewOptions(options...)...).
				Value(&choices).
				Validate(func(s []string) error {
					if len(choices) < 1 {
						return fmt.Errorf("You must choose a semester(s) to delete!")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Make a selection").
				TitleFunc(
					func() string {
						return fmt.Sprintf("Irreversibly delete '%s'?", choices)
					}, &choices).
				Value(&confirmed),

			huh.NewConfirm().
				Title("Make a selection").
				TitleFunc(
					func() string {
						return fmt.Sprintf("Are you sure you want to irreversibly delete '%s'?", choices)
					}, &choices).
				Value(&double_confirmed),
		),
	)

	err := form.Run()

	if err != nil {
		return err
	}

	if confirmed && double_confirmed {

		for _, choice := range choices {
			sem, err := find_semester(choice)

			if err != nil || sem == nil {
				return fmt.Errorf("Could not find semester '%s'.", choice)
			}

			os.RemoveAll(sem.Path) // remove from filestructure

			Semesters = remove_semester(Semesters, sem)
		}
	}

	return nil
}

func remove_semester(semesters []*Semester, toRemove *Semester) []*Semester {
	for i, semester := range semesters {
		if semester == toRemove {
			// Remove the element by creating a new slice without it
			return append(semesters[:i], semesters[i+1:]...)
		}
	}
	return semesters // Return the original slice if the element is not found
}
