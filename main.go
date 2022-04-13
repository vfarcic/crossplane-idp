package main

import (
	"crossplane-idp/src/helper"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
)

func main() {
	// cmd.Execute()
	http.HandleFunc("/xrd", xrdHandler)
	http.HandleFunc("/composition", compositionHandler)
	http.HandleFunc("/xr", xrHandler)
	http.HandleFunc("/", xrdsHandler)
	fmt.Println("Serving traffic on port 8080")
	http.ListenAndServe(":8080", nil)
}

func xrdsHandler(w http.ResponseWriter, r *http.Request) {
	xrds := helper.GetXRDs()
	tmpl := template.Must(template.ParseFiles("src/html/xrds.tmpl"))
	tmpl.Execute(w, xrds)
}

func xrdHandler(w http.ResponseWriter, r *http.Request) {
	xrdName := r.URL.Query().Get("xrdName")
	xrd := helper.GetXRD(xrdName)
	tmpl := template.Must(template.ParseFiles("src/html/xrd.tmpl"))
	tmpl.Execute(w, xrd)
}

func compositionHandler(w http.ResponseWriter, r *http.Request) {
	compositionName := r.URL.Query().Get("compositionName")
	xrdName := r.URL.Query().Get("xrdName")
	yaml := helper.GetXRYamlWithFields(xrdName, compositionName)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, yaml)
}

func xrHandler(w http.ResponseWriter, r *http.Request) {
	compositionName := r.URL.Query().Get("compositionName")
	xrdName := r.URL.Query().Get("xrdName")
	fields := []string{}
	fieldsCount, _ := strconv.Atoi(r.URL.Query().Get("fieldsCount"))
	for i := 0; i < fieldsCount; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		fields = append(fields, r.URL.Query().Get(fieldName))
	}
	yaml := helper.GetXRYamlWithValues(xrdName, compositionName, fields)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, yaml)
}
