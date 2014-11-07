# Sysctl
module Sysctl
  class << self
    def config_file(node)
      if node['sysctl']['conf_dir']
        return node['sysctl']['conf_file']
      else
        if node['sysctl']['allow_sysctl_conf']
          return node['sysctl']['conf_file']
        else
          return nil
        end
      end
    end

    def compile_attr(prefix, v)
      case v
      when Array
        return "#{prefix}=#{v.join(' ')}"
      when String, Fixnum, Bignum, Float, Symbol
        "#{prefix}=#{v}"
      when Hash, Chef::Node::Attribute
        prefix += '.' unless prefix.empty?
        return v.map { |key, value| compile_attr("#{prefix}#{key}", value) }.flatten.sort
      else
        fail Chef::Exceptions::UnsupportedAction, "Sysctl cookbook can't handle values of type: #{v.class}"
      end
    end
  end
end
