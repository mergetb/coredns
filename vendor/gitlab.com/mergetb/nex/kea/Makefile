prefix ?= /usr/local

.PHONY: all
all: nex.so

CLEANFILES = *.gcno *.gcda messages.h messages.cc nex.so
SOURCES = messages.cc version.cc nex.cc load_unload.cc pkt4_send.cc pkt4_receive.cc subnet4_select.cc lease4_select.cc lease_manager.cc network.cc
HEADERS = messages.h nex.h lease_manager.h network.h

CXXFLAGS=-std=c++17 -I /usr/include/kea -I /usr/local/include/kea -I `pwd` -I /usr/include/etcd -I /usr/local/include/etcd -O2
LDFLAGS=-L /usr/local/lib -fpic -shared -lkea-dhcpsrv -lkea-dhcp++ -lkea-hooks -lkea-log -lkea-util -lkea-exceptions -letcd-cpp-api

messages.h messages.cc: messages.mes
	kea-msg-compiler messages.mes


nex.so: $(SOURCES) $(HEADERS) Makefile
	$(CXX) $(CXXFLAGS) -o nex.so $(SOURCES) $(LDFLAGS)

.PHONY: clean
clean:
	rm -f $(CLEANFILES)

.PHONY: install
install: nex.so
	mkdir -p $(prefix)/lib/kea/hooks
	cp nex.so $(prefix)/lib/kea/hooks/nex.so

