.PHONY: setup run

BOWER_STUFF := bower_components/bootstrap/bower.json
static/politemail.css: template/politemail.less $(BOWER_STUFF)
	./node_modules/less/bin/lessc $(LESSARGS) $< $@

# Somewhat dumb way to invoke setup on first run (but not thereafter) or on
# manual invocation.
$(BOWER_STUFF):
	npm install bower@~1.3.9
	npm install less@~1.7.4
	./node_modules/bower/bin/bower install
setup: $(BOWER_STUFF)

run: static/politemail.css
	go run politemail.go tmplcache.go
