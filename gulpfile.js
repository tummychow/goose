var path = require('path');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

(function() {
  var gpm_dir = __dirname + path.sep + '.godeps';
  var argv = require('minimist')(process.argv);

  process.env.GOOSE_DEV = 'true';

  if (process.env.GOPATH.split(path.delimiter).indexOf(gpm_dir) == -1) {
    process.env.GOPATH = gpm_dir + path.delimiter + process.env.GOPATH;
  }
  if (!process.env.GOOSE_PORT) {
    process.env.GOOSE_PORT = ':' + (process.env.PORT || argv.port || '8000');
  }
  if (!process.env.GOOSE_BACKEND) {
    process.env.GOOSE_BACKEND = argv.backend || 'file:///tmp/goose';
  }
})();

gulp.task('css', function() {
  // minifyCss does not support sourcemaps, see jakubpawlowicz/clean-css#125
  return gulp.src([
      'bower_components/min/compiled/*.css',
      '!bower_components/min/compiled/*.min.css',
      'bower_components/prism/themes/prism.css',
    ])
    .pipe($.filter('**/*.css'))
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
    .pipe($.filter('**/*.js'))
    .pipe($.sourcemaps.init())
      .pipe($.concat('main.js'))
      .pipe($.uglify())
    .pipe($.sourcemaps.write('.'))
    .pipe(gulp.dest('./public'));
});

gulp.task('go', function() {
  return $.run('go build -o goose').exec();
});

gulp.task('test', function() {
  return $.run('go test ./...').exec();
});

gulp.task('serve', function() {
  console.log('Starting goose on port ' + process.env.GOOSE_PORT + ' with backend ' + process.env.GOOSE_BACKEND);
  return $.run('./goose').exec();
});

gulp.task('default', ['go', 'css', 'js'], function() {});
