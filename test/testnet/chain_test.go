package chain_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	TMVersion   = "v0.34.13"
	VotingPower = "90000"
)

func TestChain(t *testing.T) {
	testServices := getDockerComposeServices()
	for _, ts := range testServices {
		t.Run(ts.ContainerName, func(t *testing.T) {
			if ts.ContainerName[0:3] == "mpc" {
				checkMPCNodeReachable(t, ts)
			} else {
				countServiceByType := map[string]int{
					"mpc":       countService("mpc", testServices),
					"node":      countService("node", testServices),
					"validator": countService("validator", testServices),
				}
				checkNodeAndValidatorStatus(t, ts, countServiceByType)
			}
		})
	}
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func TestThresholdValidatorStatus(t *testing.T) {
	testServices := getDockerComposeServices()
	// number of mpc signers
	nrMPC := countService("mpc", testServices)
	mpcIndexes := makeRange(1, nrMPC)
	// thresholdvalidator status
	tvStatus := getResponseStatus(t, testServices["validator0"])
	tvAddress := tvStatus.Result.ValidatorInfo.Address
	require.Equal(t, VotingPower, tvStatus.Result.ValidatorInfo.VotingPower)

	for i := range mpcIndexes {
		currentMPCsigner := fmt.Sprintf("mpc%d", mpcIndexes[i])
		checkMPCNodeReachable(t, testServices[currentMPCsigner])
		if i < nrMPC-1 {
			currentMPCnode := fmt.Sprintf("node%d", i)
			currentMPCnodeStatus := getResponseStatus(t, testServices[currentMPCnode])
			currentMPCNodeAddress := currentMPCnodeStatus.Result.ValidatorInfo.Address
			require.Equal(t, tvAddress, currentMPCNodeAddress, fmt.Sprintf("%s address: %s", currentMPCnode, currentMPCNodeAddress))
		}
	}

}

func TestTVSignedBlockStatistics(t *testing.T) {
	testServices := getDockerComposeServices()
	tvStatus := getResponseStatus(t, testServices["validator0"]) // threshold validator status
	tvAddress := tvStatus.Result.ValidatorInfo.Address

	expectedHeight := 5
	actualHeight := 0
	// wait till the blockchain has reach at least height 20
	for {
		tvStatus = getResponseStatus(t, testServices["validator0"])
		var err error
		actualHeight, err = strconv.Atoi(tvStatus.Result.SyncInfo.LatestBlockHeight)
		require.NoError(t, err)
		fmt.Println("Threshold validator status:")
		fmt.Println(fmt.Sprintf("height:%d, VotingPower:%s", actualHeight, tvStatus.Result.ValidatorInfo.VotingPower))
		if actualHeight > expectedHeight {
			break
		}
	}

	// statitics contains the statistics of the validator signing block up to actualHeight
	statistics := checkBlocks(t, testServices, actualHeight)
	fmt.Println(fmt.Sprintf("**** actualHeight=%d ****", actualHeight))
	for val := range statistics {
		fmt.Println(fmt.Sprintf(`**** %s:  
	countUnsignerBlock: %d (%v),
	countsignedBlocks: %d (%v), 
	percentage signed: %f
		*****`, val, len(statistics[val]["unsignedBlock"]), statistics[val]["unsignedBlock"],
			len(statistics[val]["signedBlock"]), statistics[val]["signedBlock"], float64(len(statistics[val]["signedBlock"]))/float64(actualHeight)))
	}

	// check threshold validator
	tval := fmt.Sprintf("validator0-%s", tvAddress)
	require.LessOrEqual(t, len(statistics[tval]["unsignedBlock"]), 1, fmt.Sprintf("****countMissingBlock: expected (1) got %d *****",
		len(statistics[tval]["unsignedBlock"])))

	//rpc call to check signature at a particular height (101): http://localhost:26657/commit?height=101

}

func initializeStatisticsMap(t *testing.T) map[string]map[string][]int {
	statistics := make(map[string]map[string][]int)
	addressValidatorMap := reverseKV(getValidatorsAddress(t))

	for a := range addressValidatorMap {
		key := fmt.Sprintf("%s-%s", addressValidatorMap[a], a)
		statistics[key] = make(map[string][]int)
	}
	return statistics
}

func checkBlocks(t *testing.T, services map[string]service, lastHeight int) (statistics map[string]map[string][]int) {
	r := makeRange(1, lastHeight)
	statistics = initializeStatisticsMap(t)
	//addressValidatorMap := reverseKV(getValidatorsAddress(t))
	for s := range r {
		statistics = checkBlock(t, statistics, services, s)
	}
	return statistics
}

func checkBlock(t *testing.T, statistics map[string]map[string][]int, services map[string]service, height int) map[string]map[string][]int {
	addressValidatorMap := reverseKV(getValidatorsAddress(t))
	tv := services["validator0"]
	ResponseCommitment := getResponseCommitment(t, tv, height)
	signatures := ResponseCommitment.Result.SignedHeader.Commit.Signatures
	var signedKey []string
	// tracked validator signed block at height
	for s := range signatures {
		sva := signatures[s].ValidatorAddress
		if sva != "" {
			validator := addressValidatorMap[sva]
			key := fmt.Sprintf("%s-%s", validator, sva)
			signedKey = append(signedKey, key)
			statistics[key]["signedBlock"] = append(statistics[key]["signedBlock"], height)
		}
	}

	// tracked validator unsigned block at height
	for a := range addressValidatorMap {
		key := fmt.Sprintf("%s-%s", addressValidatorMap[a], a)
		unsignedKey := true
		for _, k := range signedKey {
			if k == key {
				unsignedKey = false
			}
		}
		if unsignedKey {
			statistics[key]["unsignedBlock"] = append(statistics[key]["unsignedBlock"], height)
		}

	}

	return statistics
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getValidatorsAddress(t *testing.T) (result map[string]string) {
	testServices := getDockerComposeServices()
	result = make(map[string]string)
	for s := range testServices {
		if strings.Contains(s, "validator") {
			vStatus := getResponseStatus(t, testServices[s])
			result[s] = vStatus.Result.ValidatorInfo.Address
		}
	}
	return result
}

func reverseKV(m map[string]string) (result map[string]string) {
	result = make(map[string]string)
	for k := range m {
		result[m[k]] = k
	}
	return result
}

func getThresholdValidatorInfo(t *testing.T) status {
	return getValidatorInfo(t, "validator0")
}

func getValidatorInfo(t *testing.T, validatorName string) status {
	testServices := getDockerComposeServices()
	v := testServices[validatorName]
	vStatus := getResponseStatus(t, v)
	return vStatus
}

func getDockerComposeServices() map[string]service {
	filename := "./docker-compose.yml" //filepath.Abs("./test/testnet/docker-compose.yml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	// var config dockerCompose
	var config dockerCompose
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	return config.Services
}

func checkNodeAndValidatorStatus(t *testing.T, s service, countServiceByType map[string]int) {
	result := getResponseStatus(t, s)
	require.Equal(t, s.ContainerName, result.Result.NodeInfo.Moniker)
	require.Equal(t, TMVersion, result.Result.NodeInfo.Version)

	mpcNr := countServiceByType["mpc"]
	// validatorNr := countServiceByType["validator"]
	// nodeNr := countServiceByType["node"]

	// mpc are assigned by convention to: validator0, node0, node1 and so on which constitute the threshold/mpc validator
	// let say mpcNr=3, validatorNr=4, nodeNr=3:
	// In this case the threshold validator is: validator0, node0, node1. All of them are seeing by the network as one entity.
	// This means the voting power of validator0, node0, node1 is by default 9000
	// voting power of validator1, validator2, validator3 is 9000
	// voting power of node2 is 0
	if s.ContainerName[0:4] == "node" {
		currentNodePosition, err := strconv.Atoi(s.ContainerName[4:])
		require.NoError(t, err, fmt.Sprintf("unable find position of %s", s.ContainerName))
		// we subtract 1 to mpcNr to find the number of nodes which are assigned to threshold/mpc validator
		// because already validator0 is assigned to the threshold/mpc validator
		if currentNodePosition >= mpcNr-1 {
			require.Equal(t, "0", result.Result.ValidatorInfo.VotingPower)
		} else {
			tv := getThresholdValidatorInfo(t)
			// the node assigned to the threshold validator will have voting power = 9000
			require.Equal(t, tv.Result.ValidatorInfo.VotingPower, result.Result.ValidatorInfo.VotingPower)
		}
	} else {
		require.Equal(t, VotingPower, result.Result.ValidatorInfo.VotingPower)
	}
	// result.Result.SyncInfo.LatestBlockHeight
	// result.Result.ValidatorInfo.Address
	// result.Result.ValidatorInfo.PubKey

}

func getResponseCommitment(t *testing.T, s service, height int) ResponseCommitment {
	ports := strings.Split(s.Ports[1], ":")
	url := fmt.Sprintf("http://127.0.0.1:%s/commit?height=%d", ports[0], height)
	resp, err := http.Get(url)
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

	var result ResponseCommitment
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Can not unmarshal JSON")
	return result
}

func getResponseStatus(t *testing.T, s service) status {
	ports := strings.Split(s.Ports[1], ":")
	url := fmt.Sprintf("http://127.0.0.1:%s/status?", ports[0])
	resp, err := http.Get(url)
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

	var result status
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Can not unmarshal JSON")
	return result
}

func countService(serviceType string, services map[string]service) int {
	count := 0
	for k := range services {
		if strings.Contains(k, serviceType) {
			count++
		}
	}
	return count
}

func checkMPCNodeReachable(t *testing.T, s service) {
	ports := strings.Split(s.Ports[0], ":")
	url := fmt.Sprintf("http://127.0.0.1:%s", ports[0])
	resp, err := http.Get(url)
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte
	require.NoError(t, err)

	result := fmt.Sprintf("%s", body)

	require.Contains(t, result, "GetEphemeralSecretPart?")
	require.Contains(t, result, "Sign?")

}

type status struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		NodeInfo struct {
			ProtocolVersion struct {
				P2P   string `json:"p2p"`
				Block string `json:"block"`
				App   string `json:"app"`
			} `json:"protocol_version"`
			ID         string `json:"id"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			Channels   string `json:"channels"`
			Moniker    string `json:"moniker"`
			Other      struct {
				TxIndex    string `json:"tx_index"`
				RPCAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string    `json:"latest_block_hash"`
			LatestAppHash       string    `json:"latest_app_hash"`
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHash   string    `json:"earliest_block_hash"`
			EarliestAppHash     string    `json:"earliest_app_hash"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			EarliestBlockTime   time.Time `json:"earliest_block_time"`
			CatchingUp          bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}

// PrettyPrint to print struct in a readable way
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

type service struct {
	ContainerName string   `yaml:"container_name"`
	Image         string   `yaml:"image"`
	Ports         []string `yaml:"ports"`
	Volumes       []string `yaml:"volumes"`
	Command       string   `yaml:"command"`
	Networks      struct {
		Localnet struct {
			Ipv4Address string `yaml:"ipv4_address"`
		} `yaml:"localnet"`
	} `yaml:"networks"`
}

type dockerCompose struct {
	Version  string             `yaml:"version"`
	Services map[string]service `yaml:"services"`
	Networks struct {
		Localnet struct {
			Driver string `yaml:"driver"`
			Ipam   struct {
				Driver string `yaml:"driver"`
				Config []struct {
					Subnet string `yaml:"subnet"`
				} `yaml:"config"`
			} `yaml:"ipam"`
		} `yaml:"localnet"`
	} `yaml:"networks"`
}

type ResponseCommitment struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		SignedHeader struct {
			Header struct {
				Version struct {
					Block string `json:"block"`
				} `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    string `json:"last_results_hash"`
				EvidenceHash       string `json:"evidence_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
			Commit struct {
				Height  string `json:"height"`
				Round   int    `json:"round"`
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Signatures []struct {
					BlockIDFlag      int       `json:"block_id_flag"`
					ValidatorAddress string    `json:"validator_address"`
					Timestamp        time.Time `json:"timestamp"`
					Signature        string    `json:"signature"`
				} `json:"signatures"`
			} `json:"commit"`
		} `json:"signed_header"`
		Canonical bool `json:"canonical"`
	} `json:"result"`
}
