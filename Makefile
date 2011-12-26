include $(GOROOT)/src/Make.inc

TARG:=github.com/dpc/gobtc

GOFILES:=\
	src/peer.go\
	src/server.go\
	src/protocol.go\

GOFILESOTHER:=\
	example.go\

all: package

example.go: | example.go.example
	ln -s example.go.example example.go



GOFMT=gofmt
BADFMT:=$(shell $(GOFMT) -l $(GOFILES) $(CGOFILES) $(GOFILESOTHER) $(wildcard *_test.go) 2> /dev/null)

gofmt: $(BADFMT) example.go
	@for F in $(BADFMT); do $(GOFMT) -w $$F && echo $$F; done


ifneq ($(BADFMT),)
ifneq ($(MAKECMDGOALS),gofmt)
	$(warning WARNING: make gofmt: $(BADFMT))
endif
endif

include $(GOROOT)/src/Make.pkg

run: example.go _obj/$(TARG).a
	$(GC) -I_obj -o example.$(O) example.go
	$(LD) -L_obj -o example example.$(O)
	./example
