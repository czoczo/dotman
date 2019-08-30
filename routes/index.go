
// Main view with menu or secret prompt screen
// =================================================

package routes

import (
    "strings"
	"net/http"
    "text/template"
    "path/filepath"
)

type IndexData struct {
    BaseURL string
    Logo string
    ClientSecret string
}

func ServeMain(w http.ResponseWriter, r *http.Request, baseurl string, secret string, logo string) {
    // check for secret presence in HTTP header
    client_secret := r.Header.Get("secret")

    // build data for template
    data := IndexData{baseurl, strings.ReplaceAll(logo,"'","'\"'\"'"), client_secret}

    // set default template path
    tmplPath := "routes/views/passprompt.sh"

    // if correct secret, print menu (index)
    if client_secret == secret {
        tmplPath = "routes/views/index.sh"
    }

    // render template
    tmpl, err := template.New(filepath.Base(tmplPath)).ParseFiles(tmplPath)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}
