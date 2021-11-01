#!/usr/bin/make -f

define docker_restart_only_if_exited
	@if [ "$(docker ps -q  -f "status=exited" | wc -l)" != "" ] ; then \
			docker restart $(docker ps -q -f "status=exited") ;\
	fi
endef

testnet-restart-only-exited:
	$(call docker_restart_only_if_exited) 

# The following command setup config files of validators and nodes and create 
# the corresponding docker-compose.yml file
# example: make testnet-init nrValidator=4 nrNode=2 mode=mpc signer=3 total=4 theshold=2
testnet-init:
	bash spin-up.sh setup_nodes 4 2 mpc 3 4 2

# Some of the nodes/validator containers may exited prematurily at the container initialization
# because they don't find the pubkey.
# Usually this happened to the nodes/validator that are using an external signing process.
# if those nodes/validators start before the signing process, they won't find the pubkey and
# exit. To prevent those node from exit prematurily we simply execute docker_restart_only_if_exited
# in the testnet-start command.
testnet-start:
	docker-compose up -d
	$(call docker_restart_only_if_exited)
	
testnet-stop:
	docker-compose down

testnet-clean:
	docker-compose down 
	bash spin-up.sh clean_setup