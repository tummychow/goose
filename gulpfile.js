var path = require('path');

var gulp = require('gulp');
var $ = require('gulp-load-plugins')();
var bower_components = require('main-bower-files')();

var newenv = process.env;
var gpm_dir = __dirname + path.sep + '.godeps';
if (newenv.GOPATH.split(path.delimiter).indexOf(gpm_dir) == -1) {
  newenv.GOPATH = gpm_dir + path.delimiter + newenv.GOPATH;
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
  return $.run('go build', {env: newenv}).exec();
});

gulp.task('default', function() {
  // place code for your default task here
});
