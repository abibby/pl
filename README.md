# PL

`pl` is a simple tool to run a command multiple times in parallel.

## Installation

pl can be installed by installing go, instructions can be found
[here](https://golang.org/doc/install), then running `go get
github.com/abibby/pl`.

## Usage

To run a command in parallel with its self simply run `pl <count> <command>`
where `<count>` is the number of times you want it to run and `<command>` is the
command you want to run.

For example you could send 100 requests to a web server concurrently using `pl`
and `curl` with `pl 100 curl http://localhost`