package signer

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"
)

type CosignerSeverMock struct {
	listenAddress string
	listener      net.Listener
}

func (csm *CosignerSeverMock) Sign(ctx context.Context, req *CosignerSignRequest) (*CosignerSignResponse, error) {
	return &CosignerSignResponse{Signature: []byte("hello world")}, nil
}

func (csm *CosignerSeverMock) GetEphemeralSecretPart(ctx context.Context, req *CosignerGetEphemeralSecretPartRequest) (*CosignerGetEphemeralSecretPartResponse, error) {
	response := &CosignerGetEphemeralSecretPartResponse{
		SourceID:                       1,
		SourceEphemeralSecretPublicKey: []byte("foo"),
		EncryptedSharePart:             []byte("bar"),
	}
	return response, nil
}

func TestRemoteCosignerSign(test *testing.T) {
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(test, err)
	defer lis.Close()

	rpcServer := &CosignerSeverMock{}
	rpcServer.listener = lis

	grpcServer := grpc.NewServer()

	RegisterCosignerServiceServer(grpcServer, rpcServer)

	go func() {
		defer lis.Close()
		//server.Serve(lis, tcpLogger, config)
		if err := grpcServer.Serve(lis); err != nil {
			//rpcServer.logger.Error("failed to serve", "error", err)
		}
	}()

	port := lis.Addr().(*net.TCPAddr).Port
	cosigner := NewRemoteCosigner(2, fmt.Sprintf("0.0.0.0:%d", port))

	resp, err := cosigner.Sign(&CosignerSignRequest{})
	require.NoError(test, err)
	require.Equal(test, resp.Signature, []byte("hello world"))
}

func TestRemoteCosignerGetEphemeralSecretPart(test *testing.T) {
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(test, err)
	defer lis.Close()

	rpcServer := &CosignerSeverMock{}
	rpcServer.listener = lis

	grpcServer := grpc.NewServer()

	RegisterCosignerServiceServer(grpcServer, rpcServer)

	go func() {
		defer lis.Close()
		//server.Serve(lis, tcpLogger, config)
		if err := grpcServer.Serve(lis); err != nil {
			//rpcServer.logger.Error("failed to serve", "error", err)
		}
	}()

	port := lis.Addr().(*net.TCPAddr).Port
	cosigner := NewRemoteCosigner(2, fmt.Sprintf("0.0.0.0:%d", port))

	resp, err := cosigner.GetEphemeralSecretPart(&CosignerGetEphemeralSecretPartRequest{})
	require.NoError(test, err)

	expectedRes := CosignerGetEphemeralSecretPartResponse{
		SourceID:                       1,
		SourceEphemeralSecretPublicKey: []byte("foo"),
		EncryptedSharePart:             []byte("bar"),
	}

	require.Equal(test, expectedRes.SourceID, resp.SourceID)
	require.Equal(test, expectedRes.SourceEphemeralSecretPublicKey, resp.SourceEphemeralSecretPublicKey)
	require.Equal(test, expectedRes.EncryptedSharePart, resp.EncryptedSharePart)

}
