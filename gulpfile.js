'use strict';

const port = '8080';

const gulp = require('gulp');
const rename = require('gulp-rename');
const notify = require("gulp-notify");
const notifier = require('node-notifier');

/* babel */
const babel = require('gulp-babel');

const jsmin = require('gulp-jsmin');


/* native build in single file */
const concat = require('gulp-concat');
// const autopolyfiller = require('gulp-autopolyfiller');


gulp.task('common-js', () => {

  let isBabel = babel({
    presets: [require('babel-preset-es2015')]
  });

  isBabel.on('error', function(e) {
    console.log(e);
    isBabel.end();
    notifier.notify(`error JS: ${ e.message }`);
  });

  gulp.src(`./common-js/**/*.js`)
    .pipe(isBabel)
    .pipe(rename({
      prefix: `common-`,
      dirname: ''
    }))
    .pipe(gulp.dest(`./js/common/`));

  require('child_process').exec(`curl localhost:${ port }/recache`, function (err, stdout, stderr) {
    console.log(stdout);
    // console.log(stderr);
    notifier.notify(`recache: ${ stdout }`)
  });
});



gulp.task('native-js', () => {
  let isBabel = babel({
    presets: [require('babel-preset-es2015')]
  });

  isBabel.on('error', function(e) {
    console.log(e);
    isBabel.end();
    notifier.notify(`error JS: ${ e.message }`);
  });

  gulp.src(`./nativeJS/**/*.js`)
    .pipe(isBabel)
    .pipe(concat('native.min.js'))
    // .pipe(autopolyfiller(`./js/autopolyfiller.js`, {
    //   browsers: ['last 2 version', 'ie 9']
    // }))
    // .pipe(jsmin())
    .pipe(rename({dirname: ''}))
    .pipe(gulp.dest(`./js/`));
    // .pipe(open({uri: recacheURL}))


  require('child_process').exec(`curl localhost:${ port }/recache`, function (err, stdout, stderr) {
    console.log(stdout);
    // console.log(stderr);
    notifier.notify(`recache: ${ stdout }`)
  });

});


/* watch */
gulp.task('watch', () => {
  gulp.watch([`./nativeJS/**/*`], ['native-js']);
  gulp.watch([`./common-js/**/*`], ['common-js']);
});


/* default */
gulp.task('default', ['native-js', 'common-js', 'watch']);