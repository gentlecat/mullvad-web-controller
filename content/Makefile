deps :
	npm ci

styles : deps
	./node_modules/sass/sass.js \
		./static/styles/main.scss:./static/styles/main.css \
		--load-path=node_modules --update --quiet-deps

styles-watch : deps
	./node_modules/sass/sass.js \
		./static/styles/main.scss:./static/styles/main.css \
		--load-path=node_modules --watch --quiet-deps

copy-choices : deps
	cp node_modules/choices.js/public/assets/scripts/choices.min.js static/js/choices.min.js
	cp node_modules/choices.js/public/assets/styles/choices.min.css static/styles/choices.min.css

build : copy-choices styles

serve : build
	echo "Starting server on http://localhost:8080/"
	python3 -m http.server 8080 --bind 127.0.0.1
