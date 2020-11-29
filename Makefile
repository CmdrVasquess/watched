.PHONY: depgraph.svg

deps: depgraph.svg

depgraph.svg:
	go mod graph | modgraphviz | dot -Tsvg -o $@
