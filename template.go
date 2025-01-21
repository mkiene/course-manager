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
		value := node.get(field)

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

	parent_file, err := get_composite_file(node.get_path())
	if err != nil {
		return err
	}

	for _, child := range node.get_children() {
		child_file, err := get_composite_file(child.get_path())
		if err != nil {
			return err
		}

		addition := fmt.Sprintf("\\input{%s} %% %s", filepath.Join(child.get_path(), filepath.Base(child_file)), child.get_title())

		data, err := os.ReadFile(parent_file)
		if err != nil {
			return err
		}
		content := string(data)

		// Check if the input already exists
		if strings.Contains(content, addition) {
			continue
		}

		// Split content into lines
		lines := strings.Split(content, "\n")

		placeholder := "% INPUT"

		// Find the last input in the `% INPUT` section
		insert_index := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == placeholder {
				insert_index = i
				break
			}
		}

		if insert_index == -1 {
			return fmt.Errorf("unable to locate placeholder '%s' in lines", placeholder)
		}

		// Append the new input at the end of the `% INPUT` section
		for i := insert_index + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" || !strings.HasPrefix(strings.TrimSpace(lines[i]), `\input{`) {
				lines = append(lines[:i], append([]string{addition}, lines[i:]...)...)
				break
			}
		}

		// Join the lines back into a single string
		updated_content := strings.Join(lines, "\n")

		// Write the updated content back to the file
		err = os.WriteFile(parent_file, []byte(updated_content), 0755)
		if err != nil {
			return fmt.Errorf("error writing updated mainfile: %v", err)
		}
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
