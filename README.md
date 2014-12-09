# Goose

Simple markdown wiki with Go backend. Written for personal use, so there is a strong dose of NIH in this project. You've been warned!

## Installation

Technically Goose has no runtime dependencies (at the moment it doesn't even support databases, just flat files, so you don't need a db server), but I can't be bothered to make a binary distribution or tarball. I do not package the compiled frontend assets or binary, so you will need a variety of tools to build or hack on goose:

- go 1.3+ with proper `GOPATH` setup and so on
- node.js 0.10.x
- npm 1.4.x
- git 1.8.x (otherwise you get caught by this [http redirect bug](https://github.com/spf13/hugo/issues/297) in old git versions)

Then you should be able to do this:

```bash
$ go get github.com/tummychow/goose
$ cd $GOPATH/src/github.com/tummychow/goose

$ npm i -g bower gulp # if you do not already have these
$ npm i
$ bower install

$ gulp go # build binary
$ gulp js # build javascript assets
$ gulp css # build css assets

$ gulp # run dev server
```

This launches the development server, which watches your JS, CSS and Go files (but you still have to run the tasks manually the first time because the compiled files don't exist yet). The server uses browsersync so you don't have to press F5 a hundred times to be productive. There are a few other gulp tasks implemented or in the works, but the only one you need to know for hacking is the default task.

At the moment, goose does not require any dependency management. I use gpm internally, but at the moment, goose's dependencies are all stable enough that I don't feel the need to implement anything more than `go get`. Future dependency management functions may be integrated into the gulpfile, or I might just use godep.

## Usage

Right now the application is very bare-bones, but it does actually do the basic jobs of a wiki (reading and writing pages). The route `/w/foo/bar` will take you to the page `/foo/bar`, while `/e/foo/bar` lets you edit or create that page. Rendering is done client-side in JS; commonmark compliance via [remarkable](https://github.com/jonschlinkert/remarkable) is on the roadmap but not really important atm.

## Configuration

Configuration is by environment variables. Most of them are set for you by gulp if you use the dev server, but if you run the binary straight from the command line then you will need to know these.

- `GOOSE_PORT` server port, eg `:4567` (note leading colon)
- `GOOSE_BACKEND` the backend URI, eg `file:///tmp/goose`
- `GOOSE_TEST_FILE` for testing the file backend, eg `file:///tmp/goose_test`. These tests will be skipped if this is unset or invalid. If you want `gulp test` to run them, make sure to set it beforehand, eg `GOOSE_TEST_FILE=file:///tmp/goose_test gulp test`.
- `GOOSE_DEV` to enable development-only behavior, eg template recompilation on every request

## License

BSD 3-clause, see [LICENSE.md](LICENSE.md)
