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

type Course struct {
	Title    string
	Path     string
	Children []*Chapter
	Parent   *Semester
}

func get_courses(s *Semester) error {
	files, err := os.ReadDir(filepath.Join(s.Path, "courses"))

	if err != nil {
		return fmt.Errorf("Error retrieving courses: %w", err)
	}

	for _, f := range files {

		if c, err := find_course(f.Name(), s); c != nil && err == nil || !f.IsDir() {
			continue
		}

		new_course := &Course{
			Title: f.Name(),
			Path:  filepath.Join(s.Path, "courses", f.Name()),
		}

		s.Children = append(s.Children, new_course)
	}

	return nil
}

func find_course(title string, semester *Semester) (*Course, error) {
	for _, c := range semester.Children {
		if c.Title == title {
			return c, nil
		}
	}

	return nil, fmt.Errorf("Unable to find course '%v'.", title)
}

func create_course(title string, semester *Semester) (*Course, error) {

	if c, _ := find_course(title, semester); c != nil {
		show_warning(fmt.Sprintf("Course '%s' already exists! Creation aborted.", title))
		return c, nil
	}

	data, err := os.ReadFile(filepath.Join(ROOT_DIR, "data/templates/structure/course.json"))

	if err != nil {
		return nil, fmt.Errorf("Failed to read JSON file: %w", err)
	}

	var parsed_data map[string]interface{}

	if err := json.Unmarshal(data, &parsed_data); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %w", err)
	}

	if err := os.Mkdir(filepath.Join(semester.Path, "courses", title), 0755); err != nil {
		return nil, err
	}

	if root, exists := parsed_data["root"]; exists {
		if root_map, ok := root.(map[string]interface{}); ok {
			if err := create_structure(filepath.Join(semester.Path, "courses", title), root_map); err != nil {
				return nil, fmt.Errorf("Error creating directories: %w", err)
			}
		}
	}

	placeholders := map[string]string{
		"%%title%%":    title,
		"%%semester%%": semester.Title,
	}

	populate_latex_fields(filepath.Join(semester.Path, "courses", title, "lectures", "lec-master.tex"), placeholders)

	new_course := &Course{
		Title:  title,
		Path:   filepath.Join(semester.Path, "courses", title),
		Parent: semester,
	}

	semester.Children = append(semester.Children, new_course)

	return new_course, nil
}

func create_course_with_form() error {

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
						return errors.New("Course must have a name!")
					}
					return nil
				}),

			huh.NewNote().
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_SEMESTER.Title)

						for _, course := range CURRENT_SEMESTER.Children {
							t.Child("Course: " + course.Title)
						}

						display_title := "Untitled"

						if title != "" {
							display_title = title
						}

						t.Child(bold_style.Render("New Course: " + display_title))

						return tree_style.Render(t.String())
					}, &title,
				),

			huh.NewConfirm().
				TitleFunc(
					func() string {
						return fmt.Sprintf("Create '%s'?", title)
					}, &title).
				Value(&confirmed),
		).WithHeight(30),
	).WithTheme(huh.ThemeBase())

	form.Run()

	if confirmed {
		new_course, err := create_course(title, CURRENT_SEMESTER)

		if err != nil {
			return err
		}

		if new_course != nil {
			show_message(fmt.Sprintf("Created new course '%v'!", title))
		}
	}

	return nil
}
