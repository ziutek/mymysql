include $(GOROOT)/src/Make.inc

TARG=github.com/ziutek/mymysql
GOFILES=\
	mysql.go\
	init.go\
	common.go\
	packet.go\
	codecs.go\
	errors.go\
	consts.go\
	command.go\
	result.go\
	addons.go\
	prepared.go\
	binding.go\
	unsafe.go\
	autoconnect.go\

include $(GOROOT)/src/Make.pkg

doc: 
	godoc -html -path='.' '.' >doc/godoc.html
	cat doc/head1.html $(GOROOT)/doc/all.css doc/head2.html >GODOC.html
	doc/make_toc.py doc/godoc.html >doc/toc.html
	cat doc/toc.html doc/godoc.html doc/tail.html >>GODOC.html
	rm -f README.md
	cp doc/readme.md README.md
	cat doc/toc.html doc/godoc.html | sed 's/^[[:space:]]*//g' >>README.md
	chmod a-w README.md
	godoc -path='.' '.' >GODOC.txt
	rm -f doc/godoc.html doc/toc.html

.PHONY : doc
