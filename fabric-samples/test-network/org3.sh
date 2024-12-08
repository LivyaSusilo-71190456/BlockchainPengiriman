#!/bin/bash

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE= ${PWD}/addOrg3/compose/docker/peercfg/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/addOrg3/compose/docker/peercfg/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:7051

