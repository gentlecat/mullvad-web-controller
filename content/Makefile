deps :
	npm install

styles : deps
	./node_modules/sass/sass.js \
		./static/styles/main.scss:./static/styles/main.css \
		--update

styles-watch : deps
	./node_modules/sass/sass.js \
		./static/styles/main.scss:./static/styles/main.css \
		--watch

copy-htmx : deps
	cp node_modules/htmx.org/dist/htmx.min.js static/js/htmx.min.js
	cp node_modules/htmx.org/dist/ext/json-enc.js static/js/htmx-json-enc.js

build : copy-htmx styles

serve : build
	echo "Starting server on http://localhost:8080/"
	python3 -m http.server 8080 --bind 127.0.0.1
