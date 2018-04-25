# nodereg

This is a simple controller that watches for annotations on a node, when that annotation equals a format like `rossedman.io/register` it will then make an HTTP request to the url contained in the annotation and mark it as registered.

This is a simple proof of concept that shows how kubernetes can be extended. This uses a watcher, lister, informer and cache to create.

Use case would be:
- Have multiple clusters and want to register nodes with central database
- When nodes register, want to interact with a queue system that triggers other events
- When nodes register, want to trigger some sort of webhook 

## Using

First build and start the controller

```
goreleaser --snapshot
./dist/darwin_amd64/nodereg --kubeconfig ~/.kube/config --anotation-prefix rossedman.io
```

Then add an annotation to your nodes that has a url where a `POST` request can be sent, for this demo I just used requestbin.

```sh
kubectl annotate nodes $nodename rossedman.io/register=http://requestbin.fullcontact.com/ybo4dcyb
```

Your controller will respond and send a request as well as add another annotation to your node, `rossedman.io/registered=true`.

To have the node be updated, just change this annotation:

```sh
kubectl annotate nodes $nodename rossedman.io/registered=false
```