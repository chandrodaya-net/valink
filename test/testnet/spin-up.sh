#!/bin/sh

BINARY=junod
BINARY_IMAGE=cosmoscontracts/juno:latest
MPC_IMAGE=dautt/valink:v2.0.0
CHAINID=test-chain-id	
CHAINDIR=./chain/${BINARY::-1}/workspace	
MPCCHAINDIR=./chain/${BINARY::-1}/mpc
MPCKEYFOLDER=./chain/${BINARY::-1}/mpc/keys

VALIDATOR='validator'
NODE='node'

KBT="--keyring-backend=test"

echo "Creating $BINARY instance with home=$CHAINDIR chain-id=$CHAINID..."	

# Build genesis file incl account for passed address	
DENOM="ujuno"
MAXCOINS="100000000000"$DENOM
COINS="90000000000"$DENOM	

clean_setup(){
    echo "rm -rf $CHAINDIR"
    echo "rm -rf $MPCCHAINDIR"
    echo "rm -f docker-compose.yaml"
    sudo rm -rf $CHAINDIR
    sudo rm -f docker-compose.yml
    sudo rm -rf $MPCCHAINDIR
}

get_home() {
    dir="$CHAINDIR/$CHAINID/$1"
    echo $dir
}

# Initialize home directories
init_node_home () { 
    echo "init_node_home $1"
    home=$(get_home $1)
    $BINARY --home $home --chain-id $CHAINID init $1 &>/dev/null	
}

# Add some keys for funds
keys_add() {
    echo "keys_add $1"
    home=$(get_home $1)
    $BINARY --home $home keys add $1 $KBT &>> $home/config/account.txt	
}

# Add addresses to genesis
add_genesis_account() {
    echo "add_genesis_account $1"
    home=$(get_home $1)
    $BINARY --home $home add-genesis-account $($BINARY --home $home keys $KBT show $1 -a) $MAXCOINS  $KBT &>/dev/null	
}

# Create gentx file
gentx() {
    echo "gentx: $1"
    home=$(get_home $1)
    $BINARY --home $home gentx $1 $COINS  --chain-id $CHAINID $KBT &>/dev/null	
}

add_genesis_account_to_node0() {
    echo "add_genesis_account_to_node0: $1"
    home=$(get_home $1)
    home0=$(get_home ${VALIDATOR}0)
    $BINARY --home $home0 add-genesis-account $($BINARY --home $home keys $KBT show $1 -a) $MAXCOINS  $KBT &>/dev/null	
}

copy_all_gentx_and_add_genesis_account_to_node0(){
    echo "copy_all_gentx_and_add_genesis_account_to_node0"
    dir0=$(get_home ${VALIDATOR}0)
    n=1
    while ((n < $1)); do
        nodeName="${VALIDATOR}$n"
        dir=$(get_home $nodeName)
        cp $dir/config/gentx/*  $dir0/config/gentx 
        add_genesis_account_to_node0 $nodeName
        let n=n+1
    done
}

# create genesis file. node0 needs to execute this cmd
collect_gentxs_from_validator0(){
    echo "collect_gentxs_from_validator0"
    home=$(get_home ${VALIDATOR}0)
    $BINARY --home $home collect-gentxs &>/dev/null	
    echo "$home/config/genesis.json"
}


# $1 = number of node
# $2 = node type (VALIDATOR|NODE)
copy_genesis_json_from_node0_to_other_node(){
    echo "copy_genesis_json_from_node0_to_other_node"
    home0=$(get_home ${VALIDATOR}0)
    # throw an error if home0 does not exist
    n=0
    while ((n < $1)); do
        nodeName="$2$n"
        home=$(get_home $nodeName)
        cp $home0/config/genesis.json  $home/config/
        let n=n+1
    done
}

replace_stake_denomination(){
    echo "replace denomination in genesis: stake->$DENOM"
    home0=$(get_home ${VALIDATOR}0)
    sed -i "s/\"stake\"/\"$DENOM\"/g" $home0/config/genesis.json
}


# $1 = number of validators
# $2 = number of nodes
# $3 = the number of the current node
# $4 = current node type (VALIDATOR|NODE)
set_persistent_peers(){
    echo "set_persistent_peers $1 $2 $3 $4"
    currentNodeName="$4$3"
    currentNodeHome=$(get_home $currentNodeName)
    
    persistent_peers=""
    n=0
    ip_nr=0
    # validator loop
    while ((n < $1)); do
        nodeName="$VALIDATOR$n"
        ipAddress="192.168.10.$ip_nr"
        if [ "$n" != "$3"  ] || [ "$4" != "$VALIDATOR" ]; then
            home=$(get_home $nodeName)
            peer="$($BINARY --home $home tendermint show-node-id)@${ipAddress}:26656"
            if [ "$persistent_peers" != "" ]; then 
                persistent_peers=$persistent_peers","$peer ;
            else
                persistent_peers=$peer
            fi 
        fi 

        let n=n+1
        let ip_nr=ip_nr+1
    done

    # nodes loop
    let n=0
    ip_nr=$1
    while ((n < $2)); do
        nodeName="$NODE$n"
        ipAddress="192.168.10.$ip_nr"
        if [ "$n" != "$3" ] || [ "$4" != "$NODE" ]; then
            home=$(get_home $nodeName)
            peer="$($BINARY --home $home tendermint show-node-id)@${ipAddress}:26656"
            if [ "$persistent_peers" != "" ]; then 
                persistent_peers=$persistent_peers","$peer ;
            else
                persistent_peers=$peer
            fi 
        fi 

        let n=n+1
        let ip_nr=ip_nr+1
    done

   echo $currentNodeHome
   echo $persistent_peers
   sed -i "s/^persistent_peers *=.*/persistent_peers = \"$persistent_peers\"/" $currentNodeHome/config/config.toml
   
}
 
# $1 = number of validators
# $2 = number of nodes
set_persistent_peers_all_nodes() {
    echo "set_persistent_peers_all_nodes"
    node=0
    # validator
    while ((node < $1)); do
        set_persistent_peers $1 $2 $node $VALIDATOR 
        let node=node+1
    done

    let node=0
     # validator
    while ((node < $2)); do
        set_persistent_peers $1 $2 $node $NODE 
        let node=node+1
    done
}

# $1 = number of node
# $2 = node type (VALIDATOR|NODE)
init_node () {
    n=0
    while ((n < $1)); do
        nodeName="$2$n"
        echo "########## $nodeName ###############" 
        init_node_home $nodeName
        keys_add $nodeName
        if  [ "$2" = "$VALIDATOR" ]; then
            add_genesis_account $nodeName
            gentx $nodeName
        fi 
        let n=n+1
    done

    echo "########## generate genesis.json ###############"
     if  [ "$2" = "$VALIDATOR" ]; then
            copy_all_gentx_and_add_genesis_account_to_node0 $1 
            collect_gentxs_from_validator0
            replace_stake_denomination
     fi 
    copy_genesis_json_from_node0_to_other_node $1 $2
} 


# $1 = number validators
# $2 = number of nodes
# $3 = mpc/normal
# $4 = number of signer
generate_docker_compose_file(){
    echo -e "version: '3'\n"
    echo -e "services:"

    n=0
    portStart=26656
    portEnd=26657
    mpcPort=1235
    # validator config
    while ((n < $1)); do
        nodeName="$VALIDATOR$n"
        echo " $nodeName:"
        echo "   container_name: $nodeName"
        echo "   image: $BINARY_IMAGE"
        echo "   ports:"
        echo "   - \"$portStart-$portEnd:26656-26657\""
        echo "   - \"$mpcPort:1235\""
        echo "   volumes:"
        echo "   - $CHAINDIR:/workspace"
        echo "   command: /bin/sh -c 'junod start --home /workspace/test-chain-id/$nodeName'"
        echo "   networks:"
        echo "     localnet:"
        echo -e "       ipv4_address: 192.168.10.$n\n"
    
        let n=n+1
        let portStart=portEnd+1
        let portEnd=portStart+1
        let mpcPort=mpcPort+1
    done

    # node config
    let n=0
    ip_nr=$1
    while ((n < $2)); do
        nodeName="$NODE$n"
       
        echo " $nodeName:"
        echo "   container_name: $nodeName"
        echo "   image: $BINARY_IMAGE"
        echo "   ports:"
        echo "   - \"$portStart-$portEnd:26656-26657\""
        echo "   - \"$mpcPort:1235\""
        echo "   volumes:"
        echo "   - $CHAINDIR:/workspace"
        echo "   command: /bin/sh -c 'junod start --home /workspace/test-chain-id/$nodeName'"
        echo "   networks:"
        echo "     localnet:"
        echo -e "       ipv4_address: 192.168.10.$ip_nr\n"
    
        let n=n+1
        let ip_nr=ip_nr+1
        let portStart=portEnd+1
        let portEnd=portStart+1
        let mpcPort=mpcPort+1
    done

    if [ "$3" = "mpc" ]; then 
         # mpc signer section   
         let n=1
         while ((n <= $4)); do
            mpcName="mpc$n"
            echo " $mpcName:"
            echo "   container_name: $mpcName"
            echo "   image: $MPC_IMAGE"
            echo "   ports:"
            echo "   - \"$mpcPort:1234\""
            echo "   volumes:"
            echo "   - $MPCCHAINDIR:/mpc"
            echo "   command: /bin/sh -c 'valink cosigner start /mpc/$mpcName/config.toml'"
            echo "   networks:"
            echo "     localnet:"
            echo -e "       ipv4_address: 192.168.10.$ip_nr\n"
        
            let n=n+1
            let mpcPort=mpcPort+1
            let ip_nr=ip_nr+1
         done
    
    fi 


    echo "networks:"
    echo "  localnet:"
    echo "    driver: bridge"
    echo "    ipam:"
    echo "      driver: default"
    echo "      config:"
    echo "      -"
    echo "        subnet: 192.168.10.0/16"

}


# $1 = number validator
# $2 = number node
# $3 = number signer
# $4 = mpc total shares
# $5 = mpc threshold
# $6 = number of the current signer
generate_mpc_config_file(){
    mpcName="mpc$6" 
    echo -e "mode = \"mpc\"\n"

    echo "# Each validator instance has its own private share."
    echo "# Avoid putting more than one share per instance."
    echo -e "key_file = \"/mpc/keys/private_share_$6.json\"\n"

    echo "# The state directory stores watermarks for double signing protection."
    echo "# Each validator instance maintains a watermark."
    echo -e "state_dir = \"/mpc/$mpcName\"\n"

    echo "# The network chain id for your p2p nodes"
    echo -e "chain_id = \"${CHAINID}\"\n"

    echo "# The required number of participant share signatures."
    echo "# This must match the \`threshold\` value specified during key2shares"
    echo -e "cosigner_threshold = $5\n"

    echo "# IP address and port for receiving communication from other validator instances."
    echo "# The validator instances must communicate during the signing process."
    echo -e "cosigner_listen_address = \"tcp://0.0.0.0:1234\"\n"

    echo "# Each validator peer appears in a \`cosigner\` section."
    echo "# This sample file is for validator ID 1, so we configure sections for peers 2 and 3."

    k=1
    let ip_nr=$1+$2
    while ((k <= $3)); do
        ipAddress="192.168.10.$ip_nr"
        if [ "$k" != "$6" ] ; then
            echo "[[cosigner]]"
            echo "# The ID of this peer, these must match the key IDs."
            echo "id = $k"
            echo "# The IP address and port for communication with this peer"
            echo -e "remote_address = \"tcp://${ipAddress}:1234\"\n"
        fi 

        let k=k+1
        let ip_nr=ip_nr+1
    done

    # mpc1 is assigned to validator0
    # mpc2 is assigned to node0
    # mpc3 is assigned to node1
    # mpcn is assigned to node(n-2)
    let ip_nr=$1-2+$6
    ipAddress="192.168.10.$ip_nr"
    if [ "$6" = "1" ] ; then
                # ip address of validator0, which is by convention the validator used to build the
                # threshold validator
                ipAddress="192.168.10.0"
    fi
    echo "# Configure any number of p2p network nodes."
    echo "# We recommend at least 2 nodes per cosigner for redundancy."
    echo "[[node]]"
    echo "address = \"tcp://${ipAddress}:1235\""

    #[[node]]
    #address = "tcp://137.184.11.214:1235"
     
}


# # $1 = number validator
# # $2 = number node
# # $3 = number signer
# # $4 = mpc total shares
# # $5 = mpc threshold
# # $6 = number of the current signer
# generate_mpc_config_file(){
 
# }

generate_mpc_sign_state(){
    echo   "{
            \"height\": \"0\",
            \"round\": \"0\",
            \"step\": 0,
            \"ephemeral_public\": null
            }"
}

# $1 = number validator
# $2 = number node
# $3 = number signer
# $4 = mpc total shares
# $5 = mpc threshold
generate_all_mpc_config_file() {
    echo "generate_all_mpc_config_file"
    n=1
    while ((n <= $3)); do
       if [ ! -f $MPCCHAINDIR/mpc${n} ]; then 
            mkdir -p $MPCCHAINDIR/mpc${n}
       fi
       generate_mpc_config_file $1 $2 $3 $4 $5 $n &> $MPCCHAINDIR/mpc${n}/config.toml
       # typically this file is generated manually to prevent double signing
       generate_mpc_sign_state &> $MPCCHAINDIR/mpc${n}/${CHAINID}_share_sign_state.json
       let n=n+1
    done
}


create_mpc_share_from_validator0(){
    if [ ! -f $MPCKEYFOLDER ]; then 
            mkdir -p $MPCKEYFOLDER
    fi
    cd $MPCKEYFOLDER
    cp ../../workspace/test-chain-id/validator0/config/priv_validator_key.json ./
    ../../../../../../build/valink create-shares  ../../workspace/test-chain-id/validator0/config/priv_validator_key.json 2 4 
    
    # go back to folder where spin-up.sh is located
    cd ../../../../
}

# $1 number of node
# $2 number mpc signer
set_all_mpc_priv_validator_laddr(){
echo "set_all_mpc_priv_validator_laddr"
    n=1
    while ((n <= $2)); do
        let mpcnodeNr=n-2
        nodeName="${NODE}$mpcnodeNr"
        if [ "$n" = "1" ] ; then
                nodeName="${VALIDATOR}0"
        fi
        nodeHome=$(get_home $nodeName)        
        sed -i "s/^priv_validator_laddr *=.*/priv_validator_laddr = \"tcp:\/\/0.0.0.0:1235\"/" $nodeHome/config/config.toml
        let n=n+1
    done
}

   

# $1 = number validators
# $2 = number node
# $3 = mpc/normal
# $4 = number signer
# $5 = mpc total shares
# $6 = mpc threshold
setup_nodes(){
    clean_setup
    init_node $1 $VALIDATOR
    init_node $2 $NODE
    set_persistent_peers_all_nodes $1 $2
    echo "generate_docker_compose_file $1 $2 $3 $4"
    generate_docker_compose_file $1 $2 $3 $4 > docker-compose.yml
    generate_all_mpc_config_file $1 $2 $4 $5 $6
    set_all_mpc_priv_validator_laddr $2 $4
    create_mpc_share_from_validator0
}


repl() {
PS3='Please enter your choice: '
options=("setup nodes"  "init validator" "init nodes" "clean setup" "docker compose file"  "mpc config files" "create mpc shares" "Quit")
select opt in "${options[@]}"
do
    case $opt in
        "setup nodes")
            read -p "number of validators: " valNr
            read -p "number of nodes: " nodeNr
            read -p "mode (mpc/normal): " mode
            read -p "mpc number signer: " mpcsignerNr
            read -p "mpc total shares: " mpctotal
            read -p "mpc threshold: " mpcthreshold

            setup_nodes $valNr $nodeNr $mode $mpcsignerNr $mpctotal $mpcthreshold
            ;;
        "init validator")
            read -p "number of validators: " valNr
            init_node $valNr $VALIDATOR
            set_persistent_peers_all_nodes $valNr 0
            ;;
        "init nodes")
            read -p "number of nodes: " nodeNr
            init_node $nodeNr $NODE
            set_persistent_peers_all_nodes 0 $nodeNr
            ;;
        "clean setup")
           clean_setup
            ;;
        "docker compose file")
            read -p "number of validators: " valNr
            read -p "number of nodes: " nodeNr
            read -p "mode (mpc/normal): " mode
            read -p "mpc number signer: " mpcsignerNr

            generate_docker_compose_file $valNr $nodeNr $mode $mpcsignerNr &> docker-compose.yml
            ;;
        "mpc config files")
            read -p "number validator: " valNr
            read -p "number of nodes: " nodeNr
            read -p "number of signer: " signerNr
            read -p "total shares: " total
            read -p "threshold: " threshold

            if (( $signerNr < ($nodeNr-1) )); then 
                echo "Error: number of signer ($signerNr) < number of nodes ($nodeNr)" 1>&2
                exit 1
            fi 

            generate_all_mpc_config_file $valNr $nodeNr $signerNr $total $threshold
            ;;
        "create mpc shares")
            create_mpc_share_from_validator0
            ;;
        "Quit")
            break
            ;;
        *) echo "invalid option $REPLY";;
    esac
done

}

"$@"
