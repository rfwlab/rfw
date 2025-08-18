(function(){
  hljs.registerLanguage('rtml', function(hljs){
    const xml = hljs.getLanguage('xml');
    const interpolation = {
      className: 'template-variable',
      begin: /\{/,
      end: /\}/,
      relevance: 0
    };
    const rtml = hljs.inherit(xml, {
      contains: xml.contains.concat([interpolation])
    });
    return rtml;
  });
})();
