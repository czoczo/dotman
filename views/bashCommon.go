
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
