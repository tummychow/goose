marked.setOptions({
  langPrefix: 'language-',
  highlight: function(code, lang) {
    if (!Prism.languages[lang]) {
      return code;
    }
    return Prism.highlight(code, Prism.languages[lang]);
  },
  renderer: (function(){
    var renderMin = new marked.Renderer();
    renderMin.table = function(header, body) {
      return '<table class=\'table\'>\n'
      + '<thead>\n'
      + header
      + '</thead>\n'
      + '<tbody>\n'
      + body
      + '</tbody>\n'
      + '</table>\n';
    };
    return renderMin;
  })()
});

var md = document.getElementById('md');
md.innerHTML = marked(md.getAttribute('data-md'));
