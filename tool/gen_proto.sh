#!/bin/bash
protoc --proto_path=../schema/ --go_out=plugins=grpc:../pb ../schema/*.proto
