package main

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Logs - object that aid to collect needed logs.
type Logs struct {
	conf        *viper.Viper
	logsFiles   map[string][]string
	logsRegexp  map[string]string
	logsChannel map[string]string
}

// LogConfig - sturct for store needed log entry.
type LogConfig struct {
	Name    string
	Type    string
	Regexp  string
	Path    interface{}
	Channel interface{}
}

var entries []*LogConfig

func (l *Logs) collectLogs() {
	confMap := l.conf.AllSettings()
	entries := makeSlice(confMap)

	for _, st := range entries {
		// TODO: Treat cast errors.
		switch st.Type {
		case "file":
			l.logsFiles[st.Name] = []string{st.Path.(string)}
		case "dir":
			files, _ := filepath.Glob(filepath.Join(st.Path.(string), "/*.log"))
			l.logsFiles[st.Name] = files
		case "array":
			for _, f := range st.Path.([]interface{}) {
				l.logsFiles[st.Name] = append(l.logsFiles[st.Name], f.(string))
			}
		}
		l.logsRegexp[st.Name] = st.Regexp

		switch st.Channel.(type) {
		case string:
			l.logsChannel[st.Name] = st.Channel.(string)
		case bool:
			l.logsChannel[st.Name] = ""
		}
	}

	// FIXME: Need do it corss map.
	for _, slice := range l.logsFiles {
		removeDuplicates(&slice)
	}
}

func makeSlice(confMap map[string]interface{}) []*LogConfig {
	for _, entry := range confMap {
		// TODO: Treat cast errors.
		vv, ok := entry.(map[string]interface{})
		if !ok {
			Logger.Panic("Error")
		}

		for _, logMap := range vv {
			st := &LogConfig{}
			fillStruct(logMap.(map[string]interface{}), st)
			entries = append(entries, st)
		}
	}
	return entries
}

func fillStruct(data map[string]interface{}, result interface{}) {
	t := reflect.ValueOf(result).Elem()
	for k, v := range data {
		val := t.FieldByName(strings.Title(k))
		val.Set(reflect.ValueOf(v))
	}
}

func removeDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}
