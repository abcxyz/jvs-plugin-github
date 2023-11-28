# Release

**JVS-PLUGIN-GITHUB is not an official Google product.**

We leverage [goreleaser](https://goreleaser.com/) for SCM (GitHub) and image
release.

## New Release

- If a new JVS base image is needed, update [Dockerfile](https://github.com/abcxyz/jvs-plugin-github/blob/main/Dockerfile)
to point to a new JVS base image, and merge the PR.

- Send a PR to update all dependencies For Go
```sh
go get -u && go mod tidy
```

- Create a tag using `.github/workflows/create-tag.yml` on default branch and
     run the workflow with below inputs.

  - tag name with format `v0.x.x`, using semantic versioning.
        -   If there are breaking changes, bump the major version.
        -   If there are new major features (but not breaking), bump the minor
            version.
        -   Nothing important, bump the patch version.
        -   Feel free to use suffixes -alpha, -beta and -rc as needed.
  - skip to use default message (tag name).

- The new tag created should trigger the release workflow which typically does
    three things:

  - Integration test.
  - Container image release and push the images to container registry
        `us-docker.pkg.dev/abcxyz-artifacts/docker-images`.
  - GitHub release with artifacts (e.g. code zip, binaries, etc.).
        Note: Goreleaser will automatically use the git change log to fill the
        release note.

## Manually Release Images

Or if you want to build/push images for local development.

**1. Prepare JVS image**

Pre-built JVS release images can be found from `us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvsctl`. Please ensure to select the
image compatible with your github plugin.

If you want to build a JVS image for local development, follow [this guidance](https://github.com/abcxyz/jvs/blob/main/docs/release.md#manually-release-images).

**2. Overlay jvs-plugin-github plugin image**

Update [Dockerfile](https://github.com/abcxyz/jvs-plugin-github/blob/main/Dockerfile)
accordingly to point to the JVS image, then run the commands below to manually
build the image.

```sh
# Set the container registry for the images, for example:
export CONTAINER_REGISTRY=us-docker.pkg.dev/my-project/images

# goreleaser expects a "clean" repo to release so commit any local changes if
# needed.
git add . && git commit -m "local changes"

# goreleaser expects a tag.
# The tag must be a semantic version https://semver.org/
# DON'T push the tag if you're not releasing.
git tag -f -a v0.0.0-$(git rev-parse --short HEAD)

# goreleaser will tag the image with the git tag, optionally, override it by:
export DOCKER_TAG=mytag

# Use goreleaser to build the images.
# All the images will be tagged with the git tag given earlier.
# Use --skip-validate flag if you are trying to release from a older commit.
# To use a new JVS release, update the base image to a new version in
# .goreleaser.docker.yaml.
goreleaser release -f .goreleaser.docker.yaml --clean
```
