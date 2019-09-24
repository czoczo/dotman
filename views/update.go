
// Update view - updates already installed dotfiles
// =================================================

package views

import (
	"net/http"
    "text/template"
)

type UpdateData struct {
    ClientSecret string
    RepoOpts string
    BaseURL string
    URLMask string
}

func ServeUpdate(w http.ResponseWriter, r *http.Request, baseurl string, client_secret string, directory string, foldersMap map[string]string, urlMask string) {

    // generate body of bash case with repo packages
    repoPackages := repoPackagesCasePrint(foldersMap, true, directory, baseurl)

    // build data for template
    data := UpdateData{client_secret, repoPackages, baseurl, urlMask}

    // render template
    tmpl, err := template.New("update").Parse(tmplUpdate)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}

var tmplUpdate = gitCloneTmpl + `
SECRET="{{ .ClientSecret }}"
selectPackage() {
    case "$1" in
    {{ .RepoOpts }}
    esac
}
if [ ! -f "$HOME/.dotman/managed" ]; then 
echo "  It appears, you don't manage any dotfiles using dotman. Exiting."
exit 1
fi

if [ -d "$HOME/.dotman/dotfiles" ]; then 
    gitCloneIfPresent "$SECRET"
fi

for NAME in $(cat "$HOME/.dotman/managed"); do
selectPackage $NAME 
done
`
