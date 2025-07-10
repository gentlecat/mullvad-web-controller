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
	cp node_modules/htmx-ext-json-enc/json-enc.js static/js/htmx-json-enc.js

copy-choices : deps
	cp node_modules/choices.js/public/assets/scripts/choices.min.js static/js/choices.min.js
	cp node_modules/choices.js/public/assets/styles/choices.min.css static/styles/choices.min.css

build : copy-htmx copy-choices styles

serve : build
	echo "Starting server on http://localhost:8080/"
	python3 -m http.server 8080 --bind 127.0.0.1
