package e2eTesting

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/archway-network/archway/app"
)

var _ grpc.ClientConnInterface = (*grpcClient)(nil)

type grpcClient struct {
	app *app.ArchwayApp
}

func (c grpcClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	req := args.(codec.ProtoMarshaler)
	resp := c.app.Query(abci.RequestQuery{
		Data:   c.app.AppCodec().MustMarshal(req),
		Path:   method,
		Height: 0, // TODO: heightened queries
		Prove:  false,
	})

	if resp.Code != abci.CodeTypeOK {
		return fmt.Errorf(resp.Log)
	}

	c.app.AppCodec().MustUnmarshal(resp.Value, reply.(codec.ProtoMarshaler))

	return nil
}

func (c grpcClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("not supported")
}

func (chain *TestChain) Client() grpc.ClientConnInterface {
	return grpcClient{app: chain.app}
}
