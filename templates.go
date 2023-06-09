package note

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gosimple/slug"
)

var DefaultTmpl = `+++ # {{ .TitleSlug }}.md
title = "{{ .Title }}"
created_at = "{{ dateFormat "2006-01-02T15:04:05Z07:00" .Date }}"
modified_at = "{{ dateFormat "2006-01-02T15:04:05Z07:00" .Date }}"
+++

{{ .Content }}
`

const DefaultFilenameTmpl = "{{ .TitleSlug }}.md"

func DateFormat(layout string, t time.Time) string {
	return t.Format(layout)
}

func MakeSlug(in string) string {
	return slug.Make(in)
}

var Tfuncs template.FuncMap = map[string]interface{}{
	"dateFormat": DateFormat,
	"makeSlug":   MakeSlug,
}

// FilenameFromFile takes a file, extracts and formats a filename-template and
// returns it as a string.
func FilenameFromFile(fname string, noteData Data) (string, error) {
	firstLine, err := GetFirstLineFromTemplateFile(fname)
	if err != nil {
		return "", err
	}

	filenameTemplate, err := ExtractTemplateFromLine(firstLine)
	if err != nil && err != ErrNoFilenameTemplate {
		return "", err
	}

	filename, err := FilenameFromTemplateStr(filenameTemplate, noteData)
	if err != nil {
		log.Debugf("in FilenameFromFile, retrurned empty: %s", err)
		return "", err
	}

	return filename, nil
}

// GetFirstLineFromTemplateFile opens the named file and returns the very first line.
func GetFirstLineFromTemplateFile(fname string) (string, error) {
	_, err := os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", err
	}

	file, err := os.Open(fname)
	if err != nil {
		return "", err
	}

	defer file.Close()

	r := bufio.NewReader(file)

	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimSuffix(line, "\n")), nil
}

// ExtractTemplateFromLine takes a raw filename-template string, and returns a
// string of only a valid go template.
func ExtractTemplateFromLine(line string) (string, error) {
	sString := strings.Split(line, "#")
	if len(sString) < 2 {
		return "", ErrNoFilenameTemplate
	}
	return strings.TrimSpace(sString[len(sString)-1]), nil
}

// FilenameFromTemplateStr takes a filename-template string and note data, then
// returns a string of the data applied to the template.
func FilenameFromTemplateStr(filenameTemplate string, noteData Data) (string, error) {
	var tmpl *template.Template

	if filenameTemplate == "" {
		filenameTemplate = DefaultFilenameTmpl
	}

	tmpl, err := template.New("filename").Funcs(Tfuncs).Parse(filenameTemplate)
	if err != nil {
		return "", err
	}

	w := bytes.Buffer{}
	err = tmpl.Execute(&w, noteData)
	if err != nil {
		return "", err
	}

	return w.String(), nil

}
