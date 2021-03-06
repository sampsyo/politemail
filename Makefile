NODE_DEPS := bower@~1.3.9 less@~1.7.4
BOWER_DEPS := bootstrap\#~3.2.0 jquery\#~2.1.1 hogan\#~3.0.2
GO_SRC := main.go tmplpool/pool.go app/app.go app/db.go app/handlers.go
STATIC := static/politemail.css static/jquery.js static/hogan.js
TARGET := politemail

.PHONY: setup clean run build

build: $(TARGET) $(STATIC)

# Somewhat dumb way to invoke setup on first run (but not thereafter) or on
# manual invocation.
BOWER_STUFF := bower_components/bootstrap/bower.json
$(BOWER_STUFF):
	npm install $(NODE_DEPS)
	./node_modules/bower/bin/bower install $(BOWER_DEPS)
setup: $(BOWER_STUFF)

# Build static components.
static/politemail.css: template/politemail.less $(BOWER_STUFF)
	./node_modules/less/bin/lessc $(LESSARGS) $< $@
static/jquery.js: bower_components/jquery/dist/jquery.min.js
	cp $< $@
static/hogan.js: bower_components/hogan/web/builds/3.0.2/hogan-3.0.2.min.js
	cp $< $@

clean:
	rm -rf node_modules bower_components $(TARGET) $(STATIC)

politemail: $(GO_SRC)
	go build

run: $(TARGET) $(STATIC)
	./$< -debug
