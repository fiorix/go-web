#
# css/js minification/compression makefile
#

#
#   JS_TARGETS -- js files to minify/gzip
#   CSS_TARGETS -- css files to minify/gzip
#   CLEANUP -- additional files to delete during "make clean"
#

JS_TARGETS = \
	public/js/recovery-confirm.js \
	public/js/recovery.js \
	public/js/signin.js \
	public/js/signup-confirm.js \
	public/js/signup.js \
	private/js/index.js \
	private/js/settings.js

CSS_TARGETS = \
	public/css/index.css \
	public/css/single-page-form.css

# google closure compiler needs every input script prefixed with --js=, as in --js=file1.js
#concatenated.min.js: file1.js file2.js
#	java -jar ~/bin/compiler.jar $(addprefix --js=,$^) >$@

#custom-concat.css: file1.css file2.css file3.css
#	cat $^ >$@

#######################################################
# you shouldn't need to edit anything below this line #
#######################################################

.DEFAULT_GOAL := all

all: js css

YUIJAR = /usr/local/Cellar/yuicompressor/2.4.7/libexec/yuicompressor-2.4.7.jar
YUI = java -jar $(YUIJAR)
YUI_FLAGS = --type css

CLOSURE = closure-compiler
CLOSURE_FLAGS = --angular_pass

.PHONY: css js

%.gz: %
	gzip -9 <$< >$@

# css
# ---

CSS_MINIFIED = $(CSS_TARGETS:.css=.min.css)
CSS_GZIP = $(CSS_TARGETS:.css=.css.gz)
CSS_MIN_GZIP = $(CSS_TARGETS:.css=.min.css.gz)

css: $(CSS_TARGETS) $(CSS_MINIFIED) $(CSS_GZIP) $(CSS_MIN_GZIP)

%.min.css: %.css
	 $(YUI) $(YUI_FLAGS) <$< | sed 's/ and(/ and (/g' >$@

# javascript
# ----------

JS_MINIFIED = $(JS_TARGETS:.js=.min.js)
JS_GZIP = $(JS_TARGETS:.js=.js.gz)
JS_MIN_GZIP = $(JS_TARGETS:.js=.min.js.gz)

js: $(JS_TARGETS) $(JS_MINIFIED) $(JS_GZIP) $(JS_MIN_GZIP)

%.min.js: %.js
	$(CLOSURE) $(CLOSURE_FLAGS) --js=$< >$@

clean:
	rm -f $(CSS_MINIFIED) $(CSS_GZIP) $(CSS_MIN_GZIP) $(JS_GZIP) $(JS_MINIFIED) $(JS_MIN_GZIP)
