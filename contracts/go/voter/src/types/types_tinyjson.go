// Code generated by tinyjson for marshaling/unmarshaling. DO NOT EDIT.

package types

import (
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

func tinyjsonA17a9c65DecodeGithubComArchwayNetworkVoterSrcTypes(in *jlexer.Lexer, out *Params) {
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
		case "owner_addr":
			out.OwnerAddr = string(in.String())
		case "new_voting_cost":
			out.NewVotingCost = string(in.String())
		case "vote_cost":
			out.VoteCost = string(in.String())
		case "ibc_send_timeout":
			out.IBCSendTimeout = uint64(in.Uint64())
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
func tinyjsonA17a9c65EncodeGithubComArchwayNetworkVoterSrcTypes(out *jwriter.Writer, in Params) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"owner_addr\":"
		out.RawString(prefix[1:])
		out.String(string(in.OwnerAddr))
	}
	{
		const prefix string = ",\"new_voting_cost\":"
		out.RawString(prefix)
		out.String(string(in.NewVotingCost))
	}
	{
		const prefix string = ",\"vote_cost\":"
		out.RawString(prefix)
		out.String(string(in.VoteCost))
	}
	{
		const prefix string = ",\"ibc_send_timeout\":"
		out.RawString(prefix)
		out.Uint64(uint64(in.IBCSendTimeout))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Params) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	tinyjsonA17a9c65EncodeGithubComArchwayNetworkVoterSrcTypes(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalTinyJSON supports tinyjson.Marshaler interface
func (v Params) MarshalTinyJSON(w *jwriter.Writer) {
	tinyjsonA17a9c65EncodeGithubComArchwayNetworkVoterSrcTypes(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Params) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	tinyjsonA17a9c65DecodeGithubComArchwayNetworkVoterSrcTypes(&r, v)
	return r.Error()
}

// UnmarshalTinyJSON supports tinyjson.Unmarshaler interface
func (v *Params) UnmarshalTinyJSON(l *jlexer.Lexer) {
	tinyjsonA17a9c65DecodeGithubComArchwayNetworkVoterSrcTypes(l, v)
}
