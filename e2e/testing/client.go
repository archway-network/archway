package e2eTesting

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	proto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/app"
)

var _ grpc.ClientConnInterface = (*grpcClient)(nil)

type grpcClient struct {
	app *app.ArchwayApp
}

func (c grpcClient) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	req := args.(proto.Message)
	resp, err := c.app.Query(ctx, &abci.RequestQuery{
		Data:   c.app.AppCodec().MustMarshal(req),
		Path:   method,
		Height: 0, // TODO: heightened queries
		Prove:  false,
	})
	if err != nil {
		return err
	}

	if resp.Code != abci.CodeTypeOK {
		return fmt.Errorf("%s", resp.Log)
	}

	c.app.AppCodec().MustUnmarshal(resp.Value, reply.(proto.Message))

	return nil
}

func (c grpcClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("not supported")
}

func (chain *TestChain) Client() grpc.ClientConnInterface {
	return grpcClient{app: chain.app}
}

// SetupClientCtx configures the client and server contexts and returns the
// resultant 'context.Context'. This is useful for executing CLI commands.
func (chain *TestChain) SetupClientCtx() context.Context {
	home := chain.t.TempDir()
	logger := chain.app.Logger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(chain.t, err)

	appCodec := moduletestutil.MakeTestEncodingConfig().Codec
	testModuleBasicManager := module.NewBasicManager(genutil.AppModuleBasic{})
	err = genutiltest.ExecInitCmd(
		testModuleBasicManager, home, appCodec)
	require.NoError(chain.t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	return ctx
}
