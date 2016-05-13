## UR Go

Official golang implementation of the UR protocol

          | Linux   | OSX | ARM | Windows | Tests
----------|---------|-----|-----|---------|------
develop   | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=Linux%20Go%20develop%20branch)](https://build.urdev.com/builders/Linux%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=Linux%20Go%20develop%20branch)](https://build.urdev.com/builders/OSX%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=ARM%20Go%20develop%20branch)](https://build.urdev.com/builders/ARM%20Go%20develop%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=Windows%20Go%20develop%20branch)](https://build.urdev.com/builders/Windows%20Go%20develop%20branch/builds/-1) | [![Buildr+Status](https://travis-ci.org/ur/go-ur.svg?branch=develop)](https://travis-ci.org/ur/go-ur) [![codecov.io](http://codecov.io/github/ur/go-ur/coverage.svg?branch=develop)](http://codecov.io/github/ur/go-ur?branch=develop)
master    | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=Linux%20Go%20master%20branch)](https://build.urdev.com/builders/Linux%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=OSX%20Go%20master%20branch)](https://build.urdev.com/builders/OSX%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=ARM%20Go%20master%20branch)](https://build.urdev.com/builders/ARM%20Go%20master%20branch/builds/-1) | [![Build+Status](https://build.urdev.com/buildstatusimage?builder=Windows%20Go%20master%20branch)](https://build.urdev.com/builders/Windows%20Go%20master%20branch/builds/-1) | [![Buildr+Status](https://travis-ci.org/ur/go-ur.svg?branch=master)](https://travis-ci.org/ur/go-ur) [![codecov.io](http://codecov.io/github/ur/go-ur/coverage.svg?branch=master)](http://codecov.io/github/ur/go-ur?branch=master)

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/ur/go-ur) 
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ur/go-ur?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

## Automated development builds

The following builds are build automatically by our build servers after each push to the [develop](https://github.com/ur/go-ur/tree/develop) branch.

* [Docker](https://registry.hub.docker.com/u/ur/client-go/)
* [OS X](http://build.urdev.com/builds/OSX%20Go%20develop%20branch/Mist-OSX-latest.dmg)
* Ubuntu
  [trusty](https://build.urdev.com/builds/Linux%20Go%20develop%20deb%20i386-trusty/latest/) |
  [utopic](https://build.urdev.com/builds/Linux%20Go%20develop%20deb%20i386-utopic/latest/)
* [Windows 64-bit](https://build.urdev.com/builds/Windows%20Go%20develop%20branch/Gur-Win64-latest.zip)
* [ARM](https://build.urdev.com/builds/ARM%20Go%20develop%20branch/gur-ARM-latest.tar.bz2)

## Building the source

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/ur/go-ur/wiki/Building-UR)
on the wiki.

Building gur requires both a Go and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make gur

## Executables

Go UR comes with several wrappers/executables found in 
[the `cmd` directory](https://github.com/ur/go-ur/tree/develop/cmd):

 Command  |         |
----------|---------|
`gur` | UR CLI (ur command line interface client) |
`bootnode` | runs a bootstrap node for the Discovery Protocol |
`urtest` | test tool which runs with the [tests](https://github.com/ur/tests) suite: `/path/to/test.json > urtest --test BlockTests --stdin`.
`evm` | is a generic UR Virtual Machine: `evm -code 60ff60ff -gas 10000 -price 0 -dump`. See `-h` for a detailed description. |
`disasm` | disassembles EVM code: `echo "6001" | disasm` |
`rlpdump` | prints RLP structures |

## Command line options

`gur` can be configured via command line options, environment variables and config files.

To get the options available:

    gur help

For further details on options, see the [wiki](https://github.com/ur/go-ur/wiki/Command-Line-Options)

## Contribution

If you'd like to contribute to go-ur please fork, fix, commit and
send a pull request. Commits who do not comply with the coding standards
are ignored (use gofmt!). If you send pull requests make absolute sure that you
commit on the `develop` branch and that you do not merge to master.
Commits that are directly based on master are simply ignored.

See [Developers' Guide](https://github.com/ur/go-ur/wiki/Developers'-Guide)
for more details on configuring your environment, testing, and
dependency management.
