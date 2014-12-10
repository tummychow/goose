# Goose

Simple markdown wiki with Go backend. Written for personal use, so there is a strong dose of NIH in this project. You've been warned!

## Installation

Technically Goose has no runtime dependencies, but I can't be bothered to make a tarball for it. I do not package the compiled frontend assets or binary, so you will need a variety of tools to build or hack on Goose:

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

At the moment, Goose does not require any dependency management. I use gpm internally, but Goose's dependencies are all stable enough that I don't feel the need to implement anything more than `go get`. Future dependency management functions may be integrated into the gulpfile, or I might just use godep.

### Using postgres

By default, the dev server uses a flat file tree rooted at `/tmp/goose`, but Goose 0.2.0+ supports a postgresql database as its backend. At the moment, the target database needs only one table, shown below. Goose does not create the table for you.

```sql
CREATE TABLE documents (
    name TEXT NOT NULL,
    content TEXT NOT NULL,
    stamp TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (name, stamp)
);
```

Once you have the database ready, you can connect to it with the appropriate `GOOSE_BACKEND`:

```bash
$ export GOOSE_BACKEND=postgres://user:password@:5432/yourdb?sslmode=disable
$ gulp # the dev server inherits its backend from your environment
```

Postgres connection URIs are documented [here](http://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING). [lib/pq](https://github.com/lib/pq) supports most of the options. Note that lib/pq sets `sslmode=require` by default. If you are using an insecure connection (eg a Docker container), be sure to add the query option `sslmode=disable`, as shown above.

## Tests

Testing is a bit lightweight right now, but already somewhat useful. You can invoke `gulp test` to run all the go tests (at the moment Goose doesn't have any JS tests). You may want to set the environment variables `GOOSE_TEST_FILE` and `GOOSE_TEST_SQL` to the appropriate URIs, to test DocumentStore implementation compliance.

```bash
$ export GOOSE_TEST_FILE=file:///tmp/goose_test
$ export GOOSE_TEST_SQL=postgres://gooser@:49153/goosetest?sslmode=disable
$ gulp test
```

## Usage

Right now the application is very bare-bones, but it does actually do the basic jobs of a wiki (reading and writing pages). The route `/w/foo/bar` will take you to the page `/foo/bar`, while `/e/foo/bar` lets you edit or create that page. Rendering is done client-side in JS; commonmark compliance via [remarkable](https://github.com/jonschlinkert/remarkable) is on the roadmap but not really important atm.

## Configuration

Configuration is by environment variables. Most of them are set for you by gulp if you use the dev server, but if you run the binary straight from the command line then you will need to know these.

- `GOOSE_PORT` server port, eg `:4567` (note leading colon)
- `GOOSE_BACKEND` the backend URI, eg `file:///tmp/goose`
- `GOOSE_DEV` to enable development-only behavior, eg template recompilation on every request
- `GOOSE_TEST_FILE`, `GOOSE_TEST_SQL` to run tests against various DocumentStore implementations. The test suite does basic API sanity checks.

## License

BSD 3-clause, see [LICENSE.md](LICENSE.md)
