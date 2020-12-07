package api

import "strings"

const (
	ERROR_UNKNOWN                     = 0x00000001
	ERROR_REQUEST_UNAUTHPRIZED        = 0x00000191
	ERROR_REQUEST_NODFOUND            = 0x00000194
	ERROR_PARAM                       = 0x00010101
	ERROR_TRANSACTION_FEES_NOTENOUGH  = 0x00020101
	ERROR_TRANSACTION_RAWTX_EMPTY     = 0x00020201
	ERROR_TRANSACTION_RAWTX_TOOBIG    = 0x00020202
	ERROR_TRANSACTION_TXID_NOTEXIST   = 0x00020301
	ERROR_TRANSACTION_TXID_USED       = 0x00020302
	ERROR_TRANSACTION_TXID_HAVE       = 0x00020303
	ERROR_TRANSACTION_CONNECT_REFUSED = 0x00020401
	ERROR_STATS_ADDR_EMPTY            = 0x00030101
	ERROR_FORM_INIT_FAILED            = 0x00050101
	ERROR_TOKEN_GET_NOTENOUGH         = 0x00060101
	ERROR_BLOCK_FIND_FAILED           = 0x00070101
	ERROR_ADDRESS_INFO_FAILED         = 0x00080101
	ERROR_AUTH_LOGIN_FAILED           = 0x00090101
	Error_BanTx                       = 0x00100001
)

type Error struct {
	Code    int
	Message string
}

func (err *Error) DealError() *Result {
	if err.Code == ERROR_UNKNOWN {
		err.parseError()
	}
	return &Result{err.Code, err.Message, ""}
}

func (err *Error) parseError() {
	if strings.Contains(err.Message, "already have transaction") {
		err.Code = ERROR_TRANSACTION_TXID_HAVE
	} else if strings.Contains(err.Message, "in the pool already spends the same coins") {
		err.Code = ERROR_TRANSACTION_TXID_USED
	} else if strings.Contains(err.Message, "connection refused") {
		err.Code = ERROR_TRANSACTION_CONNECT_REFUSED
	} else if strings.Contains(err.Message, "fees which is under the required amount of") {
		err.Code = ERROR_TRANSACTION_FEES_NOTENOUGH
	} else if strings.Contains(err.Message, "No information available about transaction") {
		err.Code = ERROR_TRANSACTION_TXID_NOTEXIST
	} else if strings.Contains(err.Message, "There is not enough balance") {
		err.Code = ERROR_TOKEN_GET_NOTENOUGH
	} else if strings.Contains(err.Message, "is larger than max allowed size of 100000") {
		err.Code = ERROR_TRANSACTION_RAWTX_TOOBIG
	} else {
		err.Code = ERROR_UNKNOWN
	}
}
