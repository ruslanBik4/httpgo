'use strict';


const recacheURL = 'localhost:8080/recache';

const gulp = require('gulp');
const rename = require('gulp-rename');


/* babel */
const babel = require('gulp-babel');

const jsmin = require('gulp-jsmin');


/* native build in single file */
const concat = require('gulp-concat');

gulp.task('native-js', () => {
  return gulp.src(`./nativeJS/**/*.js`)
    .pipe(babel({
      presets: ['es2015']
    }))
    .pipe(concat('native.min.js'))
    // .pipe(jsmin())
    .pipe(rename({dirname: ''}))
    .pipe(gulp.dest(`./js/`));
    // .pipe(open({uri: recacheURL}))
});


/* watch */
gulp.task('watch', () => {
  gulp.watch([`./nativeJS/**/*`], ['native-js']);
});


/* default */
gulp.task('default', ['native-js', 'watch']);