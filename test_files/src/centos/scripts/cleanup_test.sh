#!/usr/bin/env bash
yum install yum-utils
yum erase $(package-cleanup --leaves)
yum erase yum-utils
yum clean all
