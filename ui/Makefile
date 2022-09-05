.PHONY: clean dist

deps: ./node_modules clean

.PHONY: clean
clean:
	rm -rf ./dist

build: deps
	yarn run build

.PHONY: dist
dist: clean
	rm -rf ./dist

.PHONY: lint
lint: deps
	yarn run lint

.PHONY: format
format: deps
	yarn run format

../../node_modules: yarn.lock package.json
	yarn install
