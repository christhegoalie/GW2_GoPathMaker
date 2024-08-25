package categories

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

const fileName = "_markerCategories.xml"
const fileExtension = ".cat"

type Category struct {
	Name        string
	DisplayName string
	keys        map[string]any
	Children    []Category
}

var validateFileExists func(fname string) string

func SetValidation(f func(fname string) string) {
	validateFileExists = f
}
func encodeCategory(c Category) string {
	txt := strings.Builder{}
	txt.WriteString(fmt.Sprintf(`<markercategory name="%s" displayname="%s"`, c.Name, c.DisplayName))
	for key, val := range c.keys {
		txt.WriteString(fmt.Sprintf(" %s=%s", key, val))
	}
	txt.WriteString(">")
	for _, c := range c.Children {
		txt.WriteString(encodeCategory(c))
	}
	txt.WriteString(`</markercategory>`)
	return txt.String()
}
func Save(categories []Category, path string) error {
	f, err := os.OpenFile(fmt.Sprintf(`%s/%s`, path, fileName), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(`<?xml version="1.0" encoding="utf-8"?><overlaydata>`)
	for _, c := range categories {
		f.WriteString(encodeCategory(c))
	}
	f.WriteString(`</overlaydata>`)
	return nil
}
func Compile(path string) ([]Category, []string, error) {
	out := []Category{}
	warns := []string{}
	items, _ := os.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			catName := filepath.Base(item.Name())
			newCats, newWarns, err := Compile(fmt.Sprintf("%s/%s", path, item.Name()))
			if err != nil {
				return out, warns, err
			}
			warns = append(warns, newWarns...)
			name, displayName := getNameInfo(catName)
			out = append(out, Category{Name: name, DisplayName: displayName, Children: newCats})
		} else if strings.HasSuffix(item.Name(), fileExtension) {
			newCat, newWarns, err := readCategory(fmt.Sprintf("%s/%s", path, item.Name()))
			if err != nil {
				return out, warns, err
			}
			warns = append(warns, newWarns...)
			out = append(out, newCat)
		}
	}
	return out, warns, nil
}

func readCategory(fileName string) (Category, []string, error) {
	catName, catDisplayName := getNameInfo(filepath.Base(fileName))

	cat := Category{Name: catName, DisplayName: catDisplayName, keys: make(map[string]any)}
	warns := []string{}

	b, err := os.ReadFile(fileName)
	if err != nil {
		return cat, warns, err
	}
	txt := strings.TrimSpace(string(b))
	if txt == "" {
		warns = append(warns, "No category definition found, consider switching to a directory")
	}

	lines := strings.Split(txt, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		line := strings.TrimSpace(line)
		ls := strings.Split(line, "=")
		if len(ls) != 2 {
			return cat, warns, fmt.Errorf("error in %s, Line: %d. Expected tuple key=value", fileName, i)
		}
		key := strings.TrimSpace(ls[0])
		val := strings.TrimSpace(ls[1])
		warn := validate(key, val)
		cat.keys[key] = val
		if len(warn) > 0 {
			warns = append(warns, fmt.Sprintf("Validation failed for %s, Warnning: %s", cat.DisplayName, warn))
		}
	}
	if _, ok := cat.keys["iconfile"]; !ok {
		warns = append(warns, fmt.Sprintf("No icon for: %s", cat.DisplayName))
	}

	return cat, warns, nil
}
func getNameInfo(pathName string) (string, string) {
	catName := strings.TrimSuffix(pathName, filepath.Ext(pathName))
	catDisplayName := strings.Builder{}
	for i, c := range catName {
		if i > 0 && unicode.IsUpper(c) {
			catDisplayName.WriteString(" ")
		}
		catDisplayName.WriteRune(c)
	}
	return catName, catDisplayName.String()
}

func validate(key, val string) string {
	var warn string
	if strings.EqualFold(key, "behavior") {
		warn = validateSet(val, []int{0, 2, 3, 4, 6, 7}) //1 and 5 are currently unsupported
	} else if strings.EqualFold(key, "iconsize") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "alpha") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "fadenear") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "fadefar") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "heightoffset") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "resetlength") {
		warn = validateNumeric(val)
	} else if strings.EqualFold(key, "iconfile") {
		if validateFileExists != nil {
			warn = validateFileExists(val)
		}
	}
	if warn != "" {
		return fmt.Sprintf("[%s] Warn: %s", key, warn)
	}
	return ""
}

func validateSet(v string, set []int) string {
	v = trim(v)
	iVal, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fmt.Sprintf("Expected integer, found %s", v)
	}
	if !slices.Contains(set, int(iVal)) {
		return fmt.Sprintf("Invalid value: %d, Expected value from: %+v", iVal, set)
	}
	return ""
}
func validateNumeric(v string) string {
	v = trim(v)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return "not numeric"
	} else if f < 0 {
		return "negative value"
	}
	return ""
}

func (c Category) MatchString(st string) bool {
	return c.MatchList(strings.Split(trim(st), "."))
}
func (c Category) MatchList(st []string) bool {
	if len(st) == 0 {
		return false
	}
	if len(st) == 1 {
		return st[0] == c.Name
	}
	if c.Name == st[0] {
		for _, child := range c.Children {
			if child.MatchList(st[1:]) {
				return true
			}
		}
	}

	return false
}

func trim(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), `"`), `"`)
}
