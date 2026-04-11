package ingestion

import "errors"

var ErrCapReached = errors.New("ingestion: staging cap reached for this fetch")
