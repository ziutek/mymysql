package mymy

import (
    "os"
    "fmt"
)

func (my *MySQL) Startv(a ...interface{}) (*Result, os.Error) {
    return my.Start(fmt.Sprint(a...))
}

func (my *MySQL) Startf(format string, a ...interface{}) (*Result, os.Error) {
    return my.Start(fmt.Sprintf(format, a...))
}

func (my *MySQL) Queryv(a ...interface{}) ([]*TextRow, *Result, os.Error) {
    return my.Query(fmt.Sprint(a...))
}

func (my *MySQL) Queryf(format string, a ...interface{}) (
        []*TextRow, *Result, os.Error) {
    return my.Query(fmt.Sprintf(format, a...))
}

func NbinToNstr(nbin Nbin) Nstr {
    if nbin == nil {
        return nil
    }
    str := string(*nbin)
    return &str
}

func NstrToNbin(nstr Nstr) Nbin {
    if nstr == nil {
        return nil
    }
    bin := []byte(*nstr)
    return &bin
}

