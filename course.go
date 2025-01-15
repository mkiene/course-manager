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
	files, err := os.ReadDir(filepath.Join(s.Path, COURSES_PATH))

	if err != nil {
		return fmt.Errorf("Error retrieving courses: %w", err)
	}

	for _, f := range files {

		if c, err := find_course(f.Name(), s); c != nil && err == nil || !f.IsDir() {
			continue
		}

		new_course := &Course{
			Title: f.Name(),
			Path:  filepath.Join(s.Path, COURSES_PATH, f.Name()),
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

	data, err := os.ReadFile(filepath.Join(ROOT_DIR, COURSE_TEMPLATE_PATH))

	if err != nil {
		return nil, fmt.Errorf("Failed to read JSON file: %w", err)
	}

	var parsed_data map[string]interface{}

	if err := json.Unmarshal(data, &parsed_data); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %w", err)
	}

	if err := os.Mkdir(filepath.Join(semester.Path, COURSES_PATH, title), 0755); err != nil {
		return nil, err
	}

	if root, exists := parsed_data["root"]; exists {
		if root_map, ok := root.(map[string]interface{}); ok {
			if err := create_structure(filepath.Join(semester.Path, COURSES_PATH, title), root_map); err != nil {
				return nil, fmt.Errorf("Error creating directories: %w", err)
			}
		}
	}

	placeholders := map[string]string{
		"%%title%%":    title,
		"%%semester%%": semester.Title,
	}

	populate_latex_fields(filepath.Join(semester.Path, COURSES_PATH, title, LECTURES_MASTER_PATH), placeholders)

	new_course := &Course{
		Title:  title,
		Path:   filepath.Join(semester.Path, COURSES_PATH, title),
		Parent: semester,
	}

	semester.Children = append(semester.Children, new_course)

	set_current_course(new_course)

	return new_course, nil
}

func create_course_with_form() error {

	if CURRENT_SEMESTER == nil {
		show_warning("No semesters exist. Try creating one!")
		return fmt.Errorf("Current semester doesn't exist.")
	}

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
				Title("Filetree").
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
				).Height(len(CURRENT_SEMESTER.Children)+5),

			huh.NewConfirm().
				TitleFunc(
					func() string {
						if title == "" {
							return fmt.Sprint("Create 'Untitled'?")
						}
						return fmt.Sprintf("Create '%s'?", title)
					}, &title).
				Value(&confirmed),
		).Title("Creating a new Course"),
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

func remove_course_with_form() error {

	if len(Semesters) < 1 {
		return nil
	}

	var choices []string
	confirmed := false
	double_confirmed := false

	var options []string

	for _, s := range CURRENT_SEMESTER.Children {
		options = append(options, s.Title)
	}

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewMultiSelect[string]().
				Title("Choose a Course to delete").
				Options(huh.NewOptions(options...)...).
				Value(&choices).
				Validate(func(s []string) error {
					if len(choices) < 1 {
						return fmt.Errorf("You must choose a course(s) to delete!")
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
			course, err := find_course(choice, CURRENT_SEMESTER)

			if err != nil || course == nil {
				return fmt.Errorf("Could not find course '%s'.", choice)
			}

			os.RemoveAll(course.Path) // remove from filestructure

			CURRENT_SEMESTER.Children = remove_course(CURRENT_SEMESTER.Children, course)
		}
	}

	return nil
}

func remove_course(courses []*Course, toRemove *Course) []*Course {
	for i, course := range courses {
		if course == toRemove {
			return append(courses[:i], courses[i+1:]...)
		}
	}
	return courses
}
