# markdown-it-deflist

[![Build Status](https://img.shields.io/travis/markdown-it/markdown-it-deflist/master.svg?style=flat)](https://travis-ci.org/markdown-it/markdown-it-deflist)
[![NPM version](https://img.shields.io/npm/v/markdown-it-deflist.svg?style=flat)](https://www.npmjs.org/package/markdown-it-deflist)
[![Coverage Status](https://img.shields.io/coveralls/markdown-it/markdown-it-deflist/master.svg?style=flat)](https://coveralls.io/r/markdown-it/markdown-it-deflist?branch=master)

> Definition list (`<dl>`) tag plugin for [markdown-it](https://github.com/markdown-it/markdown-it) markdown parser.

Syntax is based on [pandoc definition lists](http://johnmacfarlane.net/pandoc/README.html#definition-lists).


## Install

```bash
npm install markdown-it-deflist --save
```

## Use

```js
var md = require('markdown-it')()
            .use(require('markdown-it-deflist'));

md.render(/*...*/);
```

_Differences in browser._ If you load script directly into the page, without
package system, module will add itself globally as `window.markdownitDeflist`.
