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

include $(GOROOT)/src/Make.pkg

doc:
	cat godoc/head1.html $(GOROOT)/doc/all.css godoc/head2.html >GODOC.html
	godoc -html -path='.' '.' >>GODOC.html
	cat godoc/tail.html >>GODOC.html
	godoc -path='.' '.' >GODOC.txt
	rm -f README.md
	cp godoc/readme.md README.md
	godoc -html -path='.' '.' |html2markdown >>README.md
	chmod a-w README.md
