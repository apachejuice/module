package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type guardStyle struct {
	Suffix        string `json:"suffix"`
	Prefix        string `json:"prefix"`
	SaveExtension bool   `json:"saveExtension"`
}

type moduleRules struct {
	Copyright    []string   `json:"copyright"`
	GuardStyle   guardStyle `json:"guardStyle"`
	EndifComment string     `json:"endifComment"`
}

func transformFilename(gs guardStyle, filename string) string {
	filename = filepath.Base(filename)

	if !gs.SaveExtension {
		filename = strings.Replace(filename, ".h", "", -1)
	}

	filename = strings.Replace(filename, ".", "_", -1)
	return gs.Prefix + strings.Replace(filename, "-", "_", -1) + gs.Suffix
}

func moduleHeaderNew(rules moduleRules, filename string) string {
	module := strings.Clone(strings.Join(rules.Copyright, "\n")) + "\n#ifndef "
	guard := transformFilename(rules.GuardStyle, filename)

	module += guard + "\n"
	return module + "#define " + guard + "\n\n#endif " + strings.Replace(rules.EndifComment, "$(GUARD)", guard, -1)
}

func moduleImplNew(rules moduleRules, header string) string {
	return strings.Join(rules.Copyright, "\n") + "\n#include \"" + filepath.Base(header) + "\"\n"
}

func getConfig() (moduleRules, error) {
	data, err := os.ReadFile("module-rules.json")
	if err != nil {
		return moduleRules{}, err
	}

	var rules moduleRules
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return moduleRules{}, err
	}

	return rules, nil
}

func handleErr(err error) {
	fmt.Printf("error: %s\n", err.Error())
	os.Exit(2)
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Printf("usage: %s [action] [modulepath]\n", os.Args[0])
		fmt.Println("action: (new) creates a new module")
		fmt.Println("modulepath: the path to create the module in")
		os.Exit(2)
	}

	dir := path.Dir(args[1])
	fm := fs.FileMode(0644)
	if args[0] != "new" {
		fmt.Printf("Invalid action: '%s'\n", args[0])
		os.Exit(2)
	}

	info, err := os.Stat(dir)
	if err != nil {
		handleErr(err)
	}

	if !info.IsDir() {
		handleErr(fmt.Errorf("'%s' is not a directory", dir))
	}

	rules, err := getConfig()
	if err != nil {
		handleErr(err)
	}

	hdr := moduleHeaderNew(rules, args[1]+".h")
	err = os.WriteFile(args[1]+".h", []byte(hdr), fm)
	if err != nil {
		handleErr(err)
	}

	cfile := moduleImplNew(rules, args[1]+".h")
	err = os.WriteFile(args[1]+".c", []byte(cfile), fm)
	if err != nil {
		handleErr(err)
	}
}
