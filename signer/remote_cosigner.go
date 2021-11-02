package signer

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
)

var (
	ctx = context.Background()
)

// RemoteCosigner uses tendermint rpc to request signing from a remote cosigner
type RemoteCosigner struct {
	id      int
	address string
}

// NewRemoteCosigner returns a newly initialized RemoteCosigner
func NewRemoteCosigner(id int, address string) *RemoteCosigner {
	cosigner := &RemoteCosigner{
		id:      id,
		address: address,
	}
	return cosigner
}

// GetID returns the ID of the remote cosigner
// Implements the cosigner interface
func (cosigner *RemoteCosigner) GetID() int {
	return cosigner.id
}

// Sign the sign request using the cosigner's share
// Return the signed bytes or an error
func (cosigner *RemoteCosigner) Sign(signReq *CosignerSignRequest) (*CosignerSignResponse, error) {
	//logger.Info("RemoteCosigner Sign", "RemoteCosignerID", cosigner.GetID())
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(cosigner.address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("could not connect: %s", err)
		return &CosignerSignResponse{}, err
	}
	defer conn.Close()

	c := NewCosignerServiceClient(conn)

	response, err := c.Sign(context.Background(), signReq)
	if err != nil {
		return &CosignerSignResponse{}, err
	}

	return response, nil
}

func (cosigner *RemoteCosigner) GetEphemeralSecretPart(req *CosignerGetEphemeralSecretPartRequest) (*CosignerGetEphemeralSecretPartResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(cosigner.address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("could not connect: %s", err)
		return &CosignerGetEphemeralSecretPartResponse{}, err
	}
	defer conn.Close()

	c := NewCosignerServiceClient(conn)

	response, err := c.GetEphemeralSecretPart(context.Background(), req)
	if err != nil {
		return &CosignerGetEphemeralSecretPartResponse{}, err
	}

	return response, nil
}

func (cosigner *RemoteCosigner) HasEphemeralSecretPart(req CosignerHasEphemeralSecretPartRequest) (CosignerHasEphemeralSecretPartResponse, error) {
	res := CosignerHasEphemeralSecretPartResponse{}
	return res, errors.New("Not Implemented")
}

func (cosigner *RemoteCosigner) SetEphemeralSecretPart(req CosignerSetEphemeralSecretPartRequest) error {
	return errors.New("Not Implemented")
}
