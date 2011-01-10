include $(GOROOT)/src/Make.inc

TARG=mymy
GOFILES=\
	mysql.go\
	common.go\
	packet.go\
	codecs.go\
	utils.go\
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
	doc/make_toc.py doc/godoc.html >>GODOC.html
	cat doc/godoc.html doc/tail.html >>GODOC.html
	rm -f doc/godoc.html
	rm -f README.md
	cp doc/readme.md README.md
	html2markdown < GODOC.html |sed 's/^    \[func/> \[func/g' >>README.md
	chmod a-w README.md
	godoc -path='.' '.' >GODOC.txt

.PHONY : doc
