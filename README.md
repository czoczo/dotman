# A must-be catchy header

First ssh login to new server? Miss your dot files? Want to be able to get them fast and manage them almost effortlessly? Keep reading...

# Problem?

Dotman tries to solve a problem of managing personal dot files (or any kind of 'git-able' data meant to live in users homedir for sake being).


The idea of storing dot files in a git repository is not a new one. The problem arises when destination machine doesn't have git installed, or you realise that you only want a subset of your configuration. 

# Can I haz solution?

Dotman is small program which connects to given git repository, clones it and shares it over http, with bash friendly CLI. You group applications dotfiles by putting them in folders in root of git repo. Folder names become select options in dotmans CLI.
Dotman supports two dotfiles delivery methods:
 - files copy: each file from selected package is downloaded relative to current user home directory. Requires only bash and curl
 - git symlinks: if git is present, you can choose to download whole repository. Dotman will create all necessary folders and symlinks. This way you can easily push any changes in dotfiles using standard git.


# Less talk, more action!

Underneath demo has following file structure in connected git repository:
<center>
```
.
├── bash_aliases
│   └── .bash_aliases
├── bashrc
│   └── .bashrc
├── BetterBash
│   ├── .bshell
│   │   ├── bb.sh
│   │   └── git-prompt.sh
│   └── .inputrc
├── mc
│   └── .config
│       └── mc
│           ├── ini
│           └── panels.ini
├── README.md
├── tags.yaml
└── vim
    └── .vimrc
```
![dotman demo](demo.gif)
</center>

# Quick start
## Step 1
Create git repository on server of your choice with folders and dotconfigs inside, or use my example repository: http://cz0.cz/git/mydotfiles.

## Step 2
Either:
- download dotman binary from
- clone this repository and start with docker-compose up
- clone this repository and compile with `go build`

## Step 3
Setup minimal configuration by setting following environment variables according to your git repository access method:

### SSH
```
URL=ssh://git@github.com:username/dotfilesrepo.git
```

### HTTP/HTTPS
```
URL=https://username@github.com/username/dotfilesrepo.git
PASSWORD=repository_access_password
```

## Step 4
Now just run dotman and see the magic happen. On first run, dotman will generate SSH key pair and print public key on standard output. Allow it to access your repository in order to use ssh connection.

Make sure dotman loaded repository correctly, by viewing the logs. If so, you're ready to go!
```
curl 127.0.0.1:1338 | sh -
```

# Additional features
Except basic usage, dotman has some optional features, that will make your dotfiles deployment even more sexy.

## Want faster? - or, how tags work. 
Often you'll install same group of packages on different kind of machines. For instance, you'll probably want to install vim and bash configuration on any host you work on, but you wouldn't need mplayer/mpv at work (right?).

Dotman allows you to group packages and install them using single tag. To configure tags, put config.yaml file in root of your git repository containing dictionary. Each entry should have tagname as key, and list of folders in repository (packages) as value, like so:
```
tags:
  work:
    - bashrc
    - screen
    - vim
  home:
    - mpv
    - mplayer
    - bashrc
    - screen
    - vim
```

Install packagee with tags by using `http://myserver.net/t/tagname` endpoint. It'll skip menu and go straight to downloading files.

## Eliminate manual work - word or two about autorun
Every now and then, there is a need to run some commands after modyfing dotconfigs. Whether it's a window manager configuration reload, or evaluating some dynamic values for your config - doesn't matter! Just script it in bash, name it dotautorun.sh and put it inside in git folder holiding your applications dotfiles. It will be run during deployment of that package. Simple.

## Useful endpoints
Dotmans CLI is pretty elastic. Some endpoints can be useful outside CLI. You can call them using curl. If you use secret, pass it inside HTTP header like on examples below.

### Update installed dotfiles
Want to update all dotfiles managed by dotman on your workstation? Run following:
```
curl -H"secret:myterriblesecret" http://myserver.net/update
```

### Make server refresh served repository
You made changes to dotfiles in repo? Refresh files server on dotman by running:
```
curl -H"secret:myterriblesecret" http://myserver.net/sync
```

# Configuration in depth
All configuraion variables can be provided either as environment variables or as program arguments. The choice is yours. Here's a description of all of them:

| Environment variable | Argument | Type | Default Value | Description |
| ----- | ----- | ----- | ----- | ----- |
| URL | -url | string | - | URL to git repository containing dot files. Can be either http://, https:// or ssh:// protocol |
| BASEURL | -baseurl | string | http://127.0.0.1:1338" | URL prefix which will be used for generating download links. It should be the exact URL under which dotman is served. Use https if you put dotman behind SSL terminator |
| SSHKEY | -sshkey | string | ssh_data/id_rsa | Path to key used to connect git repository when using ssh protocol |
| SSHACCEPT | -sshaccept | boolean | - | Whether to add ssh remote servers key to known hosts file. Use it whenever you're binding dotman with a new repository over ssh |
| PASSWORD | -password | string | - | Password to use when connecting to git repository over HTTP protocol |
| PORT | -port | integer | 1338 | Port on which dotman should listen to. If you're going production, you're most likely to set port 80 |
| SECRET | -secret | string | - | If set, bash CLI will ask for secret and all dotfiles will be protected by it |
| URLMASK | -urlmask | string | - | If using containers, your URL variable might be different, than the one you would like to be set when using git & symlink install method. Use this variable to override URL in cloned repo |

# Security
 
 I know what some of you are thinking now: `http://whatever.net | sh -` pattern is a very ugly one. True. That's I want to make it clear: unless you're doing test on disposable virtual machine, or doing it on localhost, you're forbidden to use it without correctly configured TLS. For crying out load, it's your shell you're giving access to. You wouldn't let tandom guy put commands on your terminal, would you?
