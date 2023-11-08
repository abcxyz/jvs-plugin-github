export JVSCTL_JWKS_ENDPOINT="https://ci-6163015863-419---jvs-public-key-f22e-2nhpyabgtq-uc.a.run.app/.well-known/jwks"
export $JVSCTL_SERVER_ADDRESS=jvs-api-fb61-xhlo4ajwba-uc.a.run.app:443

jvsctl token create -justification 'aaa' -server ci-6164015840-422---jvs-api-4527-2nhpyabgtq-uc.a.run.app:443 --auth-token $(gcloud auth print-identity-token)

token valite -jwks-endpoint https://jvs-public-key-4d18-xhlo4ajwba-uc.a.run.app/.well-known/jwks -token=""