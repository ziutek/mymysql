include $(GOROOT)/src/Make.inc

TARG=mymy
GOFILES=\
	common.go\
	packet.go\
	codecs.go\
	mysql.go\
	utils.go\
	errors.go\
	consts.go\
	command.go\
	result.go\
	addons.go\

include $(GOROOT)/src/Make.pkg
