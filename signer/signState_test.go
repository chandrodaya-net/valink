package signer

import (
	"testing"

	"github.com/stretchr/testify/require"
)



func TestLoadOrCreateSignState(test *testing.T) {
	state := SignState{}
	state.filePath = "../test/sign_state.json"
	state.Save()

	ssFile, err := LoadSignState(state.filePath)
	require.NoError(test,err)
	require.Equal(test, int64(0), ssFile.Height)
	require.Equal(test, int64(0), ssFile.Round)
	require.Equal(test, int8(0), ssFile.Step)
	require.Equal(test, []uint8([]byte(nil)), ssFile.EphemeralPublic)

}
