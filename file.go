package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func create_file_structure(template map[string]interface{}, root string) error {
	for name, value := range template {
		current_path := filepath.Join(root, name)
		switch v := value.(type) {
		case string:
			if err := copy_file(v, current_path); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := os.MkdirAll(current_path, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", current_path, err)
			}
			if err := create_file_structure(v, current_path); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected value type for key %s", name)
		}
	}
	return nil
}

func copy_file(src, dst string) error {
	source_file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer source_file.Close()

	destination_file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destination_file.Close()

	if _, err := io.Copy(destination_file, source_file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func find_path(mode, root, title string) (string, error) {

	if title == filepath.Base(root) {
		return root, nil
	}

	var queue []string
	queue = append(queue, root)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:] // Dequeue

		entries, err := os.ReadDir(current)
		if err != nil {
			return "", err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(current, entry.Name())

			switch mode {
			case "f", "file":
				if !entry.IsDir() && strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) == title {
					return entryPath, nil
				}
			case "d", "dir", "directory":
				if entry.IsDir() && entry.Name() == title {
					return entryPath, nil
				}
			default:
				return "", fmt.Errorf("invalid mode")
			}

			// Add directories to the queue for BFS
			if entry.IsDir() {
				queue = append(queue, entryPath)
			}
		}
	}

	return "", fmt.Errorf("directory or file '%s' not found in '%s'", title, root)
}
