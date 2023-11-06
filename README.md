# talos-operator

talosctl gen secrets
talosctl gen config --with-secrets secrets.yaml k-test https://10.171.120.200:6443

./build_images.sh http://10.171.120.1:8888 worker welcome123



kubectl patch machine 10.171.123.189 --type='json' -p='[{"op": "replace", "path": "/spec/bootstrap", "value":true}]'