_This package was developed by [Silicon Ally](https://siliconally.org) while
working on a project for [Adventure Scientists](https://adventurescientists.org).
Many thanks to Adventure Scientists for supporting [our open source
mission](https://siliconally.org/policies/open-source/)!_

# testsops

[![GoDoc](https://pkg.go.dev/badge/github.com/Silicon-Ally/testsops?status.svg)](https://pkg.go.dev/github.com/Silicon-Ally/testsops?tab=doc)
[![CI Workflow](https://github.com/Silicon-Ally/testsops/actions/workflows/test.yml/badge.svg)](https://github.com/Silicon-Ally/testsops/actions?query=branch%3Amain)

`testsops` is a simple package for testing libraries that use [Mozilla's
sops](https://github.com/mozilla/sops) tool for encryption. Specifically, this
package can be used for testing libraries that make use of sops' [decrypt
package](https://pkg.go.dev/go.mozilla.org/sops/v3@v3.7.3/decrypt). See
[testsops_test.go](/testsops_test.go) for an example.

## How does it work?

In normal usage, a sops-encrypted file uses a key stored in a cloud KMS system,
but that's impractical for local testing. However, `sops` also supports GPG +
[age](https://github.com/FiloSottile/age), the latter of which we use in
`testsops`. When a file is requested a be encrypted, a new `age` identity is
created and used to encrypt the file. To decrypt, the client points `sops` at
the `age` `key.txt` file via the `SOPS_AGE_KEY_FILE` env var.
