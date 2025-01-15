package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/tree"
)

func show_fatal(msg string) string {
	fmt.Println(fatal_style.Render("!FATAL!:", msg))

	return fatal_style.Render("!FATAL!:", msg)
}

func show_error(msg string) string {
	fmt.Println(error_style.Render("ERROR:", msg))

	return error_style.Render("ERROR:", msg)
}

func show_warning(msg string) string {
	fmt.Println(warning_style.Render("WARNING:", msg))

	return warning_style.Render("WARNING:", msg)
}

func show_message(msg string) string {
	fmt.Println(message_style.Render("MESSAGE:", msg))

	return message_style.Render("MESSAGE:", msg)
}

func get_json_value(path string, field string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Unable to locate JSON file: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", fmt.Errorf("Error reading JSON file: %w", err)
	}

	value, ok := result[field]
	if !ok {
		return "", fmt.Errorf("Field '%s' not found in JSON", field)
	}

	ret, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Field '%s' is not a string", field)
	}

	return ret, nil
}

func modify_json_value(path string, field string, new_value string) error {
	// Read the JSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Unable to locate JSON file: %w", err)
	}

	// Parse the JSON into a map
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("Error reading JSON file: %w", err)
	}

	// Check if the field exists
	if _, ok := result[field]; !ok {
		return fmt.Errorf("Field '%s' not found in JSON", field)
	}

	// Modify the value
	result[field] = new_value

	// Serialize the modified JSON
	updatedData, err := json.MarshalIndent(result, "", "  ") // Indent for readability
	if err != nil {
		return fmt.Errorf("Error writing JSON data: %w", err)
	}

	// Write the updated JSON back to the file
	err = os.WriteFile(path, updatedData, 0644) // 0644 ensures the file is readable and writable
	if err != nil {
		return fmt.Errorf("Unable to write to JSON file: %w", err)
	}

	return nil
}

func build_tree(node interface{}) *tree.Tree {
	// Create a new tree for the current node
	t := tree.New()

	// Determine the label for the current node
	var label string
	switch n := node.(type) {
	case *Semester:
		if CURRENT_SEMESTER != nil && CURRENT_SEMESTER.Title == n.Title {
			label = "Semester (current): " + n.Title
		} else {
			label = "Semester: " + n.Title
		}
	case *Course:
		if CURRENT_COURSE != nil && CURRENT_COURSE.Title == n.Title {
			label = "Course (current): " + n.Title
		} else {
			label = "Course: " + n.Title
		}
	case *Chapter:
		if CURRENT_CHAPTER != nil && CURRENT_CHAPTER.Title == n.Title {
			label = "Chapter (current): " + n.Title
		} else {
			label = "Chapter: " + n.Title
		}
	case *Lecture:
		if CURRENT_LECTURE != nil && CURRENT_LECTURE.Title == n.Title {
			label = "Lecture (current): " + n.Title
		} else {
			label = "Lecture: " + n.Title
		}
	default:
		return tree.New() // Return an empty tree instead of nil for unknown types
	}

	t.Root(label)

	// Recursively process children
	switch n := node.(type) {
	case *Semester:
		for _, child := range n.Children {
			childTree := build_tree(child)
			if childTree != nil {
				t.Child(childTree)
			}
		}
	case *Course:
		for _, child := range n.Children {
			childTree := build_tree(child)
			if childTree != nil {
				t.Child(childTree)
			}
		}
	case *Chapter:
		for _, child := range n.Children {
			childTree := build_tree(child)
			if childTree != nil {
				t.Child(childTree)
			}
		}
	case *Lecture:
		// Lectures have no children; nothing to do here
	}

	return t
}

func update_tree() {

	// First, find all semesters under the specified directory in the config file
	semesters_dir := filepath.Join(ROOT_DIR, SEMESTERS_DIR)

	get_semesters(semesters_dir)

	// Next, find all of the retrieved semester's courses
	for _, s := range Semesters {
		get_courses(s)

		// For each course, find all chapters and lectures
		for _, co := range s.Children {
			get_chapters(co)

			for _, ch := range co.Children {
				get_lectures(ch)
			}
		}
	}
}

func create_structure(path string, structure map[string]interface{}) error {
	for name, value := range structure {
		path := filepath.Join(path, name)
		switch v := value.(type) {
		case string:
			// Create the file from the template
			data, err := os.ReadFile(value.(string))

			if err != nil {
				return fmt.Errorf("Failed to read template file: %w", err)
			}

			os.WriteFile(path, data, 0755)

		case map[string]interface{}:
			// Create directory and recurse
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("Failed to create directory %s: %w", path, err)
			}
			if err := create_structure(path, v); err != nil {
				return err
			}
		default:
			show_warning(fmt.Sprintf("Unknown type for %s.", name))
		}
	}
	return nil
}

func populate_latex_fields(path string, placeholders map[string]string) error {

	data, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("Unable to open latex file: %w", err)
	}

	file_content := string(data)

	for placeholder, value := range placeholders {
		file_content = strings.ReplaceAll(file_content, placeholder, value)
	}

	os.WriteFile(path, []byte(file_content), 0755)

	return nil
}

func get_currents() error {

	current_semester_title, err := get_json_value(CONFIG_DIR, "current-semester")

	if err != nil {
		return err
	}

	if current_semester_title == "" {
		return nil
	}

	CURRENT_SEMESTER, err = find_semester(current_semester_title)

	if err != nil {
		return err
	}

	current_course_title, err := get_json_value(CONFIG_DIR, "current-course")

	if err != nil {
		return err
	}

	if current_course_title == "" {
		return nil
	}

	CURRENT_COURSE, err = find_course(current_course_title, CURRENT_SEMESTER)

	if err != nil {
		return err
	}

	current_chapter_title, err := get_json_value(CONFIG_DIR, "current-chapter")

	if err != nil {
		return err
	}

	if current_chapter_title == "" {
		return nil
	}

	CURRENT_CHAPTER, err = find_chapter(current_chapter_title, CURRENT_COURSE)

	if err != nil {
		return err
	}

	current_lecture_title, err := get_json_value(CONFIG_DIR, "current-lecture")

	if err != nil {
		return err
	}

	if current_lecture_title == "" {
		return nil
	}

	CURRENT_LECTURE, err = find_lecture(current_lecture_title, CURRENT_CHAPTER)

	if err != nil {
		return err
	}

	return nil
}

func set_currents() {

	var semester_titles []string

	var chosen_semester string

	for _, semester := range Semesters {
		semester_titles = append(semester_titles, semester.Title)
	}

	semester_form := huh.NewForm(

		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Select a Semester").
				Options(huh.NewOptions(semester_titles...)...).
				Value(&chosen_semester).Height(1),

			huh.NewNote().
				Title("Filetree").
				DescriptionFunc(
					func() string {

						t := tree.New().Root("semesters")

						for _, s := range Semesters {
							if s.Title == chosen_semester {
								t.Child(bold_style.Render(build_tree(s).String()))
							} else {
								t.Child(build_tree(s))
							}
						}

						return tree_style.Render(t.String())
					}, &chosen_semester).Height(40),
		),
	).WithTheme(huh.ThemeBase())

	err := semester_form.Run()

	if err != nil {
		return
	}

	semester_obj, err := find_semester(chosen_semester)

	if err != nil {
		return
	}

	set_current_semester(semester_obj)

	if len(semester_obj.Children) < 1 {
		return
	}

	/////

	var course_titles []string

	var chosen_course string

	for _, course := range CURRENT_SEMESTER.Children {
		course_titles = append(course_titles, course.Title)
	}

	course_form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Course").
				Options(huh.NewOptions(course_titles...)...).
				Value(&chosen_course).Height(1),

			huh.NewNote().
				Title("Filetree").
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_SEMESTER.Title)

						for _, s := range CURRENT_SEMESTER.Children {
							if s.Title == chosen_course {
								t.Child(bold_style.Render(build_tree(s).String()))
							} else {
								t.Child(build_tree(s))
							}
						}

						return tree_style.Render(t.String())
					}, &chosen_course).Height(40),
		).WithHeight(10),
	).WithTheme(huh.ThemeBase())

	err = course_form.Run()

	if err != nil {
		return
	}

	course_obj, err := find_course(chosen_course, CURRENT_SEMESTER)

	if err != nil {
		return
	}

	set_current_course(course_obj)

	if len(course_obj.Children) < 1 {
		return
	}

	/////

	var chapter_titles []string

	var chosen_chapter string

	for _, chapter := range CURRENT_COURSE.Children {
		chapter_titles = append(chapter_titles, chapter.Title)
	}

	chapter_form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Chapter").
				Options(huh.NewOptions(chapter_titles...)...).
				Value(&chosen_chapter).Height(1),

			huh.NewNote().
				Title("Filetree").
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_COURSE.Title)

						for _, s := range CURRENT_COURSE.Children {
							if s.Title == chosen_chapter {
								t.Child(bold_style.Render(build_tree(s).String()))
							} else {
								t.Child(build_tree(s))
							}
						}

						return tree_style.Render(t.String())
					}, &chosen_chapter).Height(40),
		).WithHeight(10),
	).WithTheme(huh.ThemeBase())

	err = chapter_form.Run()

	if err != nil {
		return
	}

	chapter_obj, err := find_chapter(chosen_chapter, CURRENT_COURSE)

	if err != nil {
		return
	}

	set_current_chapter(chapter_obj)

	if len(chapter_obj.Children) < 1 {
		return
	}

	/////

	var lecture_titles []string

	var chosen_lecture string

	for _, lecture := range CURRENT_CHAPTER.Children {
		lecture_titles = append(lecture_titles, lecture.Title)
	}

	lecture_form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Lecture").
				Options(huh.NewOptions(lecture_titles...)...).
				Value(&chosen_lecture).Height(1),

			huh.NewNote().
				Title("Filetree").
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_CHAPTER.Title)

						for _, s := range CURRENT_CHAPTER.Children {
							if s.Title == chosen_lecture {
								t.Child(bold_style.Render(build_tree(s).String()))
							} else {
								t.Child(build_tree(s))
							}
						}

						return tree_style.Render(t.String())
					}, &chosen_lecture).Height(40),
		).WithHeight(10),
	).WithTheme(huh.ThemeBase())

	err = lecture_form.Run()

	if err != nil {
		return
	}

	lecture_obj, err := find_lecture(chosen_lecture, CURRENT_CHAPTER)

	if err != nil {
		return
	}

	set_current_lecture(lecture_obj)
}

func set_current_semester(semester *Semester) error {
	err := modify_json_value(CONFIG_DIR, "current-semester", semester.Title)

	if err != nil {
		return err
	}

	modify_json_value(CONFIG_DIR, "current-course", "")
	modify_json_value(CONFIG_DIR, "current-chapter", "")
	modify_json_value(CONFIG_DIR, "current-lecture", "")

	get_currents()

	return nil
}

func set_current_course(course *Course) error {
	err := modify_json_value(CONFIG_DIR, "current-course", course.Title)

	if err != nil {
		return err
	}

	modify_json_value(CONFIG_DIR, "current-chapter", "")
	modify_json_value(CONFIG_DIR, "current-lecture", "")

	get_currents()

	return nil
}

func set_current_chapter(chapter *Chapter) error {
	err := modify_json_value(CONFIG_DIR, "current-chapter", chapter.Title)

	if err != nil {
		return err
	}

	modify_json_value(CONFIG_DIR, "current-lecture", "")

	get_currents()

	return nil
}

func set_current_lecture(lecture *Lecture) error {
	err := modify_json_value(CONFIG_DIR, "current-lecture", lecture.Title)

	if err != nil {
		return err
	}

	get_currents()

	return nil
}
