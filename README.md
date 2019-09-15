# A must-be catchy header

First ssh login to new server? Miss your dot files? Want to be able to get them fast and manage them easily? Keep reading...

# Problem?

Dotman tries to solve a problem of managing personal dot files (or any kind of 'git-able' data meant to live in users homedirs for sake being).


The idea of storing dot files in a git repository is not a new one. The problem arises when destination machine doesn't have git installed, or you realise that you only want a subset of your configuration.

# Can I haz solution?

Dotman is small program which connects to given git repository, clones it and shares it over http, with bash friendly CLI. Packages of dotfiles are represented by folders in root of given git repository and presented as select list in CLI. Each file from selected package is downloaded relative to current user home directory.


# Less talk, more action!

Underneath demo has following file structure in connected git repository:
```
./vim
./vim/.vimrc
./tags.yaml
./BetterBash
./BetterBash/.inputrc
./BetterBash/.bshell
./BetterBash/.bshell/bb.sh
./mc
./mc/.config
./mc/.config/mc
./mc/.config/mc/ini
./mc/.config/mc/panels.ini
./README.md
./bashrc
./bashrc/.bashrc
./bash_aliases
./bash_aliases/.bash_aliases
```
![dotman demo](demo.gif)
