package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// create_file_structure takes a template structure, where keys are either
// directories (map[string]interface{}) or source-file paths (string),
// and recursively creates the corresponding directory tree and copies files.
func create_file_structure(template map[string]interface{}, root string) error {
	for name, value := range template {
		currentPath := filepath.Join(root, name)

		switch v := value.(type) {
		case string:
			// Before copying, check that we don't overwrite an existing file
			if _, err := os.Stat(currentPath); err == nil {
				return fmt.Errorf("file %s already exists", currentPath)
			} else if !errors.Is(err, os.ErrNotExist) {
				// If we get some other error, return it
				return fmt.Errorf("error checking existence of %s: %w", currentPath, err)
			}

			if err := copy_file(v, currentPath); err != nil {
				return err
			}

		case map[string]interface{}:
			// Create the subdirectory if it doesn't exist
			if err := os.MkdirAll(currentPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", currentPath, err)
			}
			// Recursively create the structure within that subdirectory
			if err := create_file_structure(v, currentPath); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unexpected value type for key %s", name)
		}
	}
	return nil
}

// copy_file copies the contents of src to dst. If dst exists, an error is returned.
func copy_file(src, dst string) error {
	// Double-check existence check here as well for safety:
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination file already exists: %s", dst)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error checking existence of %s: %w", dst, err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

// find_path searches a directory tree for a file or directory named 'title'.
// mode can be "f"/"file" or "d"/"dir"/"directory".
func find_path(mode, root, title string) (string, error) {
	if title == filepath.Base(root) {
		return root, nil
	}

	var queue []string
	queue = append(queue, root)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

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

			if entry.IsDir() {
				queue = append(queue, entryPath)
			}
		}
	}

	return "", fmt.Errorf("directory or file '%s' not found in '%s'", title, root)
}

// open_note spawns an external editor to open a note (Node).
func open_note(node *Node) error {
	CFG_NOTE_ARGUMENTS = append(CFG_NOTE_ARGUMENTS, node.get_path())

	cmd := exec.Command(CFG_EDITOR, CFG_NOTE_ARGUMENTS...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to open tex file: %s", err.Error())
	}

	return nil
}
