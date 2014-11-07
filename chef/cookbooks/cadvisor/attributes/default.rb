# Encoding: utf-8
#
# Cookbook Name:: cadvisor
# Attributes:: default
#
# Copyright 2014, Paul Czarkowski
#

# Image for Cadvisor
default[:cadvisor][:image] = 'google/cadvisor'

# Version/Tag of Cadvisor to run
default[:cadvisor][:version] = 'latest'

# Name to give the running container
default[:cadvisor][:container] = 'cadvisor'

# port and ip to expose container's port 8080 on
default[:cadvisor][:listen_ip] = '127.0.0.1'
default[:cadvisor][:listen_port] = '8080'
