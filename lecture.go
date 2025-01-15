package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/tree"
)

type Lecture struct {
	Title  string
	Path   string
	Parent *Chapter
}

func get_lectures(c *Chapter) error {
	files, err := os.ReadDir(c.Path)

	if err != nil {
		return fmt.Errorf("Error retrieving lectures: %w", err)
	}

	for _, f := range files {

		if f.IsDir() {
			continue
		}

		true_name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

		if true_name == "composite" {
			continue
		}

		var exists bool = false

		for _, l := range c.Children {
			if l.Title == true_name {
				show_warning(fmt.Sprintf("Lecture '%v' already exists in the array. Addition aborted.", l.Title))
				exists = true
			}
		}

		if exists {
			continue
		}

		var new_lecture Lecture
		new_lecture.Title = true_name
		new_lecture.Path = filepath.Join(c.Path, true_name+".tex")
		c.Children = append(c.Children, &new_lecture)
	}

	return nil
}

func lecture_exists(title string, chapter *Chapter) (bool, error) {
	for _, l := range chapter.Children {
		if l.Title == title {
			return true, nil
		}
	}

	return false, nil
}

func find_lecture(title string, chapter *Chapter) (*Lecture, error) {

	for _, l := range chapter.Children {
		if l.Title == title {
			return l, nil
		}
	}

	return nil, fmt.Errorf("Unable to find lecture '%v'.", title)
}

func create_lecture(title string, chapter *Chapter) (*Lecture, error) {

	exists, err := lecture_exists(title, chapter)

	if err != nil {
		show_error(fmt.Sprintf("Error checking lecture existence: %v", err))
	}

	if exists {
		show_warning(fmt.Sprintf("Lecture '%s' already exists! Creation aborted.", title))

		var found_lecture *Lecture

		for _, l := range chapter.Children {
			if l.Title == title {
				found_lecture = l
			}
		}

		return found_lecture, nil
	}

	template, err := os.ReadFile(filepath.Join(ROOT_DIR, LECTURE_TEMPLATE_PATH))

	if err != nil {
		return nil, fmt.Errorf("Error reading template file: %w", err)
	}

	os.WriteFile(filepath.Join(chapter.Path, title+".tex"), template, 0755)

	// Modify created files, if needed

	placeholders := map[string]string{
		"%%title%%": title,
	}

	populate_latex_fields(filepath.Join(chapter.Path, title+".tex"), placeholders)


	new_lecture := &Lecture{
		Parent: chapter,
		Path:   filepath.Join(chapter.Path, title),
		Title:  title,
	}

	chapter.Children = append(chapter.Children, new_lecture)

	set_current_lecture(new_lecture)

	add_lecture_to_composite(new_lecture, filepath.Join(chapter.Path, "composite.tex"), filepath.Join(chapter.Path, title+".tex"))

	return new_lecture, nil
}

func add_lecture_to_composite(lec *Lecture, composite_path, lecture_path string) error {
	// Read the LaTeX file
	data, err := os.ReadFile(composite_path)
	if err != nil {
		return fmt.Errorf("Error reading latex composite: %v", err)
	}
	content := string(data)

	// Find the `% LECTURES` marker
	placeholder := "% LECTURES"
	index := strings.Index(content, placeholder)
	if index == -1 {
		return fmt.Errorf("placeholder '%s' not found in composite", placeholder)
	}

	// Build the `\input{}` line
	new_lecture := fmt.Sprintf("\\input{%s} %% %s", lecture_path, lec.Title)

	// Check if the lecture already exists
	if strings.Contains(content, new_lecture) {
		return fmt.Errorf("lecture '%s' already included", lecture_path)
	}

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Find the last included lecture in the `% LECTURES` section
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

	// Append the new lecture at the end of the `% LECTURES` section
	for i := insertIndex + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" || !strings.HasPrefix(strings.TrimSpace(lines[i]), `\input{`) {
			lines = append(lines[:i], append([]string{new_lecture}, lines[i:]...)...)
			break
		}
	}

	// Join the lines back into a single string
	updated_content := strings.Join(lines, "\n")

	// Write the updated content back to the file
	err = os.WriteFile(composite_path, []byte(updated_content), 0755)
	if err != nil {
		return fmt.Errorf("error writing updated mainfile: %v", err)
	}

	return nil
}

func remove_lecture_from_composite(lec_title, composite_path string) error {
	// Read the LaTeX file
	data, err := os.ReadFile(composite_path)
	if err != nil {
		return fmt.Errorf("error reading chapter composite: %v", err)
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
		if strings.HasPrefix(trimmed, `\input{`) && strings.Contains(trimmed, fmt.Sprintf("%% %s", lec_title)) {
			// Remove the line
			lines = append(lines[:i], lines[i+1:]...)
			removed = true
			break
		}
	}

	// If no chapter was removed, return an error
	if !removed {
		return fmt.Errorf("chapter '%s' not found in the mainfile", lec_title)
	}

	// Join the lines back into a single string
	updatedContent := strings.Join(lines, "\n")

	// Write the updated content back to the file
	err = os.WriteFile(composite_path, []byte(updatedContent), 0755)
	if err != nil {
		return fmt.Errorf("error writing updated mainfile: %v", err)
	}

	return nil
}

type Selection struct {
	Semester string
	Course   string
}

func create_lecture_with_form() error {

	if CURRENT_CHAPTER == nil {
		show_warning("No chapters exist. Try creating one!")
		return fmt.Errorf("Current chapter doesn't exist.")
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
						return errors.New("Lecture must have a name!")
					}
					return nil
				}),

			huh.NewNote().
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_CHAPTER.Title)

						for _, lecture := range CURRENT_CHAPTER.Children {
							t.Child("Lecture: " + lecture.Title)
						}

						display_title := "Untitled"

						if title != "" {
							display_title = title
						}

						t.Child(bold_style.Render("New Lecture: " + display_title))

						return tree_style.Render(t.String())
					}, &title,
				).Height(len(CURRENT_CHAPTER.Children)+5),

			huh.NewConfirm().
				TitleFunc(
					func() string {
						if title == "" {
							return fmt.Sprint("Create 'Untitled'?")
						}
						return fmt.Sprintf("Create '%s'?", title)
					}, &title).
				Value(&confirmed),
		).Title("Creating a new Lecture"),
	).WithTheme(huh.ThemeBase())

	form.Run()

	if confirmed {
		new_lecture, err := create_lecture(title, CURRENT_CHAPTER)

		if err != nil {
			return err
		}

		if new_lecture != nil {
			show_message(fmt.Sprintf("Created new lecture '%v' (or already exists)!", title))
		}
	}

	return nil
}

func open_lecture_with_form() error {

	if CURRENT_CHAPTER == nil {
		show_warning("No chapters exist. Try creating one!")
		return fmt.Errorf("Current chapter doesn't exist.")
	}

	var chosen_lecture string
	var lecture_names []string

	for _, l := range CURRENT_CHAPTER.Children {
		lecture_names = append(lecture_names, l.Title)
	}

	options := huh.NewOptions(lecture_names...)

	form := huh.NewForm(
		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Choose a Lecture").
				Options(options...).
				Value(&chosen_lecture),

			huh.NewNote().
				DescriptionFunc(
					func() string {

						t := tree.New().Root(CURRENT_CHAPTER.Title)

						for _, lecture := range CURRENT_CHAPTER.Children {
							if lecture.Title == chosen_lecture {
								t.Child(bold_style.Render("Lecture: " + lecture.Title))
							} else {
								t.Child("Lecture: " + lecture.Title)
							}
						}

						return tree_style.Render(t.String())
					}, &chosen_lecture,
				),
		).WithHeight(30),
	).WithTheme(huh.ThemeBase())

	form.Run()

	lec, err := find_lecture(chosen_lecture, CURRENT_CHAPTER)

	if err != nil {
		return fmt.Errorf("Could not find chosen lecture: %s", err.Error())
	}

	open_lecture(lec)

	return nil

}

func open_lecture(lec *Lecture) error {

	args := ":VimtexCompile"

	cmd := exec.Command("vim", "-c", args, lec.Path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("Failed to open tex file: %s", err.Error())
	}

	return nil
}

func remove_lecture_with_form() error {

	var choices []string
	confirmed := false
	double_confirmed := false

	var options []string

	for _, s := range CURRENT_CHAPTER.Children {
		options = append(options, s.Title)
	}

	form := huh.NewForm(

		huh.NewGroup(

			huh.NewMultiSelect[string]().
				Title("Choose Lecture(s) to delete").
				Options(huh.NewOptions(options...)...).
				Value(&choices).
				Validate(func(s []string) error {
					if len(choices) < 1 {
						return fmt.Errorf("You must choose lecture(s) to delete!")
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
			lec, err := find_lecture(choice, CURRENT_CHAPTER)

			if err != nil || lec == nil {
				return fmt.Errorf("Could not find lecture '%s'.", choice)
			}

			os.RemoveAll(lec.Path) // remove from filestructure

			CURRENT_CHAPTER.Children = remove_lecture(CURRENT_CHAPTER.Children, lec)

			remove_lecture_from_composite(lec.Title, filepath.Join(CURRENT_CHAPTER.Path, CHAPTER_COMPOSITE_PATH))
		}
	}

	return nil
}

func remove_lecture(lectures []*Lecture, toRemove *Lecture) []*Lecture {
	for i, lec := range lectures {
		if lec == toRemove {
			return append(lectures[:i], lectures[i+1:]...)
		}
	}
	return lectures	
}
