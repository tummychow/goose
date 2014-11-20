var path = require('path');
var child = require('child_process');

var argv = require('minimist')(process.argv);
var browsersync = require('browser-sync');
var forever = require('forever-monitor');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

(function() {
  var gpm_dir = __dirname + path.sep + '.godeps';

  process.env.GOOSE_DEV = 'true';
  // clobber this variable to make sure browsersync has its own port
  process.env.GOOSE_PORT = ':8005';

  if (process.env.GOPATH.split(path.delimiter).indexOf(gpm_dir) == -1) {
    process.env.GOPATH = gpm_dir + path.delimiter + process.env.GOPATH;
  }
  if (!process.env.GOOSE_BACKEND) {
    process.env.GOOSE_BACKEND = argv.backend || 'file:///tmp/goose';
  }
})();

// returns a gulp task that calls the given command, eg
// gulp.task('foo', spawntask(['foo', 'bar', 'baz']));
function spawntask(cmdargs) {
  return (function(cb) {
    $.util.log($.util.colors.blue(cmdargs.join(' ')));
    child.spawn(cmdargs[0], cmdargs.slice(1), {
      cwd: process.cwd(),
      env: process.env,
      stdio: 'inherit',
    }).on('exit', function(code, signal) {
      if (code !== 0) {
        cb(code);
      } else {
        cb();
      }
    });
  });
}

gulp.task('go', spawntask(['go', 'build', '-o', 'goose']));

gulp.task('test', spawntask(['go', 'test', './...', '-test.v', '-check.v']));

gulp.task('css', function() {
  // minifyCss does not support sourcemaps, see jakubpawlowicz/clean-css#125
  return gulp.src([
      'bower_components/min/compiled/*.css',
      '!bower_components/min/compiled/*.min.css',
      'bower_components/prism/themes/prism.css',
      'css/*.css',
    ])
    .pipe($.concat('main.css'))
    .pipe($.minifyCss())
    .pipe(gulp.dest('./public'));
});

gulp.task('js', function() {
  return gulp.src([
      'bower_components/marked/lib/marked.js',
      'bower_components/xss/dist/xss.js',
      // this list is order-sensitive, since some languages extend others
      // grep for "Prism.languages.extend" to find the dependencies
      'bower_components/prism/components/prism-{core,clike,css,c,bash,cpp,git,go,haskell,http,ini,java,javascript,latex,markup,python,ruby,scss,sql}.js',
      'js/*.js',
    ])
    .pipe($.sourcemaps.init())
      .pipe($.concat('main.js'))
      .pipe($.uglify())
    .pipe($.sourcemaps.write('.'))
    .pipe(gulp.dest('./public'));
});

gulp.task('default', function() {
  console.log('Using backend \"' + process.env.GOOSE_BACKEND + '\"');
  var gooseproc = forever.start(['./goose'], {
    max: 1,
    killTree: true,
    watch: false,
  });

  browsersync({
    proxy: "127.0.0.1" + process.env.GOOSE_PORT,
    port: argv.port || process.env.PORT || 8000,
    files: ['public/*.{js,css}', 'templates/*.tmpl'],
    online: false,
    open: false,
  });
  gooseproc.on('restart', function() {
    // brief delay to allow the goose binary to begin listening
    // fixes some sporadic connection reset issues where you issue a reload
    // and the browser gets stuck waiting for goose to respond
    setTimeout(browsersync.reload, 100);
  });

  gulp.watch('js/*.js', ['js']);
  gulp.watch('css/*.css', ['css']);

  gulp.watch('**/*.go', function() {
    // we cannot say this:
    // gulp.watch('**/*.go', ['go'], function(err) {...});
    // because the callback might get called before the 'go' task is finished,
    // and we need them to run in strictly series order
    gulp.start('go', function(err) {
      if (!err) {
        gooseproc.restart();
      }
    });
  });
});
