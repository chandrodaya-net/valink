package signer

import (
	"context"
	"errors"

	client "github.com/tendermint/tendermint/rpc/jsonrpc/client"
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
func (cosigner *RemoteCosigner) Sign(signReq CosignerSignRequest) (CosignerSignResponse, error) {
	logger.Info("START: Sign", "from RemoteCosigner=", cosigner.GetID())

	params := map[string]interface{}{
		"arg": CosignerSignRequest{
			SignBytes: signReq.SignBytes,
		},
	}

	remoteClient, err := client.New(cosigner.address)
	if err != nil {
		return CosignerSignResponse{}, err
	}
	result := &CosignerSignResponse{}
	_, err = remoteClient.Call(ctx, "Sign", params, result)
	if err != nil {
		return CosignerSignResponse{}, err
	}

	logger.Info("END: Sign", "from RemoteCosigner=", cosigner.GetID())
	return CosignerSignResponse{
		Timestamp: result.Timestamp,
		Signature: result.Signature,
	}, nil
}

func (cosigner *RemoteCosigner) GetEphemeralSecretPart(req CosignerGetEphemeralSecretPartRequest) (CosignerGetEphemeralSecretPartResponse, error) {
	logger.Info("START: GetEphemeralSecretPart", "from RemoteCosigner=", cosigner.GetID())
	resp := CosignerGetEphemeralSecretPartResponse{}

	params := map[string]interface{}{
		"arg": CosignerGetEphemeralSecretPartRequest{
			ID:     req.ID,
			Height: req.Height,
			Round:  req.Round,
			Step:   req.Step,
		},
	}

	remoteClient, err := client.New(cosigner.address)
	if err != nil {
		return CosignerGetEphemeralSecretPartResponse{}, err
	}
	result := &CosignerGetEphemeralSecretPartResponse{}
	_, err = remoteClient.Call(ctx, "GetEphemeralSecretPart", params, result)
	if err != nil {
		return CosignerGetEphemeralSecretPartResponse{}, err
	}

	resp.SourceID = result.SourceID
	resp.SourceEphemeralSecretPublicKey = result.SourceEphemeralSecretPublicKey
	resp.EncryptedSharePart = result.EncryptedSharePart
	resp.SourceSig = result.SourceSig

	logger.Info("END: GetEphemeralSecretPart", "from RemoteCosigner=", cosigner.GetID(), "Response.SourceID=", result.SourceID)
	return resp, nil
}

func (cosigner *RemoteCosigner) HasEphemeralSecretPart(req CosignerHasEphemeralSecretPartRequest) (CosignerHasEphemeralSecretPartResponse, error) {
	res := CosignerHasEphemeralSecretPartResponse{}
	return res, errors.New("Not Implemented")
}

func (cosigner *RemoteCosigner) SetEphemeralSecretPart(req CosignerSetEphemeralSecretPartRequest) error {
	return errors.New("Not Implemented")
}
