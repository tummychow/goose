var gulp = require('gulp');
var $ = require('gulp-load-plugins')();
var bower_components = require('main-bower-files')();

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
  return $.run('GOPATH=$PWD/.godeps:$GOPATH go build').exec();
});

gulp.task('default', function() {
  // place code for your default task here
});
