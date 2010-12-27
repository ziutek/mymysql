package mymy

import (
    "os"
    "fmt"
)

var (
    WR_BUF_ERROR          = os.NewError("write buffer/packet too short")
    SEQ_ERROR             = os.NewError("sequence error")
    PKT_ERROR             = os.NewError("malformed packet")
    PKT_LONG_ERROR        = os.NewError("packet too long")
    UNEXP_NULL_LCS_ERROR  = os.NewError("unexpected null LCS")
    UNEXP_NULL_LCB_ERROR  = os.NewError("unexpected null LCB")
    UNK_RESULT_PKT_ERROR  = os.NewError("unexpected or unknown result packet")
    NOT_CONN_ERROR        = os.NewError("not connected")
    ALREDY_CONN_ERROR     = os.NewError("not connected")
    BAD_RESULT_ERROR      = os.NewError("unexpected result")
    UNREADED_ROWS_ERROR   = os.NewError("there are unreaded rows")
)

type Error struct {
    code  uint16
    msg   []byte
}

func (err Error) String() string {
    return fmt.Sprintf("Received #%d error from MySQL server: \"%s\"",
        err.code, err.msg)
}
