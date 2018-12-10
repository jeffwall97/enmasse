include ../Makefile.common
EXEC=$(TOPDIR)/cmd/enmasse-controller-manager
SUBDIRS=templates

build: $(SUBDIRS)
	go build $(EXEC)
	@echo "Nothing to build"

$(SUBDIRS):
	$(MAKE) -C $@

package:

test:

coverage:

.PHONY: $(SUBDIRS)
