package main

const (
	CFG_CONFIG_DIR   = "/Users/max/.config/course-manager/config.json"
	CFG_SEMESTER_DIR = "data/semester"

	CFG_TEMPLATE_DIR = "data/templates"

	CFG_CURRENT_NODE_PREFIX = "current-"

	CFG_ROOT_FIELD = "root-dir"

	CFG_INFO_FILENAME = "info"

	CFG_REPLACE_MARKER = "%%"

	CFG_NOTE_FILETYPE = ".tex"

	CFG_EDITOR = "vim"
)

var CFG_NOTE_ARGUMENTS = []string{
		"-c",
		":VimtexCompile",
	}

var CFG_ALIASES = [][]string{
	{"new", "n"},
	{"remove", "rem", "rm"},
	{"semester", "sem", "s"},
	{"course", "cou", "co"},
	{"chapter", "chap", "ch"},
	{"lecture", "lec", "l"},
}

var CFG_GROUP_DEPTH = map[string]int{
	"semester": 0,
	"course":   1,
	"chapter":  2,
	"lecture":  3,
}

var CFG_DEPTH_GROUP = map[int]string{
	0: "semester",
	1: "course",
	2: "chapter",
	3: "lecture",
}

func get_config_value(field string) (string, error) {
	result, err := read_json_value(CFG_CONFIG_DIR, field)

	if err != nil {
		return "", err
	}

	return result, nil
}

func set_config_value(field, value string) error {
	err := write_json_value(CFG_CONFIG_DIR, field, value)

	if err != nil {
		return err
	}

	return nil
}

func get_alias_group(input string) string {
	for _, alias_group_outer := range CFG_ALIASES {
		for _, alias_group_inner := range alias_group_outer {
			if input == alias_group_inner {
				return alias_group_outer[0]
			}
		}
	}
	return ""
}
