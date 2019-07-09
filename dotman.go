package main

import s "strings"
import (
	"fmt"
	"net/http"
    "log"
    "regexp"
)

func main() {
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Println( r.URL.Path)
        list, _ := regexp.Compile("^/([0-9]+,?)+$")
        if r.URL.Path == "/" {
		    fmt.Fprintf(w, "Lista")
            return
        }
        if list.MatchString(r.URL.Path) {
            slice := s.Split(s.Replace(r.URL.Path,"/","",-1), ",")
            var response string
            for _, num := range slice {
                response = response +" | "+ num
            }
		    fmt.Fprintf(w, response)
            return
        }
        fmt.Fprintf(w, "404 - Not Found")
	})

	http.ListenAndServe(":1337", nil)
}
