# Encoding: utf-8

require_relative 'spec_helper'

describe 'cadvisor::default' do
  before { stub_resources }
  describe 'ubuntu' do
    let(:runner) { ChefSpec::Runner.new(::UBUNTU_OPTS) }
    let(:node) { runner.node }
    let(:chef_run) do
      runner.node.automatic['kernel']['release'] = '3.8.0'
      runner.node.set['docker']['alert_on_error_action'] = :warn
      runner.converge(described_recipe)
    end
  

    it 'installs docker by calling the docker cookbook' do
      expect(chef_run).to include_recipe('docker::default')
    end

    it 'downloads the cadvisor image' do
      expect(chef_run).to pull_docker_image('google/cadvisor').with(
        tag: 'latest'
      )
    end

    it 'runs the cadvisor container' do
      expect(chef_run).to run_docker_container('google/cadvisor').with(
        tag: 'latest',
        port: '127.0.0.1:8080:8080',
        volume: ['/var/run:/var/run:rw','/sys:/sys:ro','/var/lib/docker/:/var/lib/docker:ro'],
        container_name: 'cadvisor',
        detach: true
      )
    end      

  end
end
