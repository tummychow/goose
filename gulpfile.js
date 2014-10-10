var path = require('path');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();
var bower_components = require('main-bower-files')();

var gopath = function() {
  var gpm_dir = __dirname + path.sep + '.godeps';
  if (process.env.GOPATH.split(path.delimiter).indexOf(gpm_dir) != -1) {
    return process.env.GOPATH;
  }
  return gpm_dir + path.delimiter + process.env.GOPATH;
}

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
  var newenv = process.env;
  newenv.GOPATH = gopath();
  return $.run('go build -o goose', {env: newenv}).exec();
});

gulp.task('default', function() {
  // place code for your default task here
});
