# LSIF quickstart guide

## [Language-specific guides](languages/index.md)

We are working on creating guides for each language with an LSIF indexer, so make sure to check for the documentation for your language! If there is not a guide for your language, this general guide will help you through the LSIF setup process.

## Manual LSIF generation

We'll walk you through installing and generating LSIF data locally on your machine, and then manually uploading the LSIF data to your Sourcegraph instance for your repository. This will let you experiment with the process locally, and test your generated LSIF data on your repository before changing your CI process.

## 1. Set up your environment

1. Install the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) - used for uploading LSIF data to your Sourcegraph instance. This will work in a jiffy (replace `linux` with `darwin` for macOS):

  ```
  curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
  chmod +x /usr/local/bin/src
  ```

1. Install the LSIF indexer for your repository's language:
  1. Go to https://lsif.dev
  1. Find the LSIF indexer for your language
  1. Install the indexer as a command-line tool using the installation instructions in the indexer's README

### What is an LSIF indexer?

An LSIF indexer is a command line tool that analyzes your project's source code and generates a file in LSIF format containing all the definitions, references, and hover documentation in your project. That LSIF file is later uploaded to Sourcegraph to provide code intelligence.

## 2. Generate the LSIF data

You now need to generate the LSIF data for your repository. Each language's LSIF is unique to that language, so run the command in the _generate LSIF data_ step found in the README of the installed indexer.

## 3. Upload the LSIF data

For all languages, the upload step is the same. Make sure the current working directory is somewhere inside your repository, then use the Sourcegraph CLI to run:

```console
# for private instances
$ src -endpoint=<your sourcegraph endpoint> lsif upload -file=<LSIF file (e.g. dump.lsif)>
# to upload to Sourcegraph.com
$ src lsif upload -github-token=<your github token> -file=<LSIF file (e.g. dump.lsif)>
```

The upload command in the Sourcegraph CLI will try to infer the repository and git commit by invoking git commands on your local clone. If git is not installed, is older than version 2.7.0, or you are running on code outside of a git clone, you will need to also specify the `-repo` and `-commit` flags explicitly.

> NOTE: If you're using Sourcegraph.com or have enabled [`lsifEnforceAuth`](https://docs.sourcegraph.com/admin/config/site_config#lsifEnforceAuth), you need to [supply a GitHub token](#proving-ownership-of-a-github-repository) (to confirm you have collaborator access to the repository) supplied via the `-github-token` flag in the command above.

If successful, you'll see the following message:

```
Repository: <location of repository>
Commit: <40-character commit associated with this LSIF upload>
File: <LSIF data file>
Root: <subdirectory in the repository where this LSIF dump was generated>

LSIF dump successfully uploaded for processing.
View processing status at <link to your Sourcegraph instance LSIF status>.
```

Possible upload errors include:

- Clone in progress: the instance doesn't have the necessary data to process your upload yet, retry in a few minutes
- Unknown repository (404): check your `-endpoint` and make sure you can view the repository on your Sourcegraph instance
- Invalid commit (404): try visiting the repository at that commit on your Sourcegraph instance to trigger an update
- Invalid auth when using Sourcegraph.com or when [`lsifEnforceAuth`](https://docs.sourcegraph.com/admin/config/site_config#lsifEnforceAuth) is `true` (401 for an invalid token or 404 if the repository cannot be found on GitHub.com): make sure your GitHub token is valid and that the repository is correct
- Unexpected errors (500s): [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new)

LSIF processing failures for a repository are listed in **Repository settings > Code intelligence > Activity for this repository**. Failures can occur if the LSIF data is invalid (e.g., malformed indexer output), or problems were encountered during processing (e.g., system-level bug, flaky connections, etc). Try again or [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new) if the problem persists.

## 4. Test out code intelligence

Make sure you have [enabled LSIF code intelligence](lsif.md#enabling-lsif-on-your-sourcegraph-instance) on your Sourcegraph instance.

Once the LSIF data is uploaded, navigate to a code file for the targeted language in the repository (or sub-directory) the LSIF index was generated from. LSIF data should now be the source of hover-tooltips, definitions, and references for that file.

To verify that LSIF is correctly enabled, hover over a symbol and ensure that its hover text is not decorated with a ![tooltip](img/basic-code-intel-tooltip.svg) tooltip. This icon indicates the results are from search-based [basic code intelligence](./basic_code_intelligence.md) and should be absent when results are precise.

## 5. Productionize the process

Now that you're happy with the code intelligence on your repository, you need to make sure that is stays up to date with your repository. Our [continuous integration guide](adding_lsif_to_workflows.md) will get you started.
