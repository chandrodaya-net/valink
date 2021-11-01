#!/usr/bin/make -f

### Local validator nodes using docker and docker-compose
# setup config files of validators and nodes and create 
# the corresponding the docker-compose.yml
# example: make testnet-init nrValidator=3 nrNode=2 mode=mpc signer=3 total=4 theshold=2

# testnet-init:
# 	bash spin-up.sh setup_nodes $(nrValidator) $(nrNode) $(mode)

testnet-init:
	bash spin-up.sh setup_nodes 3 2 mpc 3 4 2


testnet-start:
	docker-compose up -d

testnet-stop:
	docker-compose down

testnet-clean:
	docker-compose down 
	bash spin-up.sh clean_setup