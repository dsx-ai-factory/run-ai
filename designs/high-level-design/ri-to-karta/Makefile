PUML_FILES := $(wildcard *.puml)
SVG_FILES  := $(PUML_FILES:.puml=.svg)

.PHONY: all clean

all: $(SVG_FILES)

%.svg: %.puml
	plantuml -tsvg $<

clean:
	rm -f $(SVG_FILES)
