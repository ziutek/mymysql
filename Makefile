install:
	./make.bash install

clean:
	./make.bash clean

test:
	./make.bash test

all:
	./make.bash clean install test

.PHONY : install clean test all
