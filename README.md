go_cwd_logger
=============

Rewrite of [cwd_logger](https://github.com/glennsb/cwd_logger) in Go

A simple MongoDB backed working directory logging system to provide quick
access to frequent or recent directories for a unix type system.

Usage
-----

The basic usage is to call ``cwd_logger`` when changing directory (chpwd_functions
in zsh is a good spot for this). Calling it as ``cwd_recently`` will print out the
20 most recent directories, runnnig it as ``cwd_frequency`` the 20 most frequent.

For either, passing a numeric argument will then print that directory. Any argument
other than a number will be treated as a regular expression for a search on all the
paths logged, returning the first match

```bash
user@host:/some/working/dir$ cwd_logger
user@host:/some/working/dir$ cd .
user@host:/some/working/dir$ cwd_recently
usage: cwd_recently index
 0 - /some/working/dir
user@host:/some/working/dir$ cd /some/other/dir && cwd_logger && cwd_frequency
usage: cwd_frequency index
 0 - /some/other/dir
 1 - /some/working/dir
user@host:/some/other/dir$ cd "$(cwd_frequency 1)"
user@host:/some/working/dir$ cwd_frequency
usage: cwd_frequence target_index|path regex
 0 - /some/working/dir
 1 - /some/other/dir
user@host:/some/working/dir$ cd "$(cwd_frequence other)"
user@host:/some/other/dir$
```

Easiest usage comes from using some shell functions & aliases to these commands,
see [cd_functions.zsh](cd_functions.zsh) for examples for zsh

Installation/Build/Setup
------------------------

#### Dependencies

* [MongoDB](http://www.mongodb.org)
* [Go](http://golang.org/)
* [mgo](http://labix.org/mgo) driver (which has some dependencies of its own)

#### Install

The standard `go get`, `go build` type of setup

    go get github.com/glennsb/go_cwd_logger
    go build github.com/glennsb/go_cwd_logger

You should end up with a `$GOPATH/bin/go_cwd_logger`, put that in your `$PATH`
as well as symlinks to it named as `cwd_recently` & `cwd_frequency`

By default it will connect to a MongoDB server as `mongodb://localhost/USERNAME`

You can override this by setting a `CWD_LOGGER_URI` environment (or editing the source)

It will use a collection named `logged_dirs` in the given database

### Why

I liked the old [cwd_logger](https://github.com/glennsb/cwd_logger), but was tired of Ruby's gem
dependency issues. I also wanted to play with Go more this morning

### License

`go_cwd_logger` is distributed under the MIT license & is copyright © 2013 Stuart Glenn.

See [LICENSE](LICENSE) for full MIT license
