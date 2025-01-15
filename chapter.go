package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/tree"
)

type Chapter struct {
	Title    string
	Path     string
	Parent   *Course
	Children []*Lecture
}

func get_chapters(c *Course) error {
	files, err := os.ReadDir(filepath.Join(c.Path, CHAPTERS_PATH))

	if err != nil {
		return fmt.Errorf("Error retrieving chapters: %w", err)
	}

	for _, f := range files {

		if !f.IsDir() {
			continue
		}

		var exists bool = false

		for _, c := range c.Children {
			if c.Title == f.Name() {
				show_warning(fmt.Sprintf("Chapter '%v' already exists in the array. Addition aborted.", c.Title))
				exists = true
			}
		}

		if exists {
			continue
		}

		var new_chapter Chapter
		new_chapter.Title = f.Name()
		new_chapter.Path = filepath.Join(c.Path, CHAPTERS_PATH, f.Name())
		c.Children = append(c.Children, &new_chapter)
	}

	return nil
}

func chapter_exists(title string, course *Course) (bool, error) {
	for _, c := range course.Children {
		if c.Title == title {
			return true, nil
		}
	}

	return false, nil
}

func find_chapter(title string, course *Course) (*Chapter, error) {

	for _, c := range course.Children {
		if c.Title == title {
			return c, nil
		}
	}

	return nil, fmt.Errorf("Unable to find chapter '%v'.", title)
}

func create_chapter(title string, course *Course) (*Chapter, error) {

	exists, err := chapter_exists(title, course)

	if err != nil {
		show_error(fmt.Sprintf("Error checking chapter existence: %v", err))
	}

	if exists {
		show_warning(fmt.Sprintf("Chapter '%s' already exists! Creation aborted.", title))

		var found_chapter *Chapter

		for _, c := range course.Children {
			if c.Title == title {
				found_chapter = c
			}
		}

		return found_chapter, nil
	}

	data, err := os.ReadFile(filepath.Join(ROOT_DIR, CHAPTER_TEMPLATE_PATH))

	if err != nil {
		return nil, fmt.Errorf("Failed to read JSON file: %w", err)
	}

	var parsed_data map[string]interface{}

	if err := json.Unmarshal(data, &parsed_data); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %w", err)
	}

	os.Mkdir(filepath.Join(course.Path, CHAPTERS_PATH, title), 0755)

	if root, exists := parsed_data["root"]; exists {
		if root_map, ok := root.(map[string]interface{}); ok {
			if err := create_structure(filepath.Join(course.Path, CHAPTERS_PATH, title), root_map); err != nil {
				return nil, fmt.Errorf("Error creating directories: %w", err)
			}
		}
	}

	// Modify created files, if needed

	placeholders := map[string]string{
		"%%chapter%%": title,
	}

	populate_latex_fields(filepath.Join(course.Path, CHAPTERS_PATH, title, CHAPTER_COMPOSITE_PATH), placeholders)

	new_chapter := &Chapter{
		Title:  title,
		Path:   filepath.Join(course.Path, CHAPTERS_PATH, title),
		Parent: course,
	}

	course.Children = append(course.Children, new_chapter)

	set_current_chapter(new_chapter)

	add_chapter_to_latex_mainfile(new_chapter, filepath.Join(course.Path, LECTURES_MASTER_PATH), filepath.Join(course.Path, CHAPTERS_PATH, title, CHAPTER_COMPOSITE_PATH))

	return new_chapter, nil
}

func add_chapter_to_latex_mainfile(chap *Chapter, main_path, chapter_path string) error {
	// Read the LaTeX file
	data, err := os.ReadFile(main_path)
	if err != nil {
		return fmt.Errorf("Error reading latex mainfile: %v", err)
	}
	content := string(data)

	// Find the `% CHAPTERS` marker
	placeholder := "% CHAPTERS"
	index := strings.Index(content, placeholder)
	if index == -1 {
		return fmt.Errorf("placeholder '%s' not found in latex mainfile", placeholder)
	}

	// Build the `\input{}` line
	new_chapter := fmt.Sprintf("\\input{%s} %% %s", chapter_path, chap.Title)

	// Check if the chapter already exists
	if strings.Contains(content, new_chapter) {
		return fmt.Errorf("chapter '%s' already included", chapter_path)
	}

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Find the last included chapter in the `% CHAPTERS` section
	insertIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == placeholder {
			insertIndex = i
			break
		}
	}

	if insertIndex == -1 {
		return fmt.Errorf("unable to locate placeholder '%s' in lines", placeholder)
	}

	// Append the new chapter at the end of the `% CHAPTERS` section
	for i := insertIndex + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" || !strings.HasPrefix(strings.TrimSpace(lines[i]), `\input{`) {
			lines = append(lines[:i], append([]string{new_chapter}, lines[i:]...)...)
			break
		}
	}

	// Join the lines back into a single string
	updated_content := strings.Join(lines, "\n")

	// Write the updated content back to the file
	err = os.WriteFile(main_path, []byte(updated_content), 0755)
	if err != nil {
		return fmt.Errorf("error writing updated mainfile: %v", err)
	}

	return nil
}

func remove_chapter_from_latex_mainfile(chapTitle, mainPath string) error {
	// Read the LaTeX file
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("error reading latex mainfile: %v", err)
	}
	content := string(data)

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Track if the chapter was removed
	removed := false

	// Loop through lines to find and remove the chapter
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if the line is an `\input{}` line and contains the chapter title
		if strings.HasPrefix(trimmed, `\input{`) && strings.Contains(trimmed, fmt.Sprintf("%% %s", chapTitle)) {
			// Remove the line
			lines = append(lines[:i], lines[i+1:]...)
			removed = true
			break
		}
	}

	// If no chapter was removed, return an error
	if !removed {
		return fmt.Errorf("chapter '%s' not found in the mainfile", chapTitle)
	}

	// Join the lines back into a single string
	updatedContent := strings.Join(lines, "\n")

	// Write the updated content back to the file
	err = os.WriteFile(mainPath, []byte(updatedContent), 0755)
	if err != nil {
		return fmt.Errorf("error writing updated mainfile: %v", err)
	}

	return nil
}

func create_chapter_with_form() error {

	if CURRENT_COURSE == nil {
		show_warning("No courses exist. Try creating one!")
		return fmt.Errorf("Current course doesn't exist.")
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
						return errors.New("Chapter must have a name!")
					}
					return nil
				}),

			huh.NewNote().
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_COURSE.Title)

						for _, chapter := range CURRENT_COURSE.Children {
							t.Child("Chapter: " + chapter.Title)
						}

						display_title := "Untitled"

						if title != "" {
							display_title = title
						}

						t.Child(bold_style.Render("New Chapter: " + display_title))

						return tree_style.Render(t.String())
					}, &title,
				).Height(len(CURRENT_COURSE.Children) + 5),

			huh.NewConfirm().
				TitleFunc(
					func() string {
						if title == "" {
							return fmt.Sprint("Create 'Untitled'?")
						}
						return fmt.Sprintf("Create '%s'?", title)
					}, &title).
				Value(&confirmed),
		).Title("Creating a new Chapter"),
	).WithTheme(huh.ThemeBase())

	form.Run()

	if confirmed {
		new_chapter, err := create_chapter(title, CURRENT_COURSE)

		if err != nil {
			return err
		}

		if new_chapter != nil {
			show_message(fmt.Sprintf("Created new chapter '%v' (or already exists)!", title))
		}
	}

	return nil
}

func remove_chapter_with_form() error {

	var choices []string
	confirmed := false
	double_confirmed := false

	var options []string

	for _, s := range CURRENT_COURSE.Children {
		options = append(options, s.Title)
	}

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewMultiSelect[string]().
				Title("Choose Chapter(s) to delete").
				Options(huh.NewOptions(options...)...).
				Value(&choices).
				Validate(func(s []string) error {
					if len(choices) < 1 {
						return fmt.Errorf("You must choose course(s) to delete!")
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
			chapter, err := find_chapter(choice, CURRENT_COURSE)

			if err != nil || chapter == nil {
				return fmt.Errorf("Could not find chapter '%s'.", choice)
			}

			os.RemoveAll(chapter.Path) // remove from filestructure

			CURRENT_COURSE.Children = remove_chapter(CURRENT_COURSE.Children, chapter)

			remove_chapter_from_latex_mainfile(chapter.Title, filepath.Join(CURRENT_COURSE.Path, LECTURES_MASTER_PATH))
		}
	}

	return nil
}

func remove_chapter(chapters []*Chapter, toRemove *Chapter) []*Chapter {
	for i, chapter := range chapters {
		if chapter == toRemove {
			return append(chapters[:i], chapters[i+1:]...)
		}
	}
	return chapters
}
