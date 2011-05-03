GOROOT ?= $(shell printf 't:;@echo $$(GOROOT)\n' | gomake -f -)
include $(GOROOT)/src/Make.inc

TARG=github.com/kr/pty
GOFILES=\
	pty_$(GOOS).go\
	run.go\

include $(GOROOT)/src/Make.pkg
