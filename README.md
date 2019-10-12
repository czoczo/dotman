# Introduction to Dotman

If you ever SSH'ed to a brand new server and wished you already had your dot files available, you just found the right tool.

While dotfiles were the original inspiration for this tool, Dotman is not limited to those. Any files (usually of configuration nature) suitable to be stored in a Git repository with intention of placing them in users home directory are suitable for use with the tool.

The idea of storing configuration files in a Git repository is not a new one. However, the problem arises when a destination machine does not have Git installed or you only require a subset of your configuration set.

# How it works

Dotman is a small program which connects to selected Git repository, clones it and shares it over http, with Bash friendly CLI. You group application dotfiles by putting them in folders in the root of a Git repository. Folder names become select options in Dotman's CLI.

Dotman supports two configuration file delivery methods:
 - File copies: each file from selected package is downloaded relative to current user's home directory. Requires only Bash and cURL.
 - Git symlinks: if git is present, you can choose to download the whole repository. Dotman will create all necessary folders and symlinks. This way you can easily push any changes to your configuration files using standard Git.

# Demo

In this demo we use Dotman to download Bash, MC, Screen and Vim configuration files using the file copy method. We then switch to the Git symlink method. Dotman server has the following file structure in the connected Git repository:

```
.
├── bashrc
│   ├── .bshell
│   │   ├── bb.sh
│   │   └── git-prompt.sh
│   ├── .bashrc
│   └── .inputrc
├── mc
│   └── .config
│       └── mc
│           ├── ini
│           └── panels.ini
├── mplayer
│   └── .mplayer
│       └── config
├── screen
│   └── .screenrc
├── vim
│   └── .vimrc
└── README.md
```

![dotman demo](demo.gif)

# Quick start

To get Dotman, you can use one of the following methods:
- download dotman binary from https://github.com/czoczo/dotman/releases
- clone this repository and compile with `go build`
- clone this repository and use `docker-compose build` to create a container

If running the binary, start with:
`./dotman -url https://github.com/czoczo/dotman-example-repo`

If using the container, start with:
`docker-compose up`

That's it! Check it by running:
```
curl localhost:1338 | sh -
```

# Setting own dotfiles repository
Create a Git repository on a server of your choice with preferred folder structure and your conguration files. You can use https://github.com/czoczo/dotman-example-repo as an example. Git repository might be set as public or private - your choice!

Most git servers offer at least two possible connection protocols: HTTP or SSH - both are supported by Dotman. Just remember to use proper prefix when setting URL like on the examples below.

## HTTP/HTTPS
Public
```
URL=https://github.com/username/dotfilesrepo.git
```

Private
```
URL=https://username@github.com/username/dotfilesrepo.git
PASSWORD=repository_access_password
```

## SSH
```
URL=ssh://git@github.com:username/dotfilesrepo.git
```

Now just run `dotman` and let the magic happen. When using ssh protocol, Dotman will generate SSH key pair and print the public key on standard output. Allow this key to access your repository in order to use the SSH connection.

If you're connecting to given repository for the first time, you're going to get this error: "error: ssh: handshake failed: knownhosts: key is unknown". Run `dotman` with `-sshaccept=true` once, to add remote key to known hosts.

Make sure dotman loaded the repository correctly by viewing the logs. If it has, you're ready to go!

# Extra features

## Grouping with tags

You will probably install the same group of packages on diffent machines. For instance, installing vim and bash configuration files go along nicely, but you may not want mplayer/mpv at work (right ;)?).

Dotman lets you group packages and install them at once using a tag. To configure tags, put config.yaml file in root of your Git repository. Each entry should have a tagname as key and list of folders in repository (packages) as value, like so:
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

Install packages with a tag by using `http://myserver.net/t/tagname` endpoint. It'll skip menu and go straight to downloading files.

## Eliminate manual work - word or two about autorun
Every now and then, there is a need to run some commands after modifying dotconfigs. Whether it's a window manager configuration reload or evaluating some dynamic values for your config! Just script it in bash, name it dotautorun.sh and put it inside git folder holding your applications dotfiles. It will be automatically executed during the deployment of that package.

## Useful endpoints
Dotman's CLI is pretty elastic. Some HTTP endpoints can be useful outside CLI. You can call them using curl. If you use secret, pass it inside HTTP header like in examples below.

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

## Keep up with repo - what is auto update?
One of dotman CLI menu options is enabling dotfiles auto updates. This simple function, adds curl request to `/update` endpoint to cron, so your dotfiles will be updated every hour. Disable by either deleting line in crontab, or using "disable auto update" option.

# Other solutions comparison
Dotman is not the first attempt of humanity to manage dotfiles. In fact there are dozens of such applications/frameworks. If you haven't yet seen website http://dotfiles.github.io/, I strongly advise you to check it out. Maybe some other app will suit your needs better.

But before you go, I want to point out some strengths of dotman:
* it doesn't require git on the host you're want to deploy dotfiles to
* you don't have to remember any commands/it has intuitive CLI
* almost effortless configuration
* server written in golang/portable binary
* capable of autoupdating dotfiles using cron 
* when installed, leverage standard git workflow to update dotfiles content

# Configuration in depth
All configuration variables can be provided either as environment variables or as program arguments. The choice is yours. Here's a description of all of them:

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
 
 I know what some of you are thinking now: `http://whatever.net | sh -` pattern looks ugly, especially to security guys. That's I want to make it clear: unless you're doing test on disposable virtual machine, or doing a test on localhost, you're forbidden to use it without correctly configured TLS infront of dotman. For crying out load, it's your shell you're giving access to. You wouldn't let random guy put commands on your terminal while you don't watch, would you?
