#!/bin/bash

MAN_DIR="$HOME/.dotman"
DOTFILES_DIR="$MAN_DIR/dotfiles"

#set -e

# UI colors
CR="$(tput sgr0)"    # Reset color
CH1="$(tput setaf 5)"  # Highlight color 1
CH2="$(tput setaf 2)"  # Highlight color 2
CW="$(tput setaf 3)"   # Warning color
CA="$(tput setaf 1)"   # Alert color
CB="$(tput bold)$(tput setaf 7)"    # Bold

#-------------------
# print help section
#-------------------

function CLI_help {
cat << EOF
$0 manages your personal configuration files.

Basic sections
  app           list, create, delete and install configuration groups (apps)
  file          start or stop managing selected configuration files
  alias         list, add, delete, and install by aliases (set of apps)
  update        run local apps update or manage auto updates
  *             any other command will be passed to 'git' command run relative to dotfiles repository directory

 Usage:
   $0 <section> <subcommand>

For subcommand usage run '$0 <section>'
EOF
}

function CLI_app_help {
cat <<EOF
App section commands
  list [app_name]                       list managed apps, or files grouped by app if app name given
  create <app_name>                     create folder for new app
  delete <app_name>                     delete app 
  install [app_name1] [app_name2]...    install chosen apps (shows interactive menu if no app names provided)
EOF
}

function CLI_app_create_help {
cat <<EOF
Creates new app folder.
Usage: $0 app create <app_name>
EOF
}

function CLI_app_delete_help {
cat <<EOF
Deletes app folder.
Usage: $0 app delete <app_name>
EOF
}

function CLI_file_help {
cat <<EOF
File section commands
  manage <app_name> <file>              add file to selected app in repository
  unmanage <app_name> <file>            delete file from app in repository, but leave the file at orginal location
  remove  <app_name> <file/path>        delete file from app in repoistory and from orginal location
EOF
}

function CLI_file_manage_help {
cat <<EOF
Adds file or directory to git repository.
Usage: $0 app manage <app> <file/directory>
EOF
}

function CLI_file_unmanage_help {
cat <<EOF
Removes file or directory from git repository.
Usage: $0 app manage <app> <file/directory>
EOF
}

function CLI_file_remove_help {
cat <<EOF
Removes file or directory from git repository and from orginal location.
Usage: $0 app manage <app> <file/directory>
EOF
}

#
#     alias create
#           list
#           delete
#           add
#           install
# 
#     update now
#            auto enable
#                 disable

#-----------------
# common functions
#-----------------

# print message
function msg {
    echo "$1"
}

# display confirm prompt
confirmPrompt() {
    echo -e  "\n\n  Proceed? [Y/n]"
    read -u 3 -n 1 -r -s
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "  Aborted."
        exit 0
    fi
}

# functions needed to run menu
function Get_odx {
    AWK=gawk
    [ -x /bin/nawk ] && AWK=nawk
    od -t o1 | $AWK '{ for (i=2; i<=NF; i++)
                        printf("%s%s", i==2 ? "" : " ", $i)
                        exit }'
}

# Grab terminal capabilities
tty_cuu1=$(tput cuu1 2>&1 | Get_odx)            # up arrow
tty_kcuu1=$(tput kcuu1 2>&1 | Get_odx)
tty_cud1=$(tput cud1 2>&1 | Get_odx)            # down arrow
tty_kcud1=$(tput kcud1 2>&1 | Get_odx)
tty_cub1=$(tput cub1 2>&1 | Get_odx)            # left arrow
tty_kcub1=$(tput kcud1 2>&1 | Get_odx)
tty_cuf1=$(tput cuf1 2>&1 | Get_odx)            # right arrow
tty_kcuf1=$(tput kcud1 2>&1 | Get_odx)
tty_ent=$(echo | Get_odx)                      # Enter key
tty_kent=$(tput kent 2>&1 | Get_odx)
tty_sc=$(echo -n " " | Get_odx)                 # Space key

# Some terminals (e.g. PuTTY) send the wrong code for certain arrow keys
if [ "$tty_cuu1" = "033 133 101" -o "$tty_kcuu1" = "033 133 101" ]; then
    tty_cudx="033 133 102"
    tty_cufx="033 133 103"
    tty_cubx="033 133 104"
fi


function print_menu {
#    echo "ADDR: $1"
    buffer=""
    for i in "${!LIST[@]}"; do
        selected=" "
        [ ${SELECTION[$i]} -eq 1 ] && selected="x"
        [ $(expr $i % $menu_columns) == 0 ] && buffer="$buffer\n"
        if [ $i -eq $1 ]; then
            buffer=$(printf "%s  [\e[32m%s\e[0m]\e[31m %-15s\e[0m" "$buffer" "$selected" "${LIST[$i]}")
        else
            buffer=$(printf "%s  [\e[32m%s\e[0m]\e[35m %-15s\e[0m" "$buffer" "$selected" "${LIST[$i]}")
        fi
    done
    tput clear
    echo -ne "$buffer"
}

function move {
    temp="$addr_x"
    addr_x=$((addr_x+${1}))
    [ $addr_x -gt 9 ] && addr_x="$temp"
    [ $addr_x -lt 0 ] && addr_x="$temp"
}

function select_item {
    SELECTION[$1]=$(( ${SELECTION[$1]} * -1 ))
}

function print_selection {
    echo -ne "\nWybrano elementy: "
    for i in "${!LIST[@]}"; do
        [ ${SELECTION[$i]} -eq 1 ] && echo -n "${LIST[$i]} "
    done
}



#-----------------------------
# second level command section
#-----------------------------

### app section

# list all defined apps, together with git last log record
function CLI_app_list {
    # if single argument
    if [ "$#" -eq 1 ]; then
        # print table column headers
        printf "%-15s $CH1%-7s $CH2%-10s $CR%s\n" "app name" "commit" "last modified" "description"
        echo -e "$CB==================================================$CR"
    
        # list all folders in dotfiles repository, leave only directory names, loop over the list
        for FILE in $(ls -l ${DOTFILES_DIR} | grep '^d' | rev | cut -d' ' -f1 | rev); do
            GIT_DATA="$(git -C "$DOTFILES_DIR" log --color -n 1 --pretty="format:%C(magenta)%h %C(green)(%cr) %C(reset)%s" "$FILE")"
            [ -z "$GIT_DATA" ] && GIT_DATA="${CW}not commited$CR"
            printf "%-15s %s\n" "$FILE" "$GIT_DATA" 
        done
    fi

    # if two arguments
    if [ "$#" -eq 2 ]; then
        APP_DIR="${DOTFILES_DIR}/$2"
        if [ ! -d "${DOTFILES_DIR}/$2" ]; then
            echo "There is no app named '$2'. Are you sure you provided a valid app name?"
            exit 1
        fi
        find "${DOTFILES_DIR}/$2"
    fi
}

function CLI_app_create {
    shift
    # print help if no single argument
    if [ "$#" -ne 1 ]; then
        CLI_app_create_help
        exit 0
    fi

    # stop if folder already exists
    if [ -e "$DOTFILES_DIR/$1" ]; then
        msg "App "$1" already exists. Choose another name."
        exit 0
    fi

    # create folder in dotfiles repository
    mkdir -p "$DOTFILES_DIR/$1"
    msg "App "$1" created. Add configuration files for dotman to manage."
}

function CLI_app_delete {
    shift
    # print help if no single argument
    if [ "$#" -ne 1 ]; then
        CLI_app_delete_help
        exit 0
    fi

    # stop if folder does not exist
    if [ ! -e "$DOTFILES_DIR/$1" ]; then
        msg "App "$1" does not exist. Choose another name."
        exit 0
    fi

    # create folder in dotfiles repository
    rm -rf "$DOTFILES_DIR/$1"
    msg "App "$1" deleted successfully. To make changes global push them to the server."
}

function CLI_app_install {
    LIST=( $(ls -l ${DOTFILES_DIR} | grep '^d' | rev | cut -d' ' -f1 | rev) )
    SELECTION=()
    for i in "${!LIST[@]}"; do
        SELECTION[$i]=-1
    done
    tty_save=$(stty -g)
    stty cs8 -icanon -echo 
    stty intr '' susp ''
    trap "stty $tty_save; exit"  INT HUP TERM
    
    menu_columns=3
    addr_x=0
    print_menu "$addr_x"
    
    while :; do
        keypress=$(dd bs=10 count=1 2> /dev/null | Get_odx)
    
        case "$keypress" in
            "$tty_ent"|"$tty_kent") print_selection; break;;
            "$tty_sc") select_item "$addr_x"; print_menu "$addr_x";;
            "$tty_cuu1"|"$tty_kcuu1") move "-3"; print_menu "$addr_x";;
            "$tty_cud1"|"$tty_kcud1"|"$tty_cudx") move "3"; print_menu "$addr_x";;
            "$tty_cub1"|"$tty_kcub1"|"$tty_cubx") move "-1"; print_menu "$addr_x";;
            "$tty_cuf1"|"$tty_kcuf1"|"$tty_cufx") move "1"; print_menu "$addr_x";;
            *) echo;;
        esac
    done
    
    # set orginal terminal setting
    stty $tty_save
}

### file section

function manage_file {
    # get absolute path
    SRC_PATH="$(realpath "$2")"

    # if $HOME in path, change to '~'
    DST_PATH_DIR="$(dirname "${SRC_PATH}")"
    if [[ "$DST_PATH_DIR" == $HOME* ]]; then
        DST_PATH_DIR="${DST_PATH_DIR/$HOME//\~}"
    fi

    # get name of the file
    FILENAME="$(basename "${SRC_PATH}")"

    # get destination folder in dotfiles repository
    DST_PATH_DIR="$DOTFILES_DIR/$1$DST_PATH_DIR"

    mkdir -p "$DST_PATH_DIR"
    mv "$SRC_PATH" "$DST_PATH_DIR/$FILENAME"
    ln -s "$DST_PATH_DIR/$FILENAME" "$SRC_PATH"

    msg "  $SRC_PATH"
}

function CLI_file_manage {
    shift
    # print help if not two argument
    if [ "$#" -ne 2 ]; then
        CLI_file_manage_help
        exit 0
    fi

    # stop if folder doesn not exist
    if [ ! -e "$DOTFILES_DIR/$1" ]; then
        msg "App "$1" does not exist. Choose another name."
        exit 0
    fi

    # stop if selected file or direcotry does not exist
    if [ ! -e "$2" ]; then
        msg "There is no file or directory under '$2' path. Make sure to provide correct path."
        exit 0
    fi

    # if adding a file
    if [ -f "$2" ]; then
        msg "Adding follwing file to repository as part of $1 app."
        manage_file "$@"
        exit 0
    fi

    # if adding a directory
    if [ -d "$2" ]; then
        msg "Adding following files to repository as part of $1 app:"
        find "$2" -type f | while read FILE; do manage_file "$1" "$FILE"; done
        exit 0
    fi

    msg "Error: wrong file type."
    exit 1
}

function unmanage_file {
    if [ ! -L "$2" ]; then
        msg "  File '$2' skipped - not managed by dotman"
        return 0
    fi
    # get absolute path
    #SRC_PATH="$(realpath "$2")"
    SRC_PATH="$2"

#echo "2: $2 || SRC_pATH: $SRC_PATH"
    # if $HOME in path, change to '~'
    DST_PATH_DIR="$(dirname "${SRC_PATH}")"
    if [[ "$DST_PATH_DIR" == $HOME* ]]; then
        DST_PATH_DIR="${DST_PATH_DIR/$HOME//\~}"
    fi

    # get name of the file
    FILENAME="$(basename "${SRC_PATH}")"

#echo "DST: $DST_PATH_DIR   || $FILENAME"
    # get destination folder in dotfiles repository
    DST_PATH_DIR="$DOTFILES_DIR/$1$DST_PATH_DIR"

#    echo "rm -f \"$SRC_PATH\""
#    echo "mv \"$DST_PATH_DIR/$FILENAME\" \"$SRC_PATH\"" 
    rm -f "$SRC_PATH"
    mv "$DST_PATH_DIR/$FILENAME" "$SRC_PATH" 
    
    find "$DST_PATH_DIR" -type d -empty -delete

    msg "  $SRC_PATH"
}

function CLI_file_unmanage {
    shift
#    echo -e "1: $1\n2: $2\n3: $3"
    # print help if not two arguments
    if [ "$#" -ne 3 ]; then
        CLI_file_unmanage_help
        exit 0
    fi

    # stop if folder doesn not exist
    if [ ! -e "$DOTFILES_DIR/$1" ]; then
        msg "App "$1" does not exist. Choose another name."
        exit 0
    fi

    # stop if selected file is not a symlink
    if [ ! -L "$2" ] && [ ! -d "$2" ]; then
        msg "There is no symlink under '$2' path. Make sure to provide path to file or directory managed by dotman."
        exit 0
    fi

    # unmanage if a symlink
    if [ -L "$2" ]; then
        msg "Removing follwing file to repository as part of $1 app."
        unmanage_file "$@"
        [ "$3" == "remove" ] && rm -f "$2" 
        exit 0
    fi

    # if adding a directory
    if [ -d "$2" ]; then
        msg "Removing following files to repository as part of $1 app:"
        find "$2" -type l | while read FILE; do unmanage_file "$1" "$FILE"; done
        [ "$3" == "remove" ] && rm -rf "$2" 
        exit 0
    fi

    msg "Error: wrong file type."
    exit 1
}



#--------------------------
# top level command section
#--------------------------

function CLI_app {
    shift
    case "$1" in
        list) CLI_app_list "$@";;
        create) CLI_app_create "$@";;
        delete) CLI_app_delete "$@";;
        install) CLI_app_install "$@";;
        *) CLI_app_help;;
    esac
}

function CLI_file {
    shift
    case "$1" in
        manage) CLI_file_manage "$@";;
        unmanage) CLI_file_unmanage "$1" "$2" "$3" "unmanage";;
        remove) CLI_file_unmanage "$1" "$2" "$3" "remove";;
        *) CLI_file_help;;
    esac
}

function CLI_alias {
    shift
    case "$1" in
        list) CLI__list;;
        *) CLI__help;;
    esac
}

function CLI_update {
    shift
    case "$1" in
        list) CLI__list;;
        *) CLI__help;;
    esac
}

#--------------
# program start
#--------------

# print usage on zero arguments
if [ "$#" -le 0 ]; then
    CLI_help
    exit 0
fi

# parse command
case "$1" in
    app) CLI_app "$@";;
    file) CLI_file "$@";;
    alias) CLI_alias "$@";;
    update) CLI_update "$@";;

    # pass all else to git relative to repo dir
    *) git -C "$DOTFILES_DIR" "$@";;
esac
