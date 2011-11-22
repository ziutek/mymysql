include $(GOROOT)/src/Make.inc

ALL=simple prepared_stmt reconnect long_data transactions

install: $(ALL)

clean:
	rm -rf *.[68] $(ALL)

all: clean install

test:

%: %.go
	$(GC) $*.go
	$(LD) -o $@ $@.$O
