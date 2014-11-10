
#Krane [![Gobuild Download](http://gobuild.io/badge/github.com/krane-io/krane/downloads.svg)](http://gobuild.io/github.com/krane-io/krane)
What is Krane?
---

Krane is an open-source platform that enables developers and sysadmins to manage Docker instances across multiple clouds. Krane is built on the Docker code base, so supports the existing Docker command-line in a multi-cloud environment. This allows Docker users to use their existing workflows to launch apps transparently in multiple clouds, saving time and further reducing the friction of moving workloads between development, QA and production.

What are Krane ships?
---

In Krane, a ship is a virtual machine (VM) running a Docker instance into which Krane will place Docker containers. Krane (running on the developer's command line) communicates with Docker running on the ship via its API (you can run Krane on the ship if you like, but it's not necessary). Using Krane, you can execute your Docker command (such as starting or stopping a Docker container) on any ship. Ships are built as standard with a cadviser container, which monitors performance of the ship and its containers.

What is a Krane shipyard?
---

A Krane shipyard is a driver for Krane that is capable of building, destroying and controlling ships. We've initially provided two drivers. Firstly, a driver for AWS. Secondly, a driver for Flexiant Concerto, a multi-cloud manager that supports AWS, Digital Ocean, Rackspace, Joyent and of course Flexiant Cloud Orchestrator.

If you want to use the driver for Flexiant Concerto, you can get a free account by visiting https://signup.krane.io/; you'll need the code as well of course. If you want to use the AWS driver, you just need the code (see below). In either case, you will need to provide the credentials for your cloud providers. We would welcome contribution of additional drivers.

Where can I get the code? What licence is used?
---

The code is on Github at https://github.com/krane-io/krane. It's written in Go, and is based on the existing Docker source code, so we've released it under the Apache 2.0 licence. We hope to get it merged into upstream Docker.

Where can I download a binary of Krane
---

Visit the following page http://gobuild.io/github.com/krane-io/krane to download a compiled version of Krane.

Who is behind Krane?
---

Krane is an open-source project started by Flexiant Ltd (http://www.flexiant.com). We welcome contributions from elsewhere.
