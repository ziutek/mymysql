include $(GOROOT)/src/Make.inc

ALL=simple prepared_stmt reconnect long_data

all: $(ALL)

clean:
	rm -rf *.[68] $(ALL)

%: %.go
	$(GC) $*.go
	$(LD) -o $@ $@.$O
