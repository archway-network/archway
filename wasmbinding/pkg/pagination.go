package pkg

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

// PageRequest is the WASM bindings version of the query.PageRequest.
type PageRequest struct {
	// Key is a value returned in the PageResponse.NextKey to begin querying the next page most efficiently.
	// Only one of (Offset, Key) should be set.
	// Go converts a byte slice into a base64 encoded string, so this field could be a string in a contract code.
	Key []byte `json:"key"`
	// Offset is a numeric offset that can be used when Key is unavailable.
	// Only one of (Offset, Key) should be set.
	Offset uint64 `json:"offset"`
	// Limit is the total number of results to be returned in the result page.
	// It is set to default if left empty.
	Limit uint64 `json:"limit"`
	// CountTotal if set to true, indicates that the result set should include a count of the total number of items available for pagination.
	// This field is only respected when the Offset is used.
	// It is ignored when Key field is set.
	CountTotal bool `json:"count_total"`
	// Reverse if set to true, results are to be returned in the descending order.
	Reverse bool `json:"reverse"`
}

// PageResponse is the WASM bindings version of the query.PageResponse.
type PageResponse struct {
	// NextKey is the key to be passed to PageRequest.Key to query the next page most efficiently.
	NextKey []byte `json:"next_key"`
	// Total is the total number of results available if PageRequest.CountTotal was set, its value is undefined otherwise.
	Total uint64 `json:"total"`
}

// NewPageRequestFromSDK converts the SDK version of the query.PageRequest to the WASM bindings version.
func NewPageRequestFromSDK(pageReq query.PageRequest) PageRequest {
	return PageRequest{
		Key:        pageReq.Key,
		Offset:     pageReq.Offset,
		Limit:      pageReq.Limit,
		CountTotal: pageReq.CountTotal,
		Reverse:    pageReq.Reverse,
	}
}

// ToSDK converts the WASM bindings version of the query.PageResponse to the SDK version.
func (r PageRequest) ToSDK() query.PageRequest {
	return query.PageRequest{
		Key:        r.Key,
		Offset:     r.Offset,
		Limit:      r.Limit,
		CountTotal: r.CountTotal,
		Reverse:    r.Reverse,
	}
}

// NewPageResponseFromSDK converts the SDK version of the query.PageResponse to the WASM bindings version.
func NewPageResponseFromSDK(pageResp query.PageResponse) PageResponse {
	return PageResponse{
		NextKey: pageResp.NextKey,
		Total:   pageResp.Total,
	}
}

// ToSDK converts the WASM bindings version of the query.PageResponse to the SDK version.
func (r PageResponse) ToSDK() query.PageResponse {
	return query.PageResponse{
		NextKey: r.NextKey,
		Total:   r.Total,
	}
}
