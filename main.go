package main

import (
	"fmt"
	"net/http"
)

func main() {
	// cmd.Execute()
	http.HandleFunc("/", rootHandler)
	fmt.Println("Serving traffic on port 8080")
	http.ListenAndServe(":8080", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// items := []list.Item{}
	// compositions := helper.getAllCompositions()
	// for _, composition := range compositions.Items {
	// 	if expectedKind == composition.Spec.CompositeTypeRef.Kind &&
	// 		expectedApi == composition.Spec.CompositeTypeRef.ApiVersion {
	// 		items = append(items, item(composition.Metadata.Name))
	// 	}
	// }
	fmt.Fprintf(w, "Hello world!")
}
