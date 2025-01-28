package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func apply_template(template_path string, node *Node) error {
	switch filepath.Ext(template_path) {
	case ".json":
		if err := os.MkdirAll(node.get_path(), os.ModePerm); err != nil {
			return err
		}
		template, err := parse_template(template_path)
		if err != nil {
			return err
		}
		if err := create_file_structure(template, node.get_path()); err != nil {
			return err
		}
		return write_info_json_values(node)
	case ".tex":
		return copy_file(template_path, node.get_path())
	}
	return nil
}

func parse_template(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}
	var template map[string]interface{}
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse JSON template: %w", err)
	}
	return template, nil
}

func write_info_json_values(node *Node) error {
	infoPath := filepath.Join(node.get_path(), "info.json")
	if err := write_json_value(infoPath, "title", node.get_title()); err != nil {
		return err
	}
	if err := write_json_value(infoPath, "group", node.get_group()); err != nil {
		return err
	}
	return write_json_value(infoPath, "id", node.get_id())
}

func populate_note_fields(node *Node) error {

	fields := get_struct_field_names(*node)

	placeholders := map[string]interface{}{}

	for _, field := range fields {
		value := node.get_field_value_by_name(field)

		if str, ok := value.(string); ok {
			tex_field := CFG_REPLACE_MARKER + strings.ToLower(field) + CFG_REPLACE_MARKER
			placeholders[tex_field] = str
		} else {
			if node, ok := value.(*Node); ok {
				if node != nil {
					tex_field := CFG_REPLACE_MARKER + strings.ToLower(node.get_group()) + CFG_REPLACE_MARKER
					placeholders[tex_field] = node.get_title()
				}
			}
		}
	}

	if filepath.Ext(node.get_path()) == CFG_NOTE_FILETYPE {
		data, err := os.ReadFile(node.get_path())
		if err != nil {
			return err
		}

		file_content := string(data)

		for placeholder, value := range placeholders {
			file_content = strings.ReplaceAll(file_content, placeholder, value.(string))
		}

		os.WriteFile(node.get_path(), []byte(file_content), os.ModePerm)

		return nil
	}

	files, err := os.ReadDir(node.get_path())
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == CFG_NOTE_FILETYPE {
			data, err := os.ReadFile(filepath.Join(node.get_path(), file.Name()))
			if err != nil {
				return err
			}

			file_content := string(data)

			for placeholder, value := range placeholders {
				file_content = strings.ReplaceAll(file_content, placeholder, value.(string))
			}

			os.WriteFile(filepath.Join(node.get_path(), file.Name()), []byte(file_content), os.ModePerm)
		}
	}

	return nil
}

func add_children_to_input_file(node *Node) error {
	// Get the "composite" file for the parent node
	parentFile, err := get_composite_file(node.get_path())
	if err != nil {
		return err
	}

	// Read the entire contents of the parent file
	parentData, err := os.ReadFile(parentFile)
	if err != nil {
		return fmt.Errorf("unable to read parent file %s: %w", parentFile, err)
	}
	parentContent := string(parentData)

	// We'll split once, keep it in memory, and rewrite at the end
	lines := strings.Split(parentContent, "\n")

	// Define the placeholder section where we insert lines
	placeholder := "% INPUT"

	for _, child := range node.get_children() {
		// Get the "composite" file for the child (the .tex file we want to \input)
		childFile, err := get_composite_file(child.get_path())
		if err != nil {
			return err
		}

		// Build the new line we want to insert, e.g.
		//   \input{.../Lec1.tex} % Lec1
		newLine := fmt.Sprintf("\\input{%s} %% %s",
			filepath.Join(child.get_path(), filepath.Base(childFile)),
			child.get_title())

		// 1) Check if the exact line is already present
		//    If it's there, skip adding it again.
		if strings.Contains(parentContent, newLine) {
			continue
		}

		// 2) If not already there, we place it after existing \input{} lines
		//    within the `% INPUT` section.
		insertIndex := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == placeholder {
				insertIndex = i
				break
			}
		}

		if insertIndex == -1 {
			// If we can't find the placeholder, bail out
			return fmt.Errorf("unable to locate placeholder '%s' in file %s", placeholder, parentFile)
		}

		// Insert after the “INPUT” placeholder lines or in the next blank line.
		// We'll scan forward for the next place to insert.
		didInsert := false
		for i := insertIndex + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])

			// We assume the “INPUT section” continues while lines have `\input{...}`
			// or are empty. Once we hit something else, insert right before it.
			if trimmed == "" || strings.HasPrefix(trimmed, `\input{`) {
				// Keep going
				continue
			} else {
				// Insert here
				lines = append(lines[:i], append([]string{newLine}, lines[i:]...)...)
				didInsert = true
				break
			}
		}

		// If we never hit a non-\input line, we can append at the end
		if !didInsert {
			lines = append(lines, newLine)
		}

		// Re-join lines to keep content up-to-date for subsequent children
		parentContent = strings.Join(lines, "\n")
	}

	// Finally, write the updated content back to the parent file
	err = os.WriteFile(parentFile, []byte(parentContent), 0755)
	if err != nil {
		return fmt.Errorf("error writing updated composite file %s: %v", parentFile, err)
	}

	return nil
}

func remove_from_parent_input_file(node *Node) error {
	parent := node.get_parent()
	if parent == nil {
		return nil
	}

	composite_file, err := get_composite_file(parent.get_path())
	if err != nil {
		return nil
	}

	if composite_file == "" {
		return nil
	}

	data, err := os.ReadFile(composite_file)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	filteredLines := make([]string, 0, len(lines)) // Create a new slice for filtered lines

	lineToRemove := fmt.Sprintf("%% %v", node.get_title()) // The line to match

	lineFound := false
	for _, line := range lines {
		if strings.Contains(line, lineToRemove) {
			lineFound = true
			continue // Skip this line, effectively removing it
		}
		filteredLines = append(filteredLines, line)
	}

	if !lineFound {
		return nil
	}

	// Join the filtered lines back into a single string
	updatedContent := strings.Join(filteredLines, "\n")

	// Open the file in write mode with truncation to ensure all old content is removed
	file, err := os.OpenFile(composite_file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	// Write the updated content back to the composite file
	_, err = file.WriteString(updatedContent)
	if err != nil {
		return fmt.Errorf("failed to write updated content to '%v': %w", composite_file, err)
	}

	return nil
}

func get_composite_file(path string) (string, error) {

	if filepath.Ext(path) == CFG_NOTE_FILETYPE {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		content := string(data)

		if strings.Contains(content, "% COMPOSITE") {
			return "", nil
		}
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != CFG_NOTE_FILETYPE {
			continue

		}

		data, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return "", err
		}
		content := string(data)

		if strings.Contains(content, "% COMPOSITE") {
			return filepath.Join(path, file.Name()), nil
		}
	}
	return "", fmt.Errorf("couldn't find composite file")
}
