# jvs-plugin-github

**This is not an official Google product.**

This repo contains component for github plugin for justification verification service (jvs).

## Prerequisite

This plugin uses github app for authentication. See [authetication with github app](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/about-authentication-with-a-github-app) for more detail.

Follow the directions from here [GitHub instructions](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app#creating-a-github-app) to create your github app, make sure to capture `app id` and `private key` during the creation.

After the app is created, install the app, and grant it with issue read permission to the repos you want to access, and capture the installation id.

## Installation

Please refer the this [example module](./terraform/example/main.tf) for setting up the infra.
