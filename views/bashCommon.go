
// Bash template commons
// =================================================

package views

var bashTemplHead = `
#!/bin/bash

barPrint() {
  echo -e "\n\e[97m-========================================================-\e[0m\n"
}

confirmPrompt() {
  echo -e  "\n\n  Proceed? [Y/n]"
  read -u 3 -n 1 -r -s
  if [[ $REPLY =~ ^[Nn]$ ]]
  then
  echo "  Aborted."
  exit 0
  fi
}
`
var gitCloneTmpl = `
gitCloneIfPresent() {
    command -v git >/dev/null 2>&1  || return
    echo -n "  GIT present. Downloading whole repository to ~/.dotman/dotfiles : "
    curl -s -H"secret:$1" {{.BaseURL}}/dotfilesrepo.tar.gz | tar -xz --no-same-owner -C ~/.dotman
    RESULT=$?; [ $RESULT -eq 0 ] && echo -e "\e[0;32mok\e[0m" || echo -e "\e[0;31merror\e[0m"
    URLMASK="{{.URLMask}}"
    [ -z "$URLMASK" ] && return
    GITCFGFILE="$HOME/.dotman/dotfiles/.git/config"
    GITCFG="$(cat ${GITCFGFILE})"
    rm "$GITCFGFILE"
    printf '%s\n' "$GITCFG" | while IFS= read -r line
    do
        (echo "$line" | grep -q "url" && echo "    url = {{.URLMask}}" || echo "$line") >> "$GITCFGFILE"
    done
}
`
