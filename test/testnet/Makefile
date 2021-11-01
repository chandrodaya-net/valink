#!/usr/bin/make -f

VALIDATORS=4
NODES=3
MODE=mpc
SIGNERS=3
TOTALSHARES=4
THRESHOLD=2

INTEGRATIONTESTS = $(shell go list ./... | grep test)

define docker_restart_only_if_exited
	@if [ "$(docker ps -q  -f "status=exited" | wc -l)" != "" ] ; then \
			docker restart $(docker ps -q -f "status=exited") ;\
	fi
endef

testnet-restart-only-exited:
	$(call docker_restart_only_if_exited) 

# The following command setup config files for nodes/validators and create 
# the corresponding docker-compose.yml file
testnet-init:
	bash  spin-up.sh setup_nodes $(VALIDATORS) $(NODES) $(MODE) $(SIGNERS) $(TOTALSHARES) $(THRESHOLD)


testnet-start:
	docker-compose up -d
	echo "Sleep 25"
	@sleep 25

# stop all cointainers
testnet-stop:	
	docker-compose stop

# exited/running container will be restarted
testnet-restart:	
	docker-compose restart

# stop/remove all cointainers
testnet-down:
	docker-compose down

testnet-clean:
	docker-compose down 
	bash spin-up.sh clean_setup

testnet-check-mpc-validator:
	check-validators $(VALIDATORS)
	check-signers $(SIGNERS)

testnet-checks:
	go test -mod=readonly $(INTEGRATIONTESTS) -count=1

testnet-integration: testnet-init testnet-start testnet-checks testnet-clean
