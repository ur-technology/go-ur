#!/bin/bash -x

# creates the necessary docker images to run testrunner.sh locally

docker build --tag="ur/cppjit-testrunner" docker-cppjit
docker build --tag="ur/python-testrunner" docker-python
docker build --tag="ur/go-testrunner" docker-go
