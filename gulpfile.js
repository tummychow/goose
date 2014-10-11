var path = require('path');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();

var bower_components = require('main-bower-files')();

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
  return gulp.src(bower_components)
    .pipe($.filter('**/*.css'))
    .pipe($.concat('main.css'))
    .pipe($.minifyCss())
    .pipe(gulp.dest('./public'));
});

gulp.task('js', function() {
  return gulp.src(bower_components)
    .pipe($.filter('**/*.js'))
    .pipe($.concat('main.js'))
    .pipe($.uglify())
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
