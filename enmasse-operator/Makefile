include ../Makefile.common
EXEC=$(TOPDIR)/cmd/enmasse-controller-manager
SUBDIRS=templates

BUILDDIR=build
INSTALLNAME=enmasse-$(TAG)
INSTALLDIR=$(BUILDDIR)/$(INSTALLNAME)
PACKAGE_INSTALL_DIR=$(INSTALLDIR)/install
PACKAGE_ANSIBLE_DIR=$(TOPDIR)/ansible
PACKAGE_DOCS_DIR=$(TOPDIR)/documentation/html

build: $(SUBDIRS)
	go build $(EXEC)
	@echo "Nothing to build"

$(SUBDIRS):
	$(MAKE) -C $@

package:
	mkdir -p $(PACKAGE_INSTALL_DIR)/bundle
	cp -r install/* $(PACKAGE_INSTALL_DIR)/bundle/
	mkdir -p $(INSTALLDIR)/docs
	cp -r $(PACKAGE_DOCS_DIR)/* $(INSTALLDIR)/docs/
	mkdir -p $(PACKAGE_INSTALL_DIR)/ansible
	cp -r $(PACKAGE_ANSIBLE_DIR)/* $(PACKAGE_INSTALL_DIR)/ansible/
	tar -czf build/$(INSTALLNAME).tgz -C build $(INSTALLNAME)

test:

coverage:

.PHONY: $(SUBDIRS)
