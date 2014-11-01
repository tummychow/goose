marked.setOptions({langPrefix: 'language-'});

var md = document.getElementById('md');
md.innerHTML = marked(md.getAttribute('data-md'));

(function() {
  var tables = document.querySelectorAll('table');
  for (var i = 0, tableElement ; tableElement = tables[i++] ; ) {
    tableElement.className = 'table';
  }
})();
