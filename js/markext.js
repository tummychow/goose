marked.setOptions({langPrefix: 'language-'});

var xsser = (function() {
  var globalAttr = ['accesskey', 'dir', 'itemprop', 'lang', 'tabindex', 'title'];
  return new filterXSS.FilterXSS({
    whiteList: {
      a: globalAttr + ['download', 'href', 'hreflang', 'media', 'rel', 'target'],
      b: globalAttr,
      blockquote: globalAttr + ['cite'],
      br: globalAttr,
      code: globalAttr,
      dd: globalAttr,
      del: globalAttr + ['cite', 'datetime'],
      div: globalAttr + ['itemscope', 'itemtype'],
      dl: globalAttr,
      dt: globalAttr,
      em: globalAttr,
      h1: globalAttr,
      h2: globalAttr,
      h3: globalAttr,
      h4: globalAttr,
      h5: globalAttr,
      h6: globalAttr,
      hr: globalAttr + ['align'],
      i: globalAttr,
      img: globalAttr + ['align', 'alt', 'height', 'src', 'width'],
      ins: globalAttr + ['cite', 'datetime'],
      kbd: globalAttr,
      li: globalAttr + ['value'],
      ol: globalAttr + ['start'],
      p: globalAttr,
      pre: globalAttr,
      q: globalAttr + ['cite'],
      rp: globalAttr,
      rt: globalAttr,
      ruby: globalAttr,
      s: globalAttr,
      samp: globalAttr,
      strong: globalAttr,
      sub: globalAttr,
      sup: globalAttr,
      table: globalAttr + ['align'],
      tbody: globalAttr + ['align'],
      td: globalAttr + ['align', 'colspan', 'rowspan'],
      tfoot: globalAttr + ['align'],
      th: globalAttr + ['align', 'colspan', 'rowspan', 'scope'],
      thead: globalAttr + ['align'],
      tr: globalAttr + ['align'],
      u: globalAttr,
      ul: globalAttr,
      'var': globalAttr,
    },
    stripIgnoreTag: true,
    stripIgnoreTagBody: ['script', 'style'],
    onIgnoreTagAttr: function(tag, name, value) {
      if (tag == 'code' && name == 'class' && value.substr(0,9) == 'language-') {
        return name + '="' + filterXSS.escapeAttrValue(value) + '"';
      }
    }
  });
})();

var md = document.getElementById('md');
md.innerHTML = xsser.process(marked(md.getAttribute('data-md')));

(function() {
  var tables = document.querySelectorAll('table');
  for (var i = 0, tableElement ; tableElement = tables[i++] ; ) {
    tableElement.className = 'table';
  }
})();
