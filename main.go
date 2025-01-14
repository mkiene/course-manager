package main

import (
	"fmt"
	"os"

	lg "github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

// Global Variables
var ROOT_DIR string
var CONFIG_DIR = "/Users/max/.config/cman/config.json"
var CURRENT_SEMESTER *Semester
var CURRENT_COURSE *Course
var CURRENT_CHAPTER *Chapter
var CURRENT_LECTURE *Lecture

// Lipgloss Styles
var new_style = lg.NewStyle().Bold(true).Foreground(lg.Color("#F4F3EE"))
var tree_style = lg.NewStyle().BorderStyle(lg.RoundedBorder()).Padding(0, 1, 0, 1)
var fatal_style = lg.NewStyle().Bold(true).Italic(true).Foreground(lg.Color("#960e2d"))
var error_style = lg.NewStyle().Italic(true).Foreground(lg.Color("#ff4466"))
var warning_style = lg.NewStyle().Italic(true).Foreground(lg.Color("#ffee66"))
var message_style = lg.NewStyle().Italic(true).Foreground(lg.Color("#9c66ff"))
var bold_style = lg.NewStyle().Bold(true).Foreground(lg.Color("#ffffff"))

func main() {

	var err error

	ROOT_DIR, err = get_json_value(CONFIG_DIR, "path")

	update_tree()

	if err != nil {
		show_error(fmt.Sprintf("Error updating tree: %s", err.Error()))
		return
	}

	err = get_currents()

	if err != nil {
		show_error(fmt.Sprintf("Error locating currents: %s", err.Error()))
	}

	handle_input()
}

func handle_input() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--set-current":
			set_currents()
		case "--create-semester":
			create_semester_with_form()
		case "--create-course":
			create_course_with_form()
		case "--create-lecture":
			create_lecture_with_form()
		case "--create-chapter":
			create_chapter_with_form()
		case "--open-lecture":
			open_lecture_with_form()
		case "--tree":
			var semesters_list []string

			for _, s := range Semesters {
				semesters_list = append(semesters_list, s.Title)
			}

			master_tree := tree.New().Root("Root")

			for _, s := range semesters_list {
				semester, err := find_semester(s)

				if err != nil {
					show_error(err.Error())
					continue
				}

				master_tree.Child(build_tree(semester))
			}

			fmt.Println(tree_style.Render(master_tree.String()))
		}
	}
}
