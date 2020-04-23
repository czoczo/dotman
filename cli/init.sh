#!/bin/bash

# stop if things go wrong
set -e

# declare variables for paths
DMAN_DIR="$HOME/.dotman"
RUNTIME_DIR="$DMAN_DIR/runtime"
SCRIPTS_DIR="$DMAN_DIR/scripts"

# create executable dirs
mkdir -p "$RUNTIME_DIR"
mkdir -p "$SCRIPTS_DIR"

# download dman CLI
# wget $URL cli.sh -O "$DMAN_DIR/runtime/dman

# create ~/.bash_profile if doesn't exist
[[ ! -f ~/.bash_profile ]] && touch ~/.bash_profile || echo "jesy"

# add loading of ~/.bashrc to .bash_profile
grep -q '.bashrc' ~/.bash_profile || echo '[[ -f ~/.bashrc ]] && . ~/.bashrc' >> ~/.bash_profile

# update $PATH with dotman executable dirs
PATH_LINE="[ -d \"$RUNTIME_DIR\" ] && [ -d \"$SCRIPTS_DIR\" ] && export PATH=\"$RUNTIME_DIR:$SCRIPTS_DIR:\$PATH\""
grep -q "$PATH_LINE" ~/.bash_profile || echo "$PATH_LINE" >> ~/.bash_profile

# display message about successfull install
echo "Restart terminal, or enter 'exec bash' to enable 'dman' command"
