# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"


Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

    config.vm.box = "phusion/ubuntu-14.04-amd64" # https://github.com/phusion/open-vagrant-boxes

    config.vm.define "incoming-demo" do |node|
        node.vm.network :private_network, :ip => '10.20.1.4'
        node.vm.provision "ansible" do |ansible|
            ansible.groups = {
                "build" => ["incoming-demo"],
                "test" => ["incoming-demo"]
            }
            ansible.extra_vars = {
                host_build_user: "vagrant",
                host_fqdn: "10.20.1.4"
            }
            ansible.playbook = "ansible/everything_in_vagrant_box.yml"
        end
    end
end
