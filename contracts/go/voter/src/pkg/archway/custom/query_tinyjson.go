// Code generated by tinyjson for marshaling/unmarshaling. DO NOT EDIT.

package custom

import (
	types "github.com/CosmWasm/cosmwasm-go/std/types"
	tinyjson "github.com/CosmWasm/tinyjson"
	jlexer "github.com/CosmWasm/tinyjson/jlexer"
	jwriter "github.com/CosmWasm/tinyjson/jwriter"
)

// suppress unused package warning
var (
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ tinyjson.Marshaler
)

func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(in *jlexer.Lexer, out *RewardsRecordsResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "records":
			if in.IsNull() {
				in.Skip()
				out.Records = nil
			} else {
				in.Delim('[')
				if out.Records == nil {
					if !in.IsDelim(']') {
						out.Records = make([]RewardsRecord, 0, 0)
					} else {
						out.Records = []RewardsRecord{}
					}
				} else {
					out.Records = (out.Records)[:0]
				}
				for !in.IsDelim(']') {
					var v1 RewardsRecord
					(v1).UnmarshalTinyJSON(in)
					out.Records = append(out.Records, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "pagination":
			(out.Pagination).UnmarshalTinyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(out *jwriter.Writer, in RewardsRecordsResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"records\":"
		out.RawString(prefix[1:])
		if in.Records == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Records {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalTinyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"pagination\":"
		out.RawString(prefix)
		(in.Pagination).MarshalTinyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RewardsRecordsResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v RewardsRecordsResponse) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RewardsRecordsResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *RewardsRecordsResponse) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(in *jlexer.Lexer, out *RewardsRecordsRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "rewards_address":
			out.RewardsAddress = string(in.String())
		case "pagination":
			if in.IsNull() {
				in.Skip()
				out.Pagination = nil
			} else {
				if out.Pagination == nil {
					out.Pagination = new(PageRequest)
				}
				(*out.Pagination).UnmarshalTinyJSON(in)
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(out *jwriter.Writer, in RewardsRecordsRequest) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"rewards_address\":"
		out.RawString(prefix[1:])
		out.String(string(in.RewardsAddress))
	}
	{
		const prefix string = ",\"pagination\":"
		out.RawString(prefix)
		if in.Pagination == nil {
			out.RawString("null")
		} else {
			(*in.Pagination).MarshalTinyJSON(out)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RewardsRecordsRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v RewardsRecordsRequest) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RewardsRecordsRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *RewardsRecordsRequest) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom1(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(in *jlexer.Lexer, out *RewardsRecord) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = uint64(in.Uint64())
		case "rewards_address":
			out.RewardsAddress = string(in.String())
		case "rewards":
			if in.IsNull() {
				in.Skip()
				out.Rewards = nil
			} else {
				in.Delim('[')
				if out.Rewards == nil {
					if !in.IsDelim(']') {
						out.Rewards = make([]types.Coin, 0, 2)
					} else {
						out.Rewards = []types.Coin{}
					}
				} else {
					out.Rewards = (out.Rewards)[:0]
				}
				for !in.IsDelim(']') {
					var v4 types.Coin
					(v4).UnmarshalTinyJSON(in)
					out.Rewards = append(out.Rewards, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "calculated_height":
			out.CalculatedHeight = int64(in.Int64())
		case "calculated_time":
			out.CalculatedTime = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(out *jwriter.Writer, in RewardsRecord) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Uint64(uint64(in.ID))
	}
	{
		const prefix string = ",\"rewards_address\":"
		out.RawString(prefix)
		out.String(string(in.RewardsAddress))
	}
	{
		const prefix string = ",\"rewards\":"
		out.RawString(prefix)
		if in.Rewards == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Rewards {
				if v5 > 0 {
					out.RawByte(',')
				}
				(v6).MarshalTinyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"calculated_height\":"
		out.RawString(prefix)
		out.Int64(int64(in.CalculatedHeight))
	}
	{
		const prefix string = ",\"calculated_time\":"
		out.RawString(prefix)
		out.String(string(in.CalculatedTime))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RewardsRecord) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v RewardsRecord) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RewardsRecord) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *RewardsRecord) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom2(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(in *jlexer.Lexer, out *PageResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "next_key":
			if in.IsNull() {
				in.Skip()
				out.NextKey = nil
			} else {
				out.NextKey = in.Bytes()
			}
		case "total":
			out.Total = uint64(in.Uint64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(out *jwriter.Writer, in PageResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"next_key\":"
		out.RawString(prefix[1:])
		out.Base64Bytes(in.NextKey)
	}
	{
		const prefix string = ",\"total\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.Total))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v PageResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v PageResponse) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *PageResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *PageResponse) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom3(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(in *jlexer.Lexer, out *PageRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "key":
			if in.IsNull() {
				in.Skip()
				out.Key = nil
			} else {
				out.Key = in.Bytes()
			}
		case "offset":
			out.Offset = uint64(in.Uint64())
		case "limit":
			out.Limit = uint64(in.Uint64())
		case "count_total":
			out.CountTotal = bool(in.Bool())
		case "reverse":
			out.Reverse = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(out *jwriter.Writer, in PageRequest) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"key\":"
		out.RawString(prefix[1:])
		out.Base64Bytes(in.Key)
	}
	{
		const prefix string = ",\"offset\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.Offset))
	}
	{
		const prefix string = ",\"limit\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.Limit))
	}
	{
		const prefix string = ",\"count_total\":"
		out.RawString(prefix)
		out.Bool(bool(in.CountTotal))
	}
	{
		const prefix string = ",\"reverse\":"
		out.RawString(prefix)
		out.Bool(bool(in.Reverse))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v PageRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v PageRequest) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *PageRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *PageRequest) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom4(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(in *jlexer.Lexer, out *CustomQuery) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "metadata":
			if in.IsNull() {
				in.Skip()
				out.Metadata = nil
			} else {
				if out.Metadata == nil {
					out.Metadata = new(ContractMetadataRequest)
				}
				(*out.Metadata).UnmarshalTinyJSON(in)
			}
		case "rewards_records":
			if in.IsNull() {
				in.Skip()
				out.RewardsRecords = nil
			} else {
				if out.RewardsRecords == nil {
					out.RewardsRecords = new(RewardsRecordsRequest)
				}
				(*out.RewardsRecords).UnmarshalTinyJSON(in)
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(out *jwriter.Writer, in CustomQuery) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"metadata\":"
		out.RawString(prefix[1:])
		if in.Metadata == nil {
			out.RawString("null")
		} else {
			(*in.Metadata).MarshalTinyJSON(out)
		}
	}
	{
		const prefix string = ",\"rewards_records\":"
		out.RawString(prefix)
		if in.RewardsRecords == nil {
			out.RawString("null")
		} else {
			(*in.RewardsRecords).MarshalTinyJSON(out)
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v CustomQuery) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v CustomQuery) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *CustomQuery) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *CustomQuery) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom5(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(in *jlexer.Lexer, out *ContractMetadataResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "owner_address":
			out.OwnerAddress = string(in.String())
		case "rewards_address":
			out.RewardsAddress = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(out *jwriter.Writer, in ContractMetadataResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"owner_address\":"
		out.RawString(prefix[1:])
		out.String(string(in.OwnerAddress))
	}
	{
		const prefix string = ",\"rewards_address\":"
		out.RawString(prefix)
		out.String(string(in.RewardsAddress))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ContractMetadataResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v ContractMetadataResponse) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ContractMetadataResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *ContractMetadataResponse) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom6(l, v)
}
func tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(in *jlexer.Lexer, out *ContractMetadataRequest) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "contract_address":
			out.ContractAddress = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(out *jwriter.Writer, in ContractMetadataRequest) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"contract_address\":"
		out.RawString(prefix[1:])
		out.String(string(in.ContractAddress))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ContractMetadataRequest) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v ContractMetadataRequest) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonAa6e548eEncodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ContractMetadataRequest) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *ContractMetadataRequest) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonAa6e548eDecodeGithubComArchwayNetworkVoterSrcPkgArchwayCustom7(l, v)
}
