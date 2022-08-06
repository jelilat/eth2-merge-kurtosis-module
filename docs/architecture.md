Module Architecture
===================
This repo is a Kurtosis module. To get general information on what a Kurtosis module is and how it works, visit [the modules documentation](https://docs.kurtosistech.com/modules.html).

The overview of this particular module's operation is as follows:

1. Parse user parameters
1. Launch a network of Ethereum participants
    1. Generate execution layer (EL) client config data
    1. Launch EL clients
    1. Wait for EL clients to start mining, such that all EL clients have a nonzero block number
    1. Generate consensus layer (CL) client config data
    1. Launch CL clients
1. Launch auxiliary services (Grafana, Forkmon, etc.)
1. Run Ethereum Merge verification logic
1. Return information to the user

Overview
--------
The module has six main components, in accordance with the above operation:

1. [Execute Function][execute-function]
1. [Module I/O][module-io]
1. [Static Files][static-files]
1. [Participant Network][participant-network]
1. Auxiliary Services
1. [Merge Verification Logic][testnet-verifier]

[Execute Function][execute-function]
------------------------------------
The execute function is the module's entrypoint/main function, where parameters are received from the user, lower-level calls are made, and a response is returned. Like all Kurtosis modules, this module receives serialized parameters and [an EnclaveContext object for manipulating the Kurtosis enclave][enclave-context], and returns a serialized response object.

[Module I/O][module-io]
-----------------------
This particular module has many configuration options (see the "Configuration" section in the README for the full list of values). These are passed in as a YAML-serialized string, and arrive to the module's execute function via the `serializedParams` variable. The process of setting defaults, overriding them with the user's desired options, and validating the resulting config object requires some space in the codebase. All this logic happens inside the `module_io` directory, so you'll want to visit this directory if you want to:

- View or change parameters that the module can receive
- Change the default values of module parameters
- View or change the validation logic that the module applies to configuration parameters
- View or change the properties that the module passes back to the user after execution is complete

[Static Files][static-files]
----------------------------
Kurtosis modules can have static files that are made available inside the container during the module's operation. For this module, [the static files included][static-files] are various key files and config templates which get used during participant network operation.

[Participant Network][participant-network]
------------------------------------------
The participant network is the beating heart at the center of the module. The participant network code is responsible for:

1. Generating EL client config data
1. Starting the EL clients
1. Waiting until the EL clients have started mining
1. Generating CL client config data
1. Starting the CL clients

We'll explain these phases one by one.

### Generating EL client data
All EL clients require both a genesis file and a JWT secret. The exact format of the genesis file differs per client, so we first leverage [a Docker image containing tools for generating this genesis data][ethereum-genesis-generator] to create the actual files that the EL clients-to-be will need. This is accomplished by filling in several genesis generation config files found in [the `static_files` directory][static-files].

The generated output files then get stored in the Kurtosis enclave, ready for use when we start the EL clients. The information about these stored files is tracked in [the `ELGenesisData` object](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/prelaunch_data_generator/el_genesis/el_genesis_data.go).

### Starting EL clients
Next, we plug the generated genesis data [into EL client "launchers"](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/participant_network/el) to start a mining network of EL nodes. The launchers are really just implementations of [the `ELClientLauncher` interface](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/el/el_client_launcher.go), with a `Launch` function that consumes EL genesis data and produces information about the running EL client node. Running EL node information is represented by [an `ELClientContext` struct](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/el/el_client_context.go). Each EL client type has its own launcher (e.g. [Geth](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/el/geth/geth_launcher.go), [Besu](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/el/besu/besu_launcher.go)) because each EL client will require different environment variables and flags to be set when launching the client's container.

### Waiting until EL clients have started mining
Once we have a network of EL nodes started, we block until all EL clients have a block number > 0 to ensure that they are in fact working. After the nodes have started mining, we're ready to move on to adding the CL client network.

### Generating CL client data
CL clients, like EL clients, also have genesis and config files that they need. We use [the same Docker image with tools for generating genesis data][ethereum-genesis-generator] to create the necessary CL files, we provide the genesis generation config using templates in [the `static_files` directory][static-files], and we store the generated output files in the Kurtosis enclave in the same way as the EL client genesis files. Like with EL nodes, CL genesis data information is tracked in [the `CLGenesisData` object](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/cl/cl_client_context.go).

The major difference with CL clients is that CL clients have validator keys, so we need to generate keys for the CL clients to use during validation. The generation happens using [the same genesis-generating Docker image][ethereum-genesis-generator], but via a call to [a different tool bundled inside the Docker image](https://github.com/protolambda/eth2-val-tools).

### Starting CL clients
Once CL genesis data and keys have been created, the CL client nodes are started via [the CL client launchers](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/participant_network/cl). Just as with EL clients:

- CL client launchers implement [the `CLClientLauncher` interface](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/cl/cl_client_launcher.go)
- One CL client launcher exists per client type (e.g. [Nimbus](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/cl/nimbus/nimbus_launcher.go), [Lighthouse](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/participant_network/cl/lighthouse))
- Launched CL node information is tracked in [a `CLClientContext` object](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/participant_network/cl/cl_client_context.go)

There are only two major difference between CL client and EL client launchers. First, the `CLClientLauncher.Launch` method also consumes an `ELClientContext`, because each CL client is connected in a 1:1 relationship with an EL client. Second, because CL clients have keys, the keystore files are passed in to the `Launch` function as well.

Auxiliary Services
------------------
After the Ethereum network is up and running, this module starts several auxiliary containers to make it easier to work with the Ethereum network. At time of writing, these are:

- [Forkmon](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/forkmon), a "fork monitor" web UI for visualizing the CL clients' forks
- [Prometheus](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/prometheus) for collecting client node metrics
- [Grafana](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/grafana) for visualizing client node metrics
- [An ETH transaction spammer](https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/transaction_spammer), which [has been forked off](https://github.com/kurtosis-tech/tx-fuzz) of [Marius' transaction spammer code](https://github.com/MariusVanDerWijden/tx-fuzz) so that it can run as a container

[Testnet Verifier][testnet-verifier]
------------------------------------
Once the Ethereum network is up and running, verification logic will be run to ensure that the Merge has happened successfully. This happens via [a testnet-verifying Docker image](https://github.com/ethereum/merge-testnet-verifier) that periodically polls the network to check the state of the merge. If the merge doesn't occur, the testnet-verifying image returns an unsuccessful exit code which in turn signals the Kurtosis module to exit with an error. This merge verification can be disabled in the module's configuration (see the "Configuration" section in the README).

<!------------------------ Only links below here -------------------------------->

[module-docs]: https://docs.kurtosistech.com/modules.html
[enclave-context]: https://docs.kurtosistech.com/kurtosis-core/lib-documentation#enclavecontext

[execute-function]: https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/blob/master/kurtosis-module/impl/module.go#L50
[module-io]: https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/module_io
[participant-network]: https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/participant_network
[ethereum-genesis-generator]: https://github.com/skylenet/ethereum-genesis-generator
[static-files]: https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/static_files
[testnet-verifier]: https://github.com/kurtosis-tech/eth2-merge-kurtosis-module/tree/master/kurtosis-module/impl/testnet_verifier
