marked.setOptions({
  langPrefix: 'language-',
  highlight: function(code, lang) {
    if (!Prism.languages[lang]) {
      return code;
    }
    return Prism.highlight(code, Prism.languages[lang]);
  },
});

var md = document.getElementById('md');
md.innerHTML = marked(md.getAttribute('data-md'));

(function() {
  var tables = document.querySelectorAll('table');
  for (var i = 0, tableElement ; tableElement = tables[i++] ; ) {
    tableElement.className = 'table';
  }
})();
