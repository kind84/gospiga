package main

import (
	"fmt"
	"strings"
	"text/template"
)

func main() {
	var sb strings.Builder
	t := template.Must(template.New("update.tmpl").ParseFiles("../../db/dgraph/update.tmpl"))
	err := t.Execute(&sb, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(sb.String())
}
