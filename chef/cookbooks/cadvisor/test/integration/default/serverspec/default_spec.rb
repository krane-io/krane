# Encoding: utf-8

require_relative 'spec_helper'

describe 'default' do

  describe service('docker') do
    it { should be_enabled }
    it { should be_running }
  end

  describe port(8080) do
    it { should be_listening }
  end

end
