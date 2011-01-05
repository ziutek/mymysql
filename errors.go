package mymy

import (
    "os"
    "fmt"
)

var (
    WR_BUF_ERROR          = os.NewError("write buffer/packet too short")
    SEQ_ERROR             = os.NewError("packet sequence error")
    PKT_ERROR             = os.NewError("malformed packet")
    PKT_LONG_ERROR        = os.NewError("packet too long")
    UNEXP_NULL_LCS_ERROR  = os.NewError("unexpected null LCS")
    UNEXP_NULL_LCB_ERROR  = os.NewError("unexpected null LCB")
    UNEXP_NULL_DATE_ERROR = os.NewError("unexpected null datetime")
    UNK_RESULT_PKT_ERROR  = os.NewError("unexpected or unknown result packet")
    NOT_CONN_ERROR        = os.NewError("not connected")
    ALREDY_CONN_ERROR     = os.NewError("not connected")
    BAD_RESULT_ERROR      = os.NewError("unexpected result")
    UNREADED_ROWS_ERROR   = os.NewError("there are unreaded rows")
    BIND_COUNT_ERROR      = os.NewError("wrong number of values for bind")
    BIND_UNK_TYPE         = os.NewError("unknown bind value type")
    RESULT_COUNT_ERROR    = os.NewError("wrong number of result columns")
    BAD_COMMAND_ERROR     = os.NewError("comand isn't text SQL nor *Statement")
    WRONG_DATE_LEN_ERROR  = os.NewError("wrong datetime/timestamp length")
    UNK_MYSQL_TYPE_ERROR  = os.NewError("unknown MySQL type")
)

type Error struct {
    code  uint16
    msg   []byte
}

func (err Error) String() string {
    return fmt.Sprintf("Received #%d error from MySQL server: \"%s\"",
        err.code, err.msg)
}
