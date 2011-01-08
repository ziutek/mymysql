#!/usr/bin/env python
import sys
import re

if len(sys.argv) != 2 or sys.argv[1] == "-h":
    print "Usage: %s GODOC_HTML_FILE" % sys.argv[0]
    sys.exit(1)

fp = open(sys.argv[1])

regexp = re.compile('.*<h(2|3) id="(.*)">(.*) <a .*>(.*)</a></h(2|3)>')

def getTypeId(line):
    ti = regexp.match(line)
    if ti is not None:
        return ti.group(2), " ".join((ti.group(3), ti.group(4)))
    return None

class Type:
    def __init__(self, idn, name):
        self.idn = idn
        self.name = name
        self.methods = []


print "<dl>"
for line in fp:
    ti = getTypeId(line)
    if ti is None:
        continue
    idn, name = ti
    if "." in idn:
        print "    <dd><a href='#%s'>%s</a></dd>" % (idn, name)
    else:
        print "  <dt><a href='#%s'>%s</a></dt>" % (idn, name)
print "</dl>"
