# Developing cbuild2cmake

Follow steps below to start developing for `cbuild2cmake`:

1. Requirements:

- [Install Make](https://www.gnu.org/software/make/)
- [Install Golang](https://golang.org/doc/install)
- [Install GolangCI-Lint](https://golangci-lint.run/usage/install/#local-installation)

2. Clone the repo:
`$ git clone https://github.com/Open-CMSIS-Pack/cbuild2cmake.git`

3. Enter the checked source
`cd cbuild2cmake`

4. Configure your local environment
`make config`

5. Make sure all tests are passing
`make test-all`

6. Make sure it builds
`make build/cbuild2cmake`

7. Done! You can now start changing the source code, please refer to [contributing guide](CONTRIBUTING.md)
 to start contributing to the project

# Releasing

If you have rights to push to the `main` branch of this repo, you might be entitled to
make releases. Do that by running:
`make release`

*NOTE*: We use [Semantic Versioning](https://semver.org/) for versioning cbuild2cmake.
