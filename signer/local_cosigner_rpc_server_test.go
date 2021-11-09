package signer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmProto "github.com/tendermint/tendermint/proto/tendermint/types"
	tm "github.com/tendermint/tendermint/types"
)

type DummyCosigner struct{}

func (cosigner *DummyCosigner) GetID() int {
	return 0
}

func (cosigner *DummyCosigner) Sign(signReq *CosignerSignRequest) (*CosignerSignResponse, error) {
	return &CosignerSignResponse{
		Signature: []byte("foobar"),
	}, nil
}

func (cosigner *DummyCosigner) GetEphemeralSecretPart(req *CosignerGetEphemeralSecretPartRequest) (*CosignerGetEphemeralSecretPartResponse, error) {
	return &CosignerGetEphemeralSecretPartResponse{
		SourceID:                       1,
		SourceEphemeralSecretPublicKey: []byte("foo"),
		EncryptedSharePart:             []byte("bar"),
		SourceSig:                      []byte("source sig"),
	}, nil
}

func (cosigner *DummyCosigner) HasEphemeralSecretPart(req CosignerHasEphemeralSecretPartRequest) (CosignerHasEphemeralSecretPartResponse, error) {
	return CosignerHasEphemeralSecretPartResponse{
		Exists: false,
	}, nil
}

func (cosigner *DummyCosigner) SetEphemeralSecretPart(req CosignerSetEphemeralSecretPartRequest) error {
	return nil
}

func TestCosignerRpcServerSign(test *testing.T) {
	dummyCosigner := &DummyCosigner{}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	config := CosignerRpcServerConfig{
		Logger:        logger,
		ListenAddress: "0.0.0.0:0",
		LocalCosigner: dummyCosigner,
	}

	rpcServer := NewCosignerRpcServer(&config)

	rpcServer.Start()

	// pack a vote into sign bytes
	var vote tmProto.Vote
	vote.Height = 1
	vote.Round = 0
	vote.Type = tmProto.PrevoteType
	signBytes := tm.VoteSignBytes("chain-id", &vote)

	remoteCosigner := NewRemoteCosigner(2, rpcServer.Addr().String())
	resp, err := remoteCosigner.Sign(&CosignerSignRequest{
		SignBytes: signBytes,
	})
	require.NoError(test, err)

	expctedRes := CosignerSignResponse{
		Signature: []byte("foobar"),
	}

	require.Equal(test, resp.Signature, expctedRes.Signature)
	rpcServer.Stop()
}

func TestCosignerRpcServerGetEphemeralSecretPart(test *testing.T) {
	dummyCosigner := &DummyCosigner{}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	config := CosignerRpcServerConfig{
		Logger:        logger,
		ListenAddress: "0.0.0.0:0",
		LocalCosigner: dummyCosigner,
	}

	rpcServer := NewCosignerRpcServer(&config)
	rpcServer.Start()

	//tempAddress := rpcServer.listener.Addr().Network() + "://" +
	tempAddress := rpcServer.Addr().String()
	remoteCosigner := NewRemoteCosigner(2, tempAddress)

	resp, err := remoteCosigner.GetEphemeralSecretPart(&CosignerGetEphemeralSecretPartRequest{})
	require.NoError(test, err)

	expctedRes := CosignerGetEphemeralSecretPartResponse{
		SourceID:                       1,
		SourceEphemeralSecretPublicKey: []byte("foo"),
		EncryptedSharePart:             []byte("bar"),
		SourceSig:                      []byte("source sig"),
	}

	require.Equal(test, expctedRes.SourceID, resp.SourceID)
	require.Equal(test, expctedRes.SourceEphemeralSecretPublicKey, resp.SourceEphemeralSecretPublicKey)
	require.Equal(test, expctedRes.EncryptedSharePart, resp.EncryptedSharePart)
	require.Equal(test, expctedRes.SourceSig, resp.SourceSig)

	rpcServer.Stop()
}

/*
func TestGRPCServer(test *testing.T) {

	var conn *grpc.ClientConn
	conn, err := grpc.Dial("127.0.0.1:1234", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("could not connect: %s", err)
	}
	defer conn.Close()

	c := NewCosignerServiceClient(conn)

	message := CosignerGetEphemeralSecretPartRequest{
		ID:     2,
		Height: 22,
		Round:  0,
		Step:   0,
	}

	response, err := c.GetEphemeralSecretPart(context.Background(), &message)
	if err != nil {
		fmt.Printf("Error when calling SayHello: %s", err)
	}

	fmt.Printf("Response from Server: %s", response.SourceEphemeralSecretPublicKey)

}
*/
