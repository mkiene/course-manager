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

var SEMESTER_TEMPLATE_PATH = "data/templates/structure/semester.json"
var SEMESTERS_DIR = "data/semesters"

var COURSE_TEMPLATE_PATH = "data/templates/structure/course.json"
var COURSES_PATH = "courses"
var LECTURES_MASTER_PATH = "lectures/lec-master.tex"

var CHAPTER_TEMPLATE_PATH = "data/templates/structure/chapter.json"
var CHAPTERS_PATH = "lectures/chapters"
var CHAPTER_COMPOSITE_PATH = "composite.tex"

var LECTURE_TEMPLATE_PATH = "data/templates/files/lecture/lecture.tex"

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
		show_error(fmt.Sprintf("Error locating currents: %s.\nPlease set:", err.Error()))
		set_currents()
	}

	fmt.Println()

	handle_input()

	fmt.Println()
}

func handle_input() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "n", "new":
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "s", "sem", "semester":
					create_semester_with_form()
				case "co", "cou", "course":
					create_course_with_form()
				case "ch", "chap", "chapter":
					create_chapter_with_form()
				case "l", "lec", "lecture":
					create_lecture_with_form()
				}
			} else {
				choose_create_form()
			}
		case "cur", "current":
			set_currents()
		case "l", "lec", "lecture":
			if len(os.Args) > 2 {
				if os.Args[2] == "choose" {
					open_lecture_with_form()
				}
			} else {
				if CURRENT_LECTURE != nil {
					open_lecture(CURRENT_LECTURE)
				} else {
					show_warning("The current lecture is not set.")
				}
			}
		case "t", "tree":
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
		case "rm", "remove":
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "s", "sem", "semester":
					err := remove_semester_with_form()

					if err != nil {
						show_error(err.Error())
					}
				case "co", "cou", "course":
					err := remove_course_with_form()

					if err != nil {
						show_error(err.Error())
					}
				case "ch", "chap", "chapter":
					err := remove_chapter_with_form()

					if err != nil {
						show_error(err.Error())
					}
				case "l", "lec", "lecture":
					err := remove_lecture_with_form()

					if err != nil {
						show_error(err.Error())
					}
				}
			}
		}
	}
}
