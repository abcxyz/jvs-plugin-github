# To use a new JVS release, update the base image to a new version.
FROM us-docker.pkg.dev/abcxyz-artifacts/docker-images/jvsctl:0.1.4

COPY jvs-plugin-github /var/jvs/plugins/jvs-plugin-github

ENTRYPOINT ["/bin/jvsctl"]
