import './src/scss/bootstrap.scss'

// highlightjs-line-numbers.js doesn't support async ES6 module loading, so we need to make sure window.hljs exists
import './bootstrap'

import RelativeTime from 'dayjs/plugin/relativeTime' // ES 2015
dayjs.extend(RelativeTime)

import './src/js/scroll-to-top'

// For node attribute table functionality
import 'bootstrap/js/dist/tab'

// For custom links functionality
import 'bootstrap/js/dist/dropdown'

import 'highlightjs-line-numbers.js';

// Register only a subset of languages since cookbook files probably don't have much else...
import ruby from 'highlight.js/lib/languages/ruby';
import erb from 'highlight.js/lib/languages/erb';
import yaml from 'highlight.js/lib/languages/yaml';
import json from 'highlight.js/lib/languages/json';
import bash from 'highlight.js/lib/languages/bash';

hljs.registerLanguage('bash', bash);
hljs.registerLanguage('ruby', ruby);
hljs.registerLanguage('erb', erb);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('json', json);
